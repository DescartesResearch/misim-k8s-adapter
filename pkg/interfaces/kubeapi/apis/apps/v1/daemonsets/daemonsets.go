package daemonsets

import (
	"go-kube/internal/broadcast"
	"go-kube/pkg/storage"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DaemonSetsResource interface {
	Get() (v1.DaemonSetList, *broadcast.BroadcastServer[metav1.WatchEvent])
}

type DaemonSetsResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl DaemonSetsResourceImpl) Get() (v1.DaemonSetList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return impl.storage.DaemonSets.GetDaemonSets()
}

func NewDeamonSetsResource(storage *storage.StorageContainer) DaemonSetsResourceImpl {
	return DaemonSetsResourceImpl{storage: storage}
}
