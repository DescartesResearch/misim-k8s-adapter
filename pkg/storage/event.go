package storage

import eventsv1 "k8s.io/api/events/v1"

type EventStorage interface {
	StoreEvent(event eventsv1.Event) eventsv1.Event
	GetEvents() eventsv1.EventList
}
