package control

import (
	"go-kube/pkg/storage"

	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/klog/v2"
)

type EventsController struct {
	storage *storage.StorageContainer
}

func (c EventsController) GetEvents() eventsv1.EventList {
	eventList := c.storage.Events.GetEvents()
	klog.V(3).Info("Returning events list with ", len(eventList.Items), " items")
	return eventList
}

func NewEventsController(storage *storage.StorageContainer) EventsController {
	return EventsController{
		storage: storage,
	}
}
