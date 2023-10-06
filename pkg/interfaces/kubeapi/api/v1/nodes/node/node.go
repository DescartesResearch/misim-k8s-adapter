package node

import (
	"go-kube/pkg/storage"
	v1 "k8s.io/api/core/v1"
)

type NodeResource interface {
	Get() v1.Node
	Put(v1.Node) v1.Node
}

type NodeResourceImpl struct {
	name         string
	nodesStorage storage.NodeStorage
}

func (impl NodeResourceImpl) Get() v1.Node {
	nodes := impl.nodesStorage.GetNode(impl.name)
	return nodes
}

func (impl NodeResourceImpl) Put(node v1.Node) v1.Node {
	return impl.nodesStorage.PutNode(impl.name, node)
}

func NewNodeResource(nodeName string, nodeStorage storage.NodeStorage) NodeResourceImpl {
	return NodeResourceImpl{
		name:         nodeName,
		nodesStorage: nodeStorage,
	}
}
