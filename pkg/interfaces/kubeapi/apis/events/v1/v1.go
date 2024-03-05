package v1

import (
	"go-kube/pkg/interfaces/kubeapi/apis/events/v1/namespaces"
	"go-kube/pkg/storage"
)

type V1Resource interface {
	Namespaces() namespaces.NamespacesResource
}

func (impl V1ResourceImpl) Namespaces() namespaces.NamespacesResource {
	return namespaces.NewNamespacesResource(impl.storage)
}

type V1ResourceImpl struct {
	storage *storage.StorageContainer
}

func NewV1Resource(storage *storage.StorageContainer) V1Resource {
	return V1ResourceImpl{storage: storage}
}
