package events

import (
	v1 "go-kube/pkg/interfaces/kubeapi/apis/events/v1"
	"go-kube/pkg/storage"
)

type EventsResource interface {
	V1() v1.V1Resource
}

type EventsResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl EventsResourceImpl) V1() v1.V1Resource {
	return v1.NewV1Resource(impl.storage)
}

func NewEventsResource(storage *storage.StorageContainer) EventsResource {
	return EventsResourceImpl{
		storage: storage,
	}
}
