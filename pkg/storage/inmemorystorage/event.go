package inmemorystorage

import (
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type EventInMemoryStorage struct {
	events []eventsv1.Event
	//eventChan        chan metav1.WatchEvent
	//eventBroadcaster *broadcast.BroadcastServer[metav1.WatchEvent]
}

func (e *EventInMemoryStorage) StoreEvent(event eventsv1.Event) eventsv1.Event {
	e.events = append(e.events, event)
	klog.V(7).Infof("EventInMemoryStorage.StoreEvent: %v", event)
	return event
}

func (e *EventInMemoryStorage) GetEvents() eventsv1.EventList {
	// create EventList from events
	typeMeta := metav1.TypeMeta{
		Kind:       "EventList",
		APIVersion: "events.k8s.io/v1",
	}

	eventList := eventsv1.EventList{}
	eventList.TypeMeta = typeMeta
	eventList.Items = e.events
	return eventList

}

func NewEventInMemoryStorage() EventInMemoryStorage {
	//eventChan := make(chan metav1.WatchEvent)
	return EventInMemoryStorage{
		events: []eventsv1.Event{},
	}
}
