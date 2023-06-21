package storage

import (
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube-rise/internal/broadcast"
)

type DaemonSetStorage interface {
	// Stores a daemonset list in the storage
	StoreDaemonSets(ds v1.DaemonSetList, events []metav1.WatchEvent)
	// Returns the current daemonsets
	GetDaemonSets() (v1.DaemonSetList, *broadcast.BroadcastServer[metav1.WatchEvent])
}
