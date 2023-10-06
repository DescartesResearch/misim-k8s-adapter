package control

import (
	"go-kube/pkg/misim"
	"go-kube/pkg/storage"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"strconv"
	"sync"
)

type PodController struct {
	storage     *storage.StorageContainer
	idGenerator IdGenerator
	mu          sync.Mutex
}

func (c *PodController) UpdatePods(ur v1.PodList, events []metav1.WatchEvent, podsToBePlaced v1.PodList, deleteEvents bool) misim.PodsUpdateResponse {
	if !deleteEvents {
		klog.V(3).Info("Pod-Update: ", len(ur.Items), " pods, ", len(podsToBePlaced.Items), " to be placed")
		// Buffers have to be cleared before storing new pods
		c.storage.Pods.PodsToBePlaced().Clear()
		c.storage.Pods.PodsToBePlaced().PutAll(podsToBePlaced.Items)
		c.storage.Pods.FailedPodBuffer().Clear()
		c.storage.Pods.BindedPodBuffer().Clear()
		c.storage.AdapterState.StoreClusterAutoscalingDone(false)
		c.storage.Nodes.NewNodes().Clear()
		c.storage.Nodes.DeletedNodes().Clear()

		// Store pods
		c.storage.Pods.StorePods(ur, events)

		// If there were pods to be placed, wait for the response
		if !c.storage.Pods.PodsToBePlaced().Empty() {
			podUpdateChannel := c.storage.Pods.PodsUpdateChannel().InitChannel()
			//c.storage.Pods.PodsUpdateChannel().InitChannel()
			klog.V(3).Infof("Wait for pods to be placed...")
			// wait for it
			select {
			case response := <-podUpdateChannel:
				return response
			}
		} else {
			return c.createDefaultResponse()
		}
	} else {
		klog.V(3).Info("Pod-Update: Deleted pods")
		c.storage.Pods.DeletePods(events)
	}
	return c.createDefaultResponse()
}

// Generates an update about all the pods that should be placed
func (c *PodController) createDefaultResponse() misim.PodsUpdateResponse {
	failedList := make([]misim.BindingFailureInformation, 0)
	bindedList := make([]misim.BindingInformation, 0)
	for _, pod := range c.storage.Pods.PodsToBePlaced().Items() {
		// Check if we received information for some pods from the scheduler
		podReported := false
		for _, bindFailure := range c.storage.Pods.FailedPodBuffer().Items() {
			if bindFailure.Pod == pod.Name {
				failedList = append(failedList, bindFailure)
				podReported = true
				break
			}
		}
		if podReported == true {
			continue
		}
		for _, bindSuccess := range c.storage.Pods.BindedPodBuffer().Items() {
			if bindSuccess.Pod == pod.Name {
				bindedList = append(bindedList, bindSuccess)
				podReported = true
				break
			}
		}
		if podReported == true {
			continue
		}
		failedList = append(failedList, misim.BindingFailureInformation{Pod: pod.Name, Message: "No new situation for the scheduler"})
	}
	if c.storage.Pods.FailedPodBuffer().Empty() && c.storage.Pods.BindedPodBuffer().Empty() && c.storage.AdapterState.IsClusterAutoscalerActive() && !c.storage.AdapterState.IsClusterAutoscalingDone() && c.storage.MachineSets.IsDownscalingPossible() {
		// TODO [Cluster Downscaling]: Integrate downscaling
		// (but maybe not here???)
	}
	return misim.PodsUpdateResponse{
		Binded:       bindedList,
		Failed:       failedList,
		NewNodes:     c.storage.Nodes.NewNodes().Items(),
		DeletedNodes: c.storage.Nodes.DeletedNodes().Items(),
	}

}

func (c *PodController) BindPod(podName string, nodeName string) {
	c.storage.Pods.BeginTransaction()

	klog.V(3).Info("Bound: " + podName + " to " + nodeName)
	// Get pod reference
	pod := c.storage.Pods.GetPod(podName)

	// Put binding information into buffer
	bindingInformation := misim.BindingInformation{Pod: podName, Node: nodeName}
	c.storage.Pods.BindedPodBuffer().Put(bindingInformation)

	// Update pod data and store it updated
	pod.ObjectMeta.ResourceVersion = c.idGenerator.GetNextResourceId()
	pod.Spec.NodeName = nodeName
	pod.Status.Phase = "Running"
	pod.Status.Conditions = append(pod.Status.Conditions, core.PodCondition{
		Type:   core.PodScheduled,
		Status: core.ConditionTrue,
	})

	c.storage.Pods.UpdatePod(podName, pod)
	c.updatePodChannel()

	c.storage.Pods.EndTransaction()
}

func (c *PodController) FailedPod(podName string, status core.PodStatus) {
	c.storage.Pods.BeginTransaction()

	klog.V(3).Info("Failed: " + podName)

	// Get pod reference
	pod := c.storage.Pods.GetPod(podName)

	// Put binding information in buffer
	failureInformation := misim.BindingFailureInformation{
		Pod:     podName,
		Message: status.Conditions[0].Message,
	}
	c.storage.Pods.FailedPodBuffer().Put(failureInformation)

	// Update pods data
	pod.Status = status
	pod.Status.Phase = "Pending"
	pod.ObjectMeta.ResourceVersion = c.idGenerator.GetNextResourceId()

	c.storage.Pods.UpdatePod(podName, pod)
	c.updatePodChannel()

	c.storage.Pods.EndTransaction()
}

func (c *PodController) updatePodChannel() {
	processedPodCount := c.storage.Pods.FailedPodBuffer().Size() + c.storage.Pods.BindedPodBuffer().Size()
	podsToBePlacedCount := c.storage.Pods.PodsToBePlaced().Size()
	klog.V(3).Info("Processessed " + strconv.Itoa(processedPodCount) + " of " + strconv.Itoa(podsToBePlacedCount))
	if processedPodCount == podsToBePlacedCount {
		if c.shouldScaleUp() {
			// @Martin Was?
			// Copied over from KubeUpdateController
			// Do we subscribe just to cancel it again???
			// Why?
			broadcaster := c.storage.Nodes.GetNodeUpscalingChannel()
			nodeChannel := broadcaster.Subscribe()
			defer broadcaster.CancelSubscription(nodeChannel)
			var newNode v1.Node
			klog.V(6).Info("Waiting for cluster-autoscaler upscaling")
			newNode = <-nodeChannel
			c.storage.Nodes.NewNodes().Put(newNode)
			// TODO [Process Status Config map from Cluster Autoscaler]: read from status config map of cluster autoscaler to track status
			c.storage.AdapterState.StoreClusterAutoscalingDone(true)
			// Empty failed pod buffer, they should be scheduled now
			c.storage.Pods.FailedPodBuffer().Clear()
		} else {
			c.storage.Pods.PodsUpdateChannel().Get() <- misim.PodsUpdateResponse{
				Failed:       c.storage.Pods.FailedPodBuffer().Items(),
				Binded:       c.storage.Pods.BindedPodBuffer().Items(),
				NewNodes:     c.storage.Nodes.NewNodes().Items(),
				DeletedNodes: c.storage.Nodes.DeletedNodes().Items(),
			}
		}
	}
}

func (c *PodController) shouldScaleUp() bool {
	return !c.storage.Pods.FailedPodBuffer().Empty() &&
		c.storage.AdapterState.IsClusterAutoscalerActive() &&
		!c.storage.AdapterState.IsClusterAutoscalingDone() &&
		c.storage.MachineSets.IsUpscalingPossible()
}

func NewPodController(storage *storage.StorageContainer) PodController {
	return PodController{
		storage: storage,
		idGenerator: IdGenerator{
			idStorage: storage.PodIds,
		},
	}
}
