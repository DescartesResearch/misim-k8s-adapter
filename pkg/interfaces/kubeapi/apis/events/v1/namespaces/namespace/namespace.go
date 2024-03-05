package namespace

import (
	"go-kube/pkg/interfaces/kubeapi/apis/events/v1/namespaces/namespace/events"
	"go-kube/pkg/storage"
)

type NamespaceResource interface {
	Events() events.EventsResource
}

type NamespaceResourceImpl struct {
	namespaceName string
	storage       *storage.StorageContainer
}

func (impl NamespaceResourceImpl) Events() events.EventsResource {
	return events.NewEventsResource(impl.storage)
}

func NewNamespaceResource(name string, storage *storage.StorageContainer) NamespaceResource {
	return NamespaceResourceImpl{
		namespaceName: name,
		storage:       storage,
	}
}
