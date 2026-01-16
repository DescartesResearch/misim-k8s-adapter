package control

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-kube/pkg/misim"
	"go-kube/pkg/storage"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

type NodeController struct {
	storage *storage.StorageContainer
}

func (c NodeController) UpdateNodes(nodes v1.NodeList, events []metav1.WatchEvent) misim.NodeUpdateResponse {
	klog.V(3).Info("Node-Update: ", len(nodes.Items), " nodes")
	c.storage.Nodes.StoreNodes(nodes, events)
	return misim.NodeUpdateResponse{
		NewNodes: []v1.Node{},
	}
}

func (c NodeController) InitMachinesNodes(nodes v1.NodeList, events []metav1.WatchEvent, machineSets []cluster.MachineSet, machines []cluster.Machine) misim.NodeUpdateResponse {
	klog.V(3).Infof("Machine-Node-Init: %d nodes, %d machine sets, %d machines", len(nodes.Items), len(machineSets), len(machines))

	newNodes := make([]v1.Node, 0)

	// Check whether machine sets etc. are already initialized, if yes, we need modified events, no additions
	oldMachineSetList, _ := c.storage.MachineSets.GetMachineSets()
	if oldMachineSetList.Items != nil {
		for _, event := range events {
			// Simulation can only modify nodes
			if event.Type == "MODIFIED" {
				var node v1.Node
				err := json.Unmarshal(event.Object.Raw, &node)
				if err != nil {
					klog.V(1).Infof("Cannot unmarshal provided node, raw data: %s", string(event.Object.Raw))
					return misim.NodeUpdateResponse{}
				}
				now := metav1.NewTime(time.Now())
				node.Status.Conditions[0].Reason = "SimulationNodeFailed"
				node.Status.Conditions[0].Message = "Node marked failed by MiSim"
				node.Status.Conditions[0].LastHeartbeatTime = now
				node.Status.Conditions[0].LastTransitionTime = now
				klog.V(5).Infof("Modified node %s with condition +%v", node.Name, node.Status.Conditions[0])
				c.storage.Nodes.PutNode(node.Name, node)
				machineList, _ := c.storage.Machines.GetMachines()
				for _, machine := range machineList.Items {
					if machine.Status.NodeRef.Name == node.Name {
						machineSetName := machine.OwnerReferences[0].Name

						// *machine.Spec.ProviderID = ""
						// machine.Status.NodeRef = nil
						machine.Status.Phase = "FAILED"

						klog.V(5).Infof("Modified machine %s to phase FAILED", machine.Name)
						c.storage.Machines.PutMachine(machine.Name, machine)

						// Get machine set by name
						set := c.storage.MachineSets.GetMachineSet(machineSetName)
						// Decrease replica counts

						(*set.Spec.Replicas)--
						set.Status.ReadyReplicas--
						set.Status.AvailableReplicas--
						set.Status.Replicas--
						set.Status.FullyLabeledReplicas--
						klog.V(5).Infof("Modified machine set %s to spec replica %d and status replica counts (%d, %d, %d, %d)", set.Name, *set.Spec.Replicas, set.Status.ReadyReplicas, set.Status.AvailableReplicas, set.Status.Replicas, set.Status.FullyLabeledReplicas)
						c.storage.MachineSets.PutMachineSet(machineSetName, set)

						// Check for minimum label
						if annotationValue, ok := set.Annotations["cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size"]; ok {
							minNodes, err := strconv.Atoi(annotationValue)
							if err != nil {
								break
							}
							klog.V(5).Infof("Minimum replica count for machine set %s is %d", set.Name, minNodes)

							// If minimum label is violated:
							if *set.Spec.Replicas < int32(minNodes) {
								// Update machine set
								(*set.Spec.Replicas)++
								set.Status.ReadyReplicas++
								set.Status.AvailableReplicas++
								set.Status.Replicas++
								set.Status.FullyLabeledReplicas++
								klog.V(5).Infof("Modified machine set %s to spec replica %d and status replica counts (%d, %d, %d, %d)", set.Name, *set.Spec.Replicas, set.Status.ReadyReplicas, set.Status.AvailableReplicas, set.Status.Replicas, set.Status.FullyLabeledReplicas)
								c.storage.MachineSets.PutMachineSet(machineSetName, set)

								// Add new machine
								nextMachineID := c.storage.Machines.GetMachineCount()
								machineName := fmt.Sprintf("%s-machine-%d", set.Name, nextMachineID)
								providerID := fmt.Sprintf("clusterapi://%s", machineName)
								nodeName := fmt.Sprintf("%s-node", machineName)
								nodeRef := v1.ObjectReference{Kind: "Node", APIVersion: "v1", Name: nodeName}

								newMachine := cluster.Machine{
									TypeMeta: metav1.TypeMeta{APIVersion: "cluster.x-k8s-io/v1beta1", Kind: "Machine"},
									ObjectMeta: metav1.ObjectMeta{
										Name: machineName, Namespace: "kube-system", Annotations: map[string]string{
											"machine-set-name": set.Name,
											"cpu":              set.Annotations["capacity.cluster-autoscaler.kubernetes.io/cpu"],
											"memory":           set.Annotations["capacity.cluster-autoscaler.kubernetes.io/memory"],
											"pods":             set.Annotations["capacity.cluster-autoscaler.kubernetes.io/maxPods"],
										}, OwnerReferences: []metav1.OwnerReference{
											{
												APIVersion: "cluster.x-k8s.io/v1beta1",
												Kind:       "MachineSet",
												Name:       set.Name,
											},
										},
									},
									Spec:   cluster.MachineSpec{ProviderID: &providerID},
									Status: cluster.MachineStatus{Phase: "Running", NodeRef: &nodeRef},
								}
								c.storage.Machines.IncrementMachineCount()
								klog.V(5).Infof("Adding new machine %s", newMachine.Name)
								c.storage.Machines.AddMachine(newMachine)
								// Add new node

								cpuQuantity, _ := resource.ParseQuantity(newMachine.Annotations["capacity.cluster-autoscaler.kubernetes.io/cpu"])
								klog.V(5).Infof("Parsed CPU quantity for node %s: %s", nodeName, &cpuQuantity)
								memoryQuantity, _ := resource.ParseQuantity(newMachine.Annotations["capacity.cluster-autoscaler.kubernetes.io/memory"])
								klog.V(5).Infof("Parsed Memory quantity for node %s: %s", nodeName, &memoryQuantity)
								podsQuantity, _ := resource.ParseQuantity(newMachine.Annotations["capacity.cluster-autoscaler.kubernetes.io/maxPods"])
								klog.V(5).Infof("Parsed Pods quantity for node %s: %s", nodeName, &podsQuantity)
								newNode := v1.Node{
									TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Node"},
									ObjectMeta: metav1.ObjectMeta{Name: nodeName, Labels: set.Spec.Template.ObjectMeta.Labels, Annotations: set.Spec.Template.ObjectMeta.Annotations},
									Spec:       v1.NodeSpec{ProviderID: providerID},
									Status: v1.NodeStatus{
										Phase: "Running", Conditions: []v1.NodeCondition{
											{
												Type:   "Ready",
												Status: "True",
											},
										},
										Allocatable: map[v1.ResourceName]resource.Quantity{
											"cpu":    cpuQuantity,
											"memory": memoryQuantity,
											"pods":   podsQuantity,
										},
										Capacity: map[v1.ResourceName]resource.Quantity{
											"cpu":    cpuQuantity,
											"memory": memoryQuantity,
											"pods":   podsQuantity,
										},
									},
								}

								klog.V(5).Infof("Adding new node %s", newNode.Name)
								c.storage.Nodes.AddNode(newNode)
								newNodes = append(newNodes, newNode)
							}

						}
						break
					}
				}
			}
		}
	} else {
		// Activate the cluster autoscaling!
		c.storage.AdapterState.StoreClusterAutoscalerActive(true)

		// first register machine sets
		machineSetList := cluster.MachineSetList{
			TypeMeta: metav1.TypeMeta{APIVersion: "cluster.x-k8s.io/v1beta1", Kind: "MachineSetList"},
			Items:    machineSets,
		}
		var machineSetsAddedEvents []metav1.WatchEvent
		// Each
		for i := range machineSets {
			temp := metav1.WatchEvent{Type: "ADDED", Object: runtime.RawExtension{Object: &machineSets[i]}}
			machineSetsAddedEvents = append(machineSetsAddedEvents, temp)
		}
		c.storage.MachineSets.StoreMachineSets(machineSetList, machineSetsAddedEvents)

		// second, store the machines
		machineList := cluster.MachineList{
			TypeMeta: metav1.TypeMeta{APIVersion: "cluster.x-k8s.io/v1beta1", Kind: "MachineList"},
			Items:    machines,
		}
		var machineAddedEvents []metav1.WatchEvent
		for i := range machines {
			temp := metav1.WatchEvent{Type: "ADDED", Object: runtime.RawExtension{Object: &machines[i]}}
			machineAddedEvents = append(machineAddedEvents, temp)
		}
		c.storage.Machines.StoreMachines(machineList, machineAddedEvents)

		// third, store the nodes
		c.storage.Nodes.StoreNodes(nodes, events)
	}

	return misim.NodeUpdateResponse{
		NewNodes: newNodes,
	}
}

func NewNodeController(storage *storage.StorageContainer) NodeController {
	return NodeController{
		storage: storage,
	}
}
