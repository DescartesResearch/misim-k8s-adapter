package control

import (
	"encoding/json"
	"time"

	"go-kube/pkg/misim"
	"go-kube/pkg/storage"

	v1 "k8s.io/api/core/v1"
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
		Data: nodes,
	}
}

func (c NodeController) InitMachinesNodes(nodes v1.NodeList, events []metav1.WatchEvent, machineSets []cluster.MachineSet, machines []cluster.Machine) misim.NodeUpdateResponse {
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
					return misim.NodeUpdateResponse{}
				}
				now := metav1.NewTime(time.Now())
				node.Status.Conditions[0].Reason = "SimulationNodeFailed"
				node.Status.Conditions[0].Message = "Node marked failed by MiSim"
				node.Status.Conditions[0].LastHeartbeatTime = now
				node.Status.Conditions[0].LastTransitionTime = now
				klog.V(5).Infof("Modified node %s with condition +%v", node.Name, node.Status.Conditions[0])
				c.storage.Nodes.PutNode(node.Name, node)
				// c.storage.Nodes.DeleteNode(node.Name)
				machineList, _ := c.storage.Machines.GetMachines()
				machineSetName := ""
				for _, machine := range machineList.Items {
					if machine.Status.NodeRef.Name == node.Name {
						machineSetName = machine.OwnerReferences[0].Name

						*machine.Spec.ProviderID = ""
						machine.Status.NodeRef = nil
						machine.Status.Phase = "FAILED"

						// *machine.Status.FailureReason = "SimulationNodeFailed"
						// *machine.Status.FailureMessage = "Machine marked failed by MiSim"
						klog.V(5).Infof("Modified machine %s to phase FAILED", machine.Name)
						c.storage.Machines.PutMachine(machine.Name, machine)
						// c.storage.Machines.DeleteMachine(machine.Name)
						break
					}
				}
				if machineSetName != "" {
					set := c.storage.MachineSets.GetMachineSet(machineSetName)
					(*set.Spec.Replicas)--
					set.Status.ReadyReplicas--
					set.Status.AvailableReplicas--
					set.Status.Replicas--
					set.Status.FullyLabeledReplicas--
					klog.V(5).Infof("Modified machine set %s to spec replica %d and status replica counts (%d, %d, %d, %d)", set.Name, *set.Spec.Replicas, set.Status.ReadyReplicas, set.Status.AvailableReplicas, set.Status.Replicas, set.Status.FullyLabeledReplicas)
					c.storage.MachineSets.PutMachineSet(machineSetName, set)
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
		Data: nodes,
	}
}

func NewNodeController(storage *storage.StorageContainer) NodeController {
	return NodeController{
		storage: storage,
	}
}
