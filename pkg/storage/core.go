package storage

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube-rise/internal/broadcast"
	"net/http"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

type PodStorage interface {
	// Stores a podlist in the storage
	StorePods(pods v1.PodList, events []metav1.WatchEvent)
	// Retrieves the current podList from the storage
	GetPods() (v1.PodList, *broadcast.BroadcastServer[metav1.WatchEvent])
	// UpdatePodStatus(pod v1.Pod)
	DeletePods(events []metav1.WatchEvent)
	// Binds a pod to a node
	BindPod(podIndex int, nodeName string)
	// Reacts on scheduling fail
	FailedPod(podIndex int, status v1.PodStatus)
}

type NodeStorage interface {
	// Stores a nodelist in the storage
	StoreNodes(nodes v1.NodeList, events []metav1.WatchEvent)
	// Retrieves the current nodeList from the storage
	GetNodes() (v1.NodeList, *broadcast.BroadcastServer[metav1.WatchEvent])
	// Edits a node
	PutNode(w http.ResponseWriter, r *http.Request)
	// Gets a single node
	GetNode(w http.ResponseWriter, r *http.Request)
	// Scale nodes
	ScaleNodes(addedMachines []cluster.Machine, amount int) ([]v1.Node, error)
	// Channel for Node Upscaling
	GetNodeUpscalingChannel() *broadcast.BroadcastServer[v1.Node]
	// Channel for node downscaling
	GetNodeDownscalingChannel() *broadcast.BroadcastServer[v1.Node]
}

type NamespaceStorage interface {
	// Returns the current namespaces
	GetNamespaces() (v1.NamespaceList, *broadcast.BroadcastServer[metav1.WatchEvent])
	// Stores a namespace list in the storage
	StoreNamespaces(ns v1.NamespaceList)
}
