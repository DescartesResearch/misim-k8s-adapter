package events

import (
	"go-kube/pkg/storage"

	eventsv1 "k8s.io/api/events/v1"
)

// /apis/events.k8s.io/v1/namespaces/{namespace}/events

// TODO: GET /apis/events.k8s.io/v1/namespaces/{namespace}/events/{name}
// TODO: PUT /apis/events.k8s.io/v1/namespaces/{namespace}/events/{name}
// TODO: DELETE /apis/events.k8s.io/v1/namespaces/{namespace}/events/{name}
type EventsResource interface {
	Post(event eventsv1.Event) eventsv1.Event
}

type EventsResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl EventsResourceImpl) Post(event eventsv1.
	Event) eventsv1.
	Event {
	impl.storage.Events.StoreEvent(event)
	return event
}

func NewEventsResource(storage *storage.StorageContainer) EventsResourceImpl {
	return EventsResourceImpl{storage}
}
