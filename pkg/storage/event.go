package storage

import (
	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
)

type EventStorage interface {
	StoreEventsApiEvent(event eventsv1.Event) eventsv1.Event
	GetEventsApiEvents() eventsv1.EventList
	StoreCoreApiEvent(event v1.Event) v1.Event
	GetCoreApiEvents() v1.EventList
}
