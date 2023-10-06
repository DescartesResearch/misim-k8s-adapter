package nodes

import (
	"go-kube/internal/broadcast"
	"go-kube/pkg/interfaces/kubeapi/api/v1/nodes/node"
	"go-kube/pkg/storage"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodesResource interface {
	Get() (v1.NodeList, *broadcast.BroadcastServer[metav1.WatchEvent])
	Node(nodeName string) node.NodeResource
}

type NodesResourceImpl struct {
	storage  *storage.StorageContainer
	nodeImpl node.NodeResource
}

func (impl NodesResourceImpl) Node(nodeName string) node.NodeResource {
	return node.NewNodeResource(nodeName, impl.storage.Nodes)
}

func (impl NodesResourceImpl) Get() (v1.NodeList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return impl.storage.Nodes.GetNodes()
}

func NewNodesResource(storage *storage.StorageContainer) NodesResourceImpl {
	return NodesResourceImpl{
		storage: storage,
	}
}
