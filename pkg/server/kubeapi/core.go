package kubeapi

import (
	"kube-rise/pkg/server/infrastructure"
	"kube-rise/pkg/storage"
)

type CoreKubeAPIServer struct {
	GetPods       infrastructure.Endpoint
	GetNodes      infrastructure.Endpoint
	GetNamespaces infrastructure.Endpoint
	PutNode       infrastructure.Endpoint
	GetNode       infrastructure.Endpoint
}

func NewCoreKubeAPIServer(storageContainer *storage.StorageContainer) *CoreKubeAPIServer {
	var s = &CoreKubeAPIServer{}
	s.GetPods = infrastructure.HandleWatchableRequest(storageContainer.Pods.GetPods)
	s.GetNodes = infrastructure.HandleWatchableRequest(storageContainer.Nodes.GetNodes)
	s.GetNamespaces = infrastructure.HandleWatchableRequest(storageContainer.Namespaces.GetNamespaces)
	s.PutNode = storageContainer.Nodes.PutNode
	s.GetNode = storageContainer.Nodes.GetNode
	return s
}
