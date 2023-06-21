package kubeapi

import (
	"kube-rise/pkg/server/infrastructure"
	"kube-rise/pkg/storage"
)

type AppsKubeAPIServer struct {
	GetDaemonSets infrastructure.Endpoint
}

func NewAppsKubeAPIServer(storageContainer *storage.StorageContainer) *AppsKubeAPIServer {
	var s = &AppsKubeAPIServer{}
	s.GetDaemonSets = infrastructure.HandleWatchableRequest(storageContainer.DaemonSets.GetDaemonSets)
	return s
}
