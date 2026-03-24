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

// NodeMonitorGracePeriodSeconds defines the duration (in seconds) that the
// Kubernetes node controller waits without receiving a heartbeat from a node
// before marking it as NotReady.
//
// This corresponds to the kube-controller-manager flag --node-monitor-grace-period.
// See the official documentation for details:
// https://kubernetes.io/docs/reference/command-line-tools-reference/kube-controller-manager/
const NodeMonitorGracePeriodSeconds = 50

// NodeMonitorPeriodSeconds defines the period for syncing NodeStatus updates in the node controller.
//
// This corresponds to the kube-controller-manager flag --node-monitor-period.
// See the official documentation for details:
// https://kubernetes.io/docs/reference/command-line-tools-reference/kube-controller-manager/
const NodeMonitorPeriodSeconds = 5

// DefaultNotReadyTolerationSeconds defines the duration (in seconds)
// of the toleration for notReady:NoExecute that is added by default
// to every pod that does not already have such a toleration.
const DefaultNotReadyTolerationSeconds = 300

var (
	// UnreachableTaintTemplate is the taint for when a node becomes unreachable.
	UnreachableTaintTemplate = &v1.Taint{
		Key:    v1.TaintNodeUnreachable,
		Effect: v1.TaintEffectNoExecute,
	}

	// NotReadyTaintTemplate is the taint for when a node is not ready for
	// executing pods
	NotReadyTaintTemplate = &v1.Taint{
		Key:    v1.TaintNodeNotReady,
		Effect: v1.TaintEffectNoExecute,
	}
)

// NodeController handles all node-related requests sent by the simulation.
type NodeController struct {
	storage *storage.StorageContainer
}

// UpdateNodes handles node updates when there is no cluster autoscaler.
func (c *NodeController) UpdateNodes(nodes v1.NodeList, events []metav1.WatchEvent) {
	klog.V(3).Info("Node-Update: ", len(nodes.Items), " nodes")
	c.storage.Nodes.StoreNodes(nodes, events)
}

// InitMachinesNodes handles node updates when machines are present
// (i.e., the cluster autoscaler is connected).
func (c *NodeController) InitMachinesNodes(nodes v1.NodeList, events []metav1.WatchEvent, machineSets []cluster.MachineSet, machines []cluster.Machine) {
	klog.V(3).Infof("Machine-Node-Init: %d nodes, %d machine sets, %d machines", len(nodes.Items), len(machineSets), len(machines))

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
					return
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

								klog.V(5).Infof("Machine Set CPU: %s", set.Annotations["capacity.cluster-autoscaler.kubernetes.io/cpu"])
								klog.V(5).Infof("Machine Set Memory: %s", set.Annotations["capacity.cluster-autoscaler.kubernetes.io/memory"])
								klog.V(5).Infof("Machine Set Pods: %s", set.Annotations["capacity.cluster-autoscaler.kubernetes.io/maxPods"])

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

								cpuQuantity, _ := resource.ParseQuantity(newMachine.Annotations["cpu"])
								klog.V(5).Infof("Parsed CPU quantity for node %s: %s", nodeName, &cpuQuantity)
								memoryQuantity, _ := resource.ParseQuantity(newMachine.Annotations["memory"])
								klog.V(5).Infof("Parsed Memory quantity for node %s: %s", nodeName, &memoryQuantity)
								podsQuantity, _ := resource.ParseQuantity(newMachine.Annotations["pods"])
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
								c.storage.Nodes.NewNodeUpdateBuffer().Put(newNode)
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
}

// HandleNodeFailure processes a list of pods running on failed nodes and determines
// when each pod should be evicted based on its tolerations.
//
// For each pod in `failedPods`, the function checks the pod's tolerations for the
// "NotReady" taint with effect `NoExecute`:
//   - If the pod has a toleration with `TolerationSeconds == nil`, it is never evicted
//   - If the toleration is zero or negative, the pod is evicted immediately (0 seconds).
//   - If the toleration is positive, the pod is scheduled for eviction after that many seconds.
//
// Pods without a matching toleration are scheduled using the default eviction delay
// defined by `DefaultNotReadyTolerationSeconds`.
func (c *NodeController) HandleNodeFailure(failedPods []string) misim.NodeFailureResponse {
	evictionEvents := make(map[int64][]string, len(failedPods))
NEXT:
	for _, name := range failedPods {
		pod := c.storage.Pods.GetPod(name)
		for _, toleration := range pod.Spec.Tolerations {
			if toleration.Effect == v1.TaintEffectNoExecute && toleration.Key == v1.TaintNodeUnreachable {
				tolerationSeconds := toleration.TolerationSeconds
				switch {
				case tolerationSeconds == nil:
					// The pod should never be evicted
					// Thus, we do not include it in the response
					continue NEXT
				case *tolerationSeconds <= int64(0):
					// Treat negative and zero values as immediate eviction
					// (as per the official toleration docs)
					evictionEvents[0] = append(evictionEvents[0], name)
					continue NEXT
				case *tolerationSeconds > int64(0):
					evictionEvents[*tolerationSeconds] = append(evictionEvents[*tolerationSeconds], name)
					continue NEXT
				}
			}
		}
		evictionEvents[DefaultNotReadyTolerationSeconds] = append(evictionEvents[DefaultNotReadyTolerationSeconds], name)
	}

	return misim.NodeFailureResponse{
		NodeMonitorGracePeriodSeconds: NodeMonitorGracePeriodSeconds,
		NoExecuteTaintDelaySeconds:    NodeMonitorPeriodSeconds,
		PodEvictionEvents:             evictionEvents,
	}
}

// HandleMarkNodeNotReady marks all given nodes as "NotReady" and returns their updated kubernetes
// representations.
func (c *NodeController) HandleMarkNodeNotReady(nodes []string) misim.NodeNotReadyResponse {
	updatedNodes := make([]v1.Node, 0, len(nodes))
	updatedMachines := make([]cluster.Machine, 0, len(nodes))
	updatedMachineSets := make([]cluster.MachineSet, 0, len(nodes))
	for _, name := range nodes {
		node := c.storage.Nodes.GetNode(name)

		// Node Condition
		foundNotReady := false
		for i, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady {
				foundNotReady = true
				condition.Status = v1.ConditionUnknown
				condition.Reason = "SimulationNodeFailed"
				node.Status.Conditions[i] = condition
				break
			}
		}
		if !foundNotReady {
			condition := v1.NodeCondition{Type: v1.NodeReady, Status: v1.ConditionUnknown, Reason: "SimulationNodeFailed"}
			node.Status.Conditions = append(node.Status.Conditions, condition)
		}

		// Taints
		found := false
		for i, taint := range node.Spec.Taints {
			if taint.Key == v1.TaintNodeUnschedulable {
				found = true
				taint.Effect = v1.TaintEffectNoSchedule
				node.Spec.Taints[i] = taint
				break
			}
		}
		if !found {
			taint := v1.Taint{Key: v1.TaintNodeUnschedulable, Effect: v1.TaintEffectNoSchedule}
			node.Spec.Taints = append(node.Spec.Taints, taint)
		}
		updatedNodes = append(updatedNodes, node)

		machines, _ := c.storage.Machines.GetMachines()
		for _, machine := range machines.Items {
			if machine.Status.NodeRef.Name == node.Name {
				for _, owner := range machine.OwnerReferences {
					machineSet := c.storage.MachineSets.GetMachineSet(owner.Name)
					machineSet.Status.ReadyReplicas--
					machineSet.Status.AvailableReplicas--
					updatedMachineSets = append(updatedMachineSets, machineSet)
					c.storage.MachineSets.PutMachineSet(owner.Name, machineSet)
				}
				break
			}
		}
		c.storage.Nodes.PutNode(node.Name, node)

	}

	return misim.NodeNotReadyResponse{
		Nodes:       updatedNodes,
		Machines:    updatedMachines,
		MachineSets: updatedMachineSets,
	}
}

// HandleMarkNodeNoExecute adds a NoExecute taint to all requested nodes and returns their updated
// Kubernetes representations.
func (c *NodeController) HandleMarkNodeNoExecute(nodes []string) misim.NodeNoExecuteResponse {
	updatedNodes := make([]v1.Node, 0, len(nodes))
	for _, name := range nodes {
		node := c.storage.Nodes.GetNode(name)
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady {
				taintToAdd := v1.Taint{}
				oppositeTaint := v1.Taint{}
				// Because we want to mimic NodeStatus.Condition["Ready"] we make "unreachable" and "not ready" taints mutually exclusive.
				// See: https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/nodelifecycle/node_lifecycle_controller.go
				switch condition.Status {
				case v1.ConditionFalse:
					taintToAdd = *NotReadyTaintTemplate
					oppositeTaint = *UnreachableTaintTemplate
				case v1.ConditionUnknown:
					taintToAdd = *UnreachableTaintTemplate
					oppositeTaint = *NotReadyTaintTemplate
				default:
					// It seems that the Node is ready again, so there's no need to taint it.
					klog.V(4).Info("Node %s was in a taint queue, but it's ready now. Ignoring taint request", name)
				}

				now := metav1.Now()
				taintToAdd.TimeAdded = &now
				for i := range node.Spec.Taints {
					taint := node.Spec.Taints[i]
					if taint.Key == oppositeTaint.Key && taint.Effect == oppositeTaint.Effect {
						l := len(node.Spec.Taints)
						node.Spec.Taints[i] = node.Spec.Taints[l-1]
						node.Spec.Taints = node.Spec.Taints[:l-1]
						break
					}
				}
				node.Spec.Taints = append(node.Spec.Taints, taintToAdd)
				c.storage.Nodes.PutNode(node.Name, node)
			}
		}
	}

	return misim.NodeNoExecuteResponse{
		Nodes: updatedNodes,
	}
}

// NewNodeController creates a new node controller.
func NewNodeController(storage *storage.StorageContainer) NodeController {
	return NodeController{
		storage: storage,
	}
}
