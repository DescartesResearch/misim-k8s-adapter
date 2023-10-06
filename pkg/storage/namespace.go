package storage

import (
	"go-kube/internal/broadcast"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NamespaceStorage interface {
	// Returns the current namespaces
	GetNamespaces() (v1.NamespaceList, *broadcast.BroadcastServer[metav1.WatchEvent])
	// Stores a namespace list in the storage
	StoreNamespaces(ns v1.NamespaceList)
	// Returns a single namespace by name
	GetNamespace(name string) v1.Namespace
}
