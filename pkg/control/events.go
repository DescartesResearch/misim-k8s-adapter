package control

import (
	"go-kube/pkg/storage"

	eventsv1 "k8s.io/api/events/v1"
)

type EventsResource interface {
	Get() eventsv1.EventList
}

type EventsResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl EventsResourceImpl) Get() eventsv1.EventList {
	controller := NewEventsController(impl.storage)
	return controller.GetEvents()
}

func NewEventsResource(storage *storage.StorageContainer) EventsResourceImpl {
	return EventsResourceImpl{
		storage: storage,
	}
}
