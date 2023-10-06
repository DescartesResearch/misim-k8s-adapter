package storage

import (
	"go-kube/internal/broadcast"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeStorage interface {
	// Stores a nodelist in the storage
	StoreNodes(nodes v1.NodeList, events []metav1.WatchEvent)
	// Retrieves the current nodeList from the storage
	GetNodes() (v1.NodeList, *broadcast.BroadcastServer[metav1.WatchEvent])
	// Edits a node
	PutNode(name string, node v1.Node) v1.Node
	// Gets a single node
	GetNode(name string) v1.Node
	// Deletes a node from the node list
	DeleteNode(name string) v1.Node
	// Adds a node
	AddNode(v1.Node)
	// Channel for Node Upscaling
	GetNodeUpscalingChannel() *broadcast.BroadcastServer[v1.Node]
	// Channel for node downscaling
	GetNodeDownscalingChannel() *broadcast.BroadcastServer[v1.Node]
	// New nodes that should be created because of updates
	NewNodes() Buffer[v1.Node]
	// Nodes that should be deleted
	DeletedNodes() Buffer[v1.Node]
}
