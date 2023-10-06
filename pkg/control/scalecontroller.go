package control

import (
	"errors"
	"fmt"
	"go-kube/pkg/storage"
	autoscaling "k8s.io/api/autoscaling/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

type ScaleController struct {
	storage *storage.StorageContainer
}

func (c ScaleController) ScaleMachineSet(machineSetName string, u autoscaling.Scale) error {
	desiredReplicas := u.Spec.Replicas
	scaleAmount := int32(0)
	machineSet := c.storage.MachineSets.GetMachineSet(machineSetName)
	scaleAmount = desiredReplicas - *machineSet.Spec.Replicas
	klog.V(3).Infof("Scaling machine set %s to desired amount %d (scale amount %d)", machineSetName, desiredReplicas, scaleAmount)

	// Update machineset
	machineSet.Spec.Replicas = &desiredReplicas
	machineSet.Status.AvailableReplicas = desiredReplicas
	machineSet.Status.FullyLabeledReplicas = desiredReplicas
	machineSet.Status.ReadyReplicas = desiredReplicas

	c.storage.MachineSets.PutMachineSet(machineSetName, machineSet)

	// Check if we have to scale down
	if scaleAmount < 0 {
		// For downscaling, we first delete nodes then machines
		scaledDownNodes, err := c.ScaleDownNodes(-int(scaleAmount))

		if err != nil {
			return err
		}

		// Scale machines
		c.ScaleDownMachines(machineSet, scaledDownNodes, -int(scaleAmount))
	} else if scaleAmount > 0 {
		addedMachines, err := c.ScaleUpMachines(machineSet, int(scaleAmount))

		if err != nil {
			return err
		}

		// Scale nodes
		c.ScaleUpNodes(addedMachines)
	}

	return nil
}

// TODO [Cluster Downscaling]: Fix this method
func (c ScaleController) ScaleDownMachines(machineSet cluster.MachineSet, changedNodes []core.Node, amount int) {
	// In case of downscaling we need to delete machines
	for _, changedNode := range changedNodes {
		allMachines, _ := c.storage.Machines.GetMachines()
		var nodeMachine cluster.Machine
		for _, machine := range allMachines.Items {
			if machine.Status.NodeRef.Name == changedNode.Name {
				nodeMachine = machine
				break
			}
		}

		c.storage.Machines.DeleteMachine(nodeMachine.Name)
	}
}

func (c ScaleController) ScaleUpMachines(machineSet cluster.MachineSet, amount int) ([]cluster.Machine, error) {
	var addedMachines []cluster.Machine
	providerIds := make([]string, amount)
	nodeRefs := make([]core.ObjectReference, amount)
	for amount > 0 {
		nextMachineId := c.storage.Machines.GetMachineCount()
		providerIds[amount-1] = "clusterapi://" + fmt.Sprintf("%s-machine-%d", machineSet.Name, nextMachineId)
		nodeRefs[amount-1] = core.ObjectReference{Kind: "Node", APIVersion: "v1", Name: fmt.Sprintf("%s-machine-%d", machineSet.Name, nextMachineId) + "-node"}

		newMachine := cluster.Machine{
			TypeMeta: metav1.TypeMeta{APIVersion: "cluster.x-k8s.io/v1beta1", Kind: "Machine"},
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-machine-%d", machineSet.Name, nextMachineId), Namespace: "kube-system", Annotations: map[string]string{
				"machine-set-name": machineSet.Name,
				"cpu":              machineSet.Annotations["capacity.cluster-autoscaler.kubernetes.io/cpu"],
				"memory":           machineSet.Annotations["capacity.cluster-autoscaler.kubernetes.io/memory"],
				"pods":             machineSet.Annotations["capacity.cluster-autoscaler.kubernetes.io/maxPods"],
			}, OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "cluster.x-k8s.io/v1beta1",
					Kind:       "MachineSet",
					Name:       machineSet.Name,
				},
			}},
			Spec:   cluster.MachineSpec{ProviderID: &providerIds[amount-1]},
			Status: cluster.MachineStatus{Phase: "Running", NodeRef: &nodeRefs[amount-1]},
		}
		klog.V(5).Infof("Prepared new machine %s", newMachine.Name)
		c.storage.Machines.IncrementMachineCount()
		amount = amount - 1
		addedMachines = append(addedMachines, newMachine)
	}
	for _, machine := range addedMachines {
		klog.V(5).Infof("Adding machine %s", machine.Name)
		c.storage.Machines.AddMachine(machine)
	}
	return addedMachines, nil
}

func (c ScaleController) ScaleDownNodes(amount int) ([]core.Node, error) {
	// Find nodes that should be deleted to scale down
	var nodesToDelete []core.Node
	allNodes, _ := c.storage.Nodes.GetNodes()
	for _, node := range allNodes.Items {
		for _, taint := range node.Spec.Taints {
			if taint.Key == "ToBeDeletedByClusterAutoscaler" {
				nodesToDelete = append(nodesToDelete, node)
				// Break taint iteration (not node iteration)
				break
			}
		}
	}

	// Check if the right amount of nodes is marked for deletion
	if len(nodesToDelete) != amount {
		return nil, errors.New(fmt.Sprintf("Mismatch: found %d desired nodes to delete, got %d tainted nodes", amount, len(nodesToDelete)))
	}

	var changedNodes []core.Node
	for amount > 0 {
		nodeToDelete := nodesToDelete[amount-1]

		changedNodes = append(changedNodes, nodeToDelete)
		c.storage.Nodes.DeleteNode(nodeToDelete.Name)

		amount = amount - 1
	}
	// TODO [Cluster Downscaling]: Check whether this works (especially pointer to changedNodes should not be overwritten)

	return changedNodes, nil
}

func (c ScaleController) ScaleUpNodes(addedMachines []cluster.Machine) {
	newNodes := make([]core.Node, len(addedMachines))
	for i, machine := range addedMachines {
		cpuQuantity, _ := resource.ParseQuantity(machine.Annotations["cpu"])
		memoryQuantity, _ := resource.ParseQuantity(machine.Annotations["memory"])
		podsQuantity, _ := resource.ParseQuantity(machine.Annotations["pods"])
		newNodes[i] = core.Node{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Node"},
			ObjectMeta: metav1.ObjectMeta{Name: machine.Name + "-node"},
			Spec:       core.NodeSpec{ProviderID: "clusterapi://" + machine.Name},
			Status: core.NodeStatus{Phase: "Running", Conditions: []core.NodeCondition{
				{
					Type:   "Ready",
					Status: "True",
				},
			}, Allocatable: map[core.ResourceName]resource.Quantity{
				"cpu":    cpuQuantity,
				"memory": memoryQuantity,
				"pods":   podsQuantity,
			}, Capacity: map[core.ResourceName]resource.Quantity{
				"cpu":    cpuQuantity,
				"memory": memoryQuantity,
				"pods":   podsQuantity,
			}},
		}

	}
	for i := range newNodes {
		c.storage.Nodes.AddNode(newNodes[i])
	}
}

func NewScaleController(storage *storage.StorageContainer) ScaleController {
	return ScaleController{storage: storage}
}
