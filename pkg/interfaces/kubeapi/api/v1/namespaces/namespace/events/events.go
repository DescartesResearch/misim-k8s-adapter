package events

import (
	"go-kube/pkg/storage"

	v1 "k8s.io/api/core/v1"
)

// Core API events
// /api/v1/namespaces/{namespace}/events

type EventsResource interface {
	Post(event v1.Event) v1.Event
}

type EventsResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl EventsResourceImpl) Post(event v1.
	Event) v1.Event {
	impl.storage.Events.StoreCoreApiEvent(event)
	return event
}

func NewEventsResource(storage *storage.StorageContainer) EventsResourceImpl {
	return EventsResourceImpl{storage}
}
