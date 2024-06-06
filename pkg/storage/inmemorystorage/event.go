package inmemorystorage

import (
	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type EventInMemoryStorage struct {
	eventsApiEvents []eventsv1.Event
	coreApiEvents   []v1.Event
	//eventChan        chan metav1.WatchEvent
	//eventBroadcaster *broadcast.BroadcastServer[metav1.WatchEvent]
}

func (e *EventInMemoryStorage) StoreEventsApiEvent(event eventsv1.Event) eventsv1.Event {
	e.eventsApiEvents = append(e.eventsApiEvents, event)
	klog.V(7).Infof("EventInMemoryStorage.StoreEvent: %v", event)
	return event
}

func (e *EventInMemoryStorage) GetEventsApiEvents() eventsv1.EventList {
	// create EventList from events
	typeMeta := metav1.TypeMeta{
		Kind:       "EventList",
		APIVersion: "events.k8s.io/v1",
	}

	eventList := eventsv1.EventList{}
	eventList.TypeMeta = typeMeta
	eventList.Items = e.eventsApiEvents
	return eventList

}

func (e *EventInMemoryStorage) StoreCoreApiEvent(event v1.Event) v1.Event {
	e.coreApiEvents = append(e.coreApiEvents, event)
	klog.V(7).Infof("EventInMemoryStorage.StoreEvent: %v", event)
	return event
}

func (e *EventInMemoryStorage) GetCoreApiEvents() v1.EventList {
	// create EventList from events
	typeMeta := metav1.TypeMeta{
		Kind:       "EventList",
		APIVersion: "v1",
	}

	eventList := v1.EventList{}
	eventList.TypeMeta = typeMeta
	eventList.Items = e.coreApiEvents
	return eventList
}

func NewEventInMemoryStorage() EventInMemoryStorage {
	//eventChan := make(chan metav1.WatchEvent)
	return EventInMemoryStorage{
		eventsApiEvents: []eventsv1.Event{},
		coreApiEvents:   []v1.Event{},
	}
}
