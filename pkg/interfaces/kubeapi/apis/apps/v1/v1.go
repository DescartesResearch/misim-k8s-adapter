package v1

import (
	"go-kube/pkg/interfaces/kubeapi/apis/apps/v1/daemonsets"
	"go-kube/pkg/storage"
)

type V1Resource interface {
	DaemonSets() daemonsets.DaemonSetsResource
}

type V1ResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl V1ResourceImpl) DaemonSets() daemonsets.DaemonSetsResource {
	return daemonsets.NewDeamonSetsResource(impl.storage)
}

func NewV1Resource(storage *storage.StorageContainer) V1ResourceImpl {
	return V1ResourceImpl{
		storage: storage,
	}
}
