package v1

import (
	"go-kube/pkg/storage"
)

// /apis/events.k8s.io/v1/events

type EventsResource interface {
	Namespace(namespaceName string)
}

type EventsResourceImpl struct {
	storage *storage.StorageContainer
}

func NewEventsResource(storage *storage.StorageContainer) EventsResourceImpl {
	return EventsResourceImpl{
		storage: storage,
	}
}
