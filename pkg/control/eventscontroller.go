package control

import (
	"go-kube/pkg/storage"

	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/klog/v2"
)

type EventsController struct {
	storage *storage.StorageContainer
}

func (c EventsController) GetEventsApiEvents() eventsv1.EventList {
	eventList := c.storage.Events.GetEventsApiEvents()
	klog.V(3).Info("Returning events API event list with ", len(eventList.Items), " items")
	return eventList
}

func (c EventsController) GetCoreApiEvents() v1.EventList {
	eventList := c.storage.Events.GetCoreApiEvents()
	klog.V(3).Info("Returning core API event list with ", len(eventList.Items), " items")
	return eventList
}

func NewEventsController(storage *storage.StorageContainer) EventsController {
	return EventsController{
		storage: storage,
	}
}
