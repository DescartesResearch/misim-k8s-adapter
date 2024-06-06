package control

import (
	"go-kube/pkg/storage"

	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
)

type EventsResource interface {
	GetEventsApiEvents() eventsv1.EventList
	GetCoreApiEvents() v1.EventList
}

type EventsResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl EventsResourceImpl) GetEventsApiEvents() eventsv1.EventList {
	controller := NewEventsController(impl.storage)
	return controller.GetEventsApiEvents()
}

func (impl EventsResourceImpl) GetCoreApiEvents() v1.EventList {
	controller := NewEventsController(impl.storage)
	return controller.GetCoreApiEvents()
}

func NewEventsResource(storage *storage.StorageContainer) EventsResourceImpl {
	return EventsResourceImpl{
		storage: storage,
	}
}
