package storage

import (
	"go-kube/internal/broadcast"
	"go-kube/pkg/misim"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodStorage interface {
	BeginTransaction()
	EndTransaction()
	// Stores a podlist in the storage
	StorePods(pods v1.PodList, events []metav1.WatchEvent)
	// Retrieves the current podList from the storage
	GetPods() (v1.PodList, *broadcast.BroadcastServer[metav1.WatchEvent])
	// UpdatePodStatus(pod v1.Pod)
	DeletePods(events []metav1.WatchEvent)
	// Get Pod by name
	GetPod(podName string) v1.Pod
	// Updates the pod with the passed name
	// and triggers watch event
	UpdatePod(podName string, newValues v1.Pod)

	// Buffer for failed pods
	FailedPodBuffer() Buffer[misim.BindingFailureInformation]
	// Buffer for binded pods
	BindedPodBuffer() Buffer[misim.BindingInformation]
	// Buffer for pods that should be placed
	PodsToBePlaced() Buffer[v1.Pod]

	// Channel to return the pod update request when all pods are placed (or failed)
	PodsUpdateChannel() ChannelWrapper[misim.PodsUpdateResponse]
}
