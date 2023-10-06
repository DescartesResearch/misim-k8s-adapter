package api

import (
	v1 "go-kube/pkg/interfaces/kubeapi/api/v1"
	"go-kube/pkg/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ApiResource interface {
	Get() metav1.APIVersions
	V1() v1.V1Resource
}

type ApiResourceImpl struct {
	storage *storage.StorageContainer
}

func (api ApiResourceImpl) Get() metav1.APIVersions {
	return metav1.APIVersions{
		Versions: []string{"v1"},
	}
}

func (api ApiResourceImpl) V1() v1.V1Resource {
	return v1.NewV1Resource(api.storage)
}

func NewApiResource(storage *storage.StorageContainer) ApiResourceImpl {
	return ApiResourceImpl{
		storage: storage,
	}
}
