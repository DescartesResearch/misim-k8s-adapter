package inmemorystorage

import (
	"context"
	"go-kube/internal/broadcast"
	"go-kube/pkg/storage"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type NodeInMemoryStorage struct {
	nodes                      core.NodeList
	nodeEventChan              chan metav1.WatchEvent
	nodeBroadcaster            *broadcast.BroadcastServer[metav1.WatchEvent]
	nodeUpscalingChan          chan core.Node
	nodeDownscalingChan        chan core.Node
	nodeUpscalingBroadcaster   *broadcast.BroadcastServer[core.Node]
	nodeDownscalingBroadcaster *broadcast.BroadcastServer[core.Node]
	newNodes                   InMemBuffer[core.Node]
	deletedNodes               InMemBuffer[core.Node]
}

func (s *NodeInMemoryStorage) GetNodes() (core.NodeList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return s.nodes, s.nodeBroadcaster
}

func (s *NodeInMemoryStorage) StoreNodes(nodes core.NodeList, events []metav1.WatchEvent) {
	s.nodes = nodes
	for _, n := range events {
		s.nodeEventChan <- n
	}
}

func (s *NodeInMemoryStorage) PutNode(nodeName string, node core.Node) core.Node {
	indexForReplacement := -1
	for index, node := range s.nodes.Items {
		if node.Name == nodeName {
			indexForReplacement = index
			break
		}
	}
	s.nodes.Items[indexForReplacement] = node
	return node
}

func (s *NodeInMemoryStorage) GetNode(name string) core.Node {
	var u core.Node

	for _, node := range s.nodes.Items {
		if node.Name == name {
			u = node
			break
		}
	}

	return u
}

func (s *NodeInMemoryStorage) AddNode(node core.Node) {
	s.nodes.Items = append(s.nodes.Items, node)
	// Fire added event
	nodeAddEvent := metav1.WatchEvent{Type: "ADDED", Object: runtime.RawExtension{Object: &node}}
	s.nodeEventChan <- nodeAddEvent
	s.nodeUpscalingChan <- node
}

func (s *NodeInMemoryStorage) DeleteNode(nodeName string) core.Node {
	// Get index
	var index int
	var deletedNode core.Node
	for i, node := range s.nodes.Items {
		if node.Name == nodeName {
			index = i
			deletedNode = node
			break
		}
	}
	s.nodes.Items[index] = s.nodes.Items[len(s.nodes.Items)-1]
	s.nodes.Items = s.nodes.Items[:len(s.nodes.Items)-1]
	// Fire event
	nodeDeleteEvent := metav1.WatchEvent{Type: "DELETED", Object: runtime.RawExtension{Object: &deletedNode}}
	s.nodeEventChan <- nodeDeleteEvent
	s.nodeDownscalingChan <- deletedNode
	return deletedNode
}

func (s *NodeInMemoryStorage) GetNodeUpscalingChannel() *broadcast.BroadcastServer[core.Node] {
	return s.nodeUpscalingBroadcaster
}

func (s *NodeInMemoryStorage) GetNodeDownscalingChannel() *broadcast.BroadcastServer[core.Node] {
	return s.nodeDownscalingBroadcaster
}

func (s *NodeInMemoryStorage) NewNodes() storage.Buffer[core.Node] {
	return &s.newNodes
}

func (s *NodeInMemoryStorage) DeletedNodes() storage.Buffer[core.Node] {
	return &s.deletedNodes
}

func NewNodeInMemoryStorage() NodeInMemoryStorage {
	nodeEventChan := make(chan metav1.WatchEvent, 500)
	nodeUpscalingChan := make(chan core.Node)
	nodeDownscalingChan := make(chan core.Node)
	return NodeInMemoryStorage{
		nodes:                      core.NodeList{TypeMeta: metav1.TypeMeta{Kind: "NodeList", APIVersion: "v1"}, Items: nil},
		nodeEventChan:              nodeEventChan,
		nodeBroadcaster:            broadcast.NewBroadcastServer(context.TODO(), "NodeBroadcaster", nodeEventChan),
		nodeUpscalingChan:          nodeUpscalingChan,
		nodeDownscalingChan:        nodeDownscalingChan,
		nodeDownscalingBroadcaster: broadcast.NewBroadcastServer(context.TODO(), "NodeDownscalingBroadcaster", nodeDownscalingChan),
		nodeUpscalingBroadcaster:   broadcast.NewBroadcastServer(context.TODO(), "NodeUpscalingBroadcaster", nodeUpscalingChan),

		newNodes:     NewInMemBuffer[core.Node](),
		deletedNodes: NewInMemBuffer[core.Node](),
	}
}
