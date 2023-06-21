package kubeapi

import (
	"kube-rise/pkg/mocks"
	"kube-rise/pkg/server/infrastructure"
	"kube-rise/pkg/storage"
)

type ClusterKubeAPIServer struct {
	GetMachineSets      infrastructure.Endpoint
	GetMachines         infrastructure.Endpoint
	GetClusters         infrastructure.Endpoint
	GetStatusConfigMap  infrastructure.Endpoint
	PutStatusConfigMap  infrastructure.Endpoint
	GetMachineSetsScale infrastructure.Endpoint
	PutMachineSetsScale infrastructure.Endpoint
	GetMachine          infrastructure.Endpoint
	PutMachine          infrastructure.Endpoint
}

func NewClusterKubeAPIServer(storageContainer *storage.StorageContainer) *ClusterKubeAPIServer {
	var server = &ClusterKubeAPIServer{}
	server.GetMachines = infrastructure.HandleWatchableRequest(storageContainer.Machines.GetMachines)
	server.GetMachineSets = infrastructure.HandleWatchableRequest(storageContainer.MachineSets.GetMachineSets)
	server.GetClusters = infrastructure.HandleRequest(mocks.GetClusters)
	server.GetStatusConfigMap = infrastructure.HandleRequest(storageContainer.StatusConfigMap.GetStatusConfigMap)
	server.PutStatusConfigMap = storageContainer.StatusConfigMap.StoreStatusConfigMap
	server.GetMachineSetsScale = storageContainer.MachineSets.GetMachineSetsScale
	server.PutMachineSetsScale = storageContainer.MachineSets.PutMachineSetsScale
	server.GetMachine = storageContainer.Machines.GetMachine
	server.PutMachine = storageContainer.Machines.PutMachine
	return server
}
