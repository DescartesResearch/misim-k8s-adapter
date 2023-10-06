package apps

import (
	v1 "go-kube/pkg/interfaces/kubeapi/apis/apps/v1"
	"go-kube/pkg/storage"
)

type AppsResource interface {
	V1() v1.V1Resource
}

type AppsResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl AppsResourceImpl) V1() v1.V1Resource {
	return v1.NewV1Resource(impl.storage)
}

func NewAppsResource(storage *storage.StorageContainer) AppsResourceImpl {
	return AppsResourceImpl{
		storage: storage,
	}
}
