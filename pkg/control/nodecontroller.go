package control

import (
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

	return misim.NodeUpdateResponse{
		Data: nodes,
	}
}

func NewNodeController(storage *storage.StorageContainer) NodeController {
	return NodeController{
		storage: storage,
	}
}
