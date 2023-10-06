package inmemorystorage

import (
	"context"
	"go-kube/internal/broadcast"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DaemonSetInMemoryStorage struct {
	daemonSets           apps.DaemonSetList
	daemonSetEventChan   chan metav1.WatchEvent
	daemonSetBroadcaster *broadcast.BroadcastServer[metav1.WatchEvent]
}

// DaemonSetStorage interface

func (d DaemonSetInMemoryStorage) StoreDaemonSets(ds apps.DaemonSetList, events []metav1.WatchEvent) {
	d.daemonSets = ds
	for _, event := range events {
		d.daemonSetEventChan <- event
	}
}

func (d DaemonSetInMemoryStorage) GetDaemonSets() (apps.DaemonSetList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return d.daemonSets, d.daemonSetBroadcaster
}

// Constructors

func NewDaemonSetInMemoryStorage() DaemonSetInMemoryStorage {
	daemonSetEventChan := make(chan metav1.WatchEvent)
	return DaemonSetInMemoryStorage{
		daemonSets:           apps.DaemonSetList{TypeMeta: metav1.TypeMeta{Kind: "DaemonSetList", APIVersion: "apps/v1"}, Items: nil},
		daemonSetEventChan:   daemonSetEventChan,
		daemonSetBroadcaster: broadcast.NewBroadcastServer(context.TODO(), "DaemonSetBroadcaster", daemonSetEventChan),
	}
}
