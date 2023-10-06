package kubeapi

import (
	"go-kube/pkg/interfaces/kubeapi/api"
	"go-kube/pkg/interfaces/kubeapi/apis"
	"go-kube/pkg/storage"
)

type KubeApi interface {
	Api() api.ApiResource
	Apis() apis.ApisResource
}

type KubeApiImpl struct {
	storage *storage.StorageContainer
}

func (impl KubeApiImpl) Api() api.ApiResource {
	return api.NewApiResource(impl.storage)
}

func (impl KubeApiImpl) Apis() apis.ApisResource {
	return apis.NewApisResource(impl.storage)
}

func NewKubeApi(storage *storage.StorageContainer) KubeApiImpl {
	return KubeApiImpl{
		storage: storage,
	}
}
