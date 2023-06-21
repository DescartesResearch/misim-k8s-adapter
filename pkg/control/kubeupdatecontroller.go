package control

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kube-rise/pkg/entity"
	"kube-rise/pkg/storage"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

// Controls the specific adapter logic for different Kubernetes API
type KubeUpdateController interface {
	// Updates the nodes and returns a channel for the response
	// The update might be executed async and returns the response via the channel
	UpdateNodes(ur v1.NodeList, events []metav1.WatchEvent) entity.NodeUpdateResponse
	// Initializes nodes and machines when cluster scaling is activated
	InitMachinesNodes(ur v1.NodeList, events []metav1.WatchEvent, machineSets []cluster.MachineSet, machines []cluster.Machine) entity.NodeUpdateResponse
	// Updates the pods and returns a channel for the response
	// The update might be executed async and returns the response via the channel
	UpdatePods(ur v1.PodList, events []metav1.WatchEvent, podsToBePlaced v1.PodList, deleteEvents bool) chan entity.PodsUpdateResponse
	// Fallback function in case scheduler is stuck
	CreateDefaultResponse() entity.PodsUpdateResponse
	// Called when a pod is failed
	Failed(status v1.PodStatus, podName string) v1.Pod
	// Called when a pod is binded
	Binded(binding v1.Binding, podName string)
}

type KubeUpdateControllerImpl struct {
	storageContainer *storage.StorageContainer

	nodesUpdateChannel chan entity.NodeUpdateResponse
	podsUpdateChannel  chan entity.PodsUpdateResponse

	failedPodBuffer         []entity.BindingFailureInformation
	bindedPodBuffer         []entity.BindingInformation
	podsToBePlaced          v1.PodList
	clusterAutoscalerActive bool
	clusterAutoscalerDone   bool
	newNodes                []v1.Node
	deletedNodes            []v1.Node
}

func NewKubeUpdateController(storageContainer *storage.StorageContainer) KubeUpdateController {
	return &KubeUpdateControllerImpl{storageContainer: storageContainer,
		failedPodBuffer:         []entity.BindingFailureInformation{},
		bindedPodBuffer:         []entity.BindingInformation{},
		podsToBePlaced:          v1.PodList{Items: []v1.Pod{}},
		clusterAutoscalerActive: false,
		clusterAutoscalerDone:   false,
		newNodes:                []v1.Node{},
		deletedNodes:            []v1.Node{},
	}
}

func (k *KubeUpdateControllerImpl) UpdateNodes(ur v1.NodeList, events []metav1.WatchEvent) entity.NodeUpdateResponse {
	k.storageContainer.Nodes.StoreNodes(ur, events)
	return entity.NodeUpdateResponse{Data: ur}
}

func (k *KubeUpdateControllerImpl) InitMachinesNodes(ur v1.NodeList, events []metav1.WatchEvent, machineSets []cluster.MachineSet, machines []cluster.Machine) entity.NodeUpdateResponse {
	k.clusterAutoscalerActive = true

	// first register machine sets
	machineSetList := cluster.MachineSetList{
		TypeMeta: metav1.TypeMeta{APIVersion: "cluster.x-k8s.io/v1beta1", Kind: "MachineSetList"},
		Items:    machineSets,
	}
	var machineSetsAddedEvents []metav1.WatchEvent
	for i := range machineSets {
		temp := metav1.WatchEvent{Type: "ADDED", Object: runtime.RawExtension{Object: &machineSets[i]}}
		machineSetsAddedEvents = append(machineSetsAddedEvents, temp)
	}
	k.storageContainer.MachineSets.StoreMachineSets(machineSetList, machineSetsAddedEvents)

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
	k.storageContainer.Machines.StoreMachines(machineList, machineAddedEvents)

	// third, store the nodes
	k.storageContainer.Nodes.StoreNodes(ur, events)

	return entity.NodeUpdateResponse{Data: ur}
}

func (k *KubeUpdateControllerImpl) UpdatePods(ur v1.PodList, events []metav1.WatchEvent, podsToBePlaced v1.PodList, deleteEvents bool) chan entity.PodsUpdateResponse {
	if deleteEvents == false {
		// We need to reset all buffers
		k.podsToBePlaced = podsToBePlaced
		k.failedPodBuffer = make([]entity.BindingFailureInformation, 0)
		k.bindedPodBuffer = make([]entity.BindingInformation, 0)
		k.clusterAutoscalerDone = false
		k.newNodes = []v1.Node{}
		k.deletedNodes = []v1.Node{}
		k.storageContainer.Pods.StorePods(ur, events)
		if len(k.podsToBePlaced.Items) > 0 {
			k.podsUpdateChannel = make(chan entity.PodsUpdateResponse)
		} else {
			k.podsUpdateChannel = nil
		}
		return k.podsUpdateChannel
	} else {
		k.storageContainer.Pods.DeletePods(events)
		return nil
	}
}

func (k *KubeUpdateControllerImpl) CreateDefaultResponse() entity.PodsUpdateResponse {
	failedList := make([]entity.BindingFailureInformation, 0)
	bindedList := make([]entity.BindingInformation, 0)
	for _, pod := range k.podsToBePlaced.Items {
		// Check if we received information for some pods from the scheduler
		podReported := false
		for _, bindFailure := range k.failedPodBuffer {
			if bindFailure.Pod == pod.Name {
				failedList = append(failedList, bindFailure)
				podReported = true
				break
			}
		}
		if podReported == true {
			continue
		}
		for _, bindSuccess := range k.bindedPodBuffer {
			if bindSuccess.Pod == pod.Name {
				bindedList = append(bindedList, bindSuccess)
				podReported = true
				break
			}
		}
		if podReported == true {
			continue
		}
		failedList = append(failedList, entity.BindingFailureInformation{Pod: pod.Name, Message: "No new situation for the scheduler"})
	}
	if len(failedList)+len(bindedList) == 0 && k.clusterAutoscalerActive && !k.clusterAutoscalerDone && k.storageContainer.MachineSets.IsDownscalingPossible() {
		// TODO: Integrate downscaling
	}
	return entity.PodsUpdateResponse{Binded: bindedList, Failed: failedList, NewNodes: k.newNodes, DeletedNodes: k.deletedNodes}
}

func (k *KubeUpdateControllerImpl) updatePodChannel() {
	if len(k.failedPodBuffer)+len(k.bindedPodBuffer) == len(k.podsToBePlaced.Items) {
		if len(k.failedPodBuffer) > 0 && k.clusterAutoscalerActive && !k.clusterAutoscalerDone && k.storageContainer.MachineSets.IsUpscalingPossible() {
			broadcaster := k.storageContainer.Nodes.GetNodeUpscalingChannel()
			nodeChannel := broadcaster.Subscribe()
			defer broadcaster.CancelSubscription(nodeChannel)
			var newNode v1.Node
			fmt.Println("Waiting for cluster-autoscaler upscaling")
			newNode = <-nodeChannel
			k.newNodes = append(k.newNodes, newNode)
			// TODO: Check if cluster autoscaler could react two times
			// TODO: read from status config map of cluster autoscaler to track status
			k.clusterAutoscalerDone = true
			// Empty failed pod buffer, they should be scheduled now
			k.failedPodBuffer = make([]entity.BindingFailureInformation, 0)
		} else {
			k.podsUpdateChannel <- entity.PodsUpdateResponse{Failed: k.failedPodBuffer, Binded: k.bindedPodBuffer, NewNodes: k.newNodes, DeletedNodes: k.deletedNodes}
		}
	}
}

func (k *KubeUpdateControllerImpl) Failed(status v1.PodStatus, podName string) v1.Pod {
	allPods, _ := k.storageContainer.Pods.GetPods()
	var result v1.Pod
	for i, element := range allPods.Items {
		if element.Name == podName {
			fmt.Printf("Pod %s cannot be scheduled, reason: %s\n", podName, status.Conditions[0].Message)
			k.failedPodBuffer = append(k.failedPodBuffer, entity.BindingFailureInformation{Pod: podName, Message: status.Conditions[0].Message})
			result = element
			k.storageContainer.Pods.FailedPod(i, status)
			break
		}
	}
	k.updatePodChannel()
	result.Status = status
	return result
}

func (k *KubeUpdateControllerImpl) Binded(binding v1.Binding, podName string) {
	allPods, _ := k.storageContainer.Pods.GetPods()
	for i, element := range allPods.Items {
		if element.Name == podName {
			fmt.Printf("Pod %s will be bound to Node %s\n", podName, binding.Target.Name)
			k.storageContainer.Pods.BindPod(i, binding.Target.Name)
			k.bindedPodBuffer = append(k.bindedPodBuffer, entity.BindingInformation{Pod: podName, Node: binding.Target.Name})
			// k.storageContainer.PodStorage.UpdatePodStatus(element)
			break
		}
	}
	k.updatePodChannel()
}
