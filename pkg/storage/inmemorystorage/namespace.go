package inmemorystorage

import (
	"context"
	"go-kube/internal/broadcast"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NamespaceInMemoryStorage struct {
	namespaces           core.NamespaceList
	namespaceEventChan   chan metav1.WatchEvent
	namespaceBroadcaster *broadcast.BroadcastServer[metav1.WatchEvent]
}

func (s *NamespaceInMemoryStorage) GetNamespaces() (core.NamespaceList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return s.namespaces, s.namespaceBroadcaster
}

func (s *NamespaceInMemoryStorage) StoreNamespaces(namespaces core.NamespaceList) {
	s.namespaces = namespaces
}

func (s *NamespaceInMemoryStorage) GetNamespace(namespaceName string) core.Namespace {
	var u core.Namespace
	for _, element := range s.namespaces.Items {
		if namespaceName == element.Name {
			u = element
			break
		}
	}
	return u
}

func NewNamespaceInMemoryStorage() NamespaceInMemoryStorage {
	var namespace core.Namespace
	namespace.SetName("default")
	namespace.Status = core.NamespaceStatus{Phase: "Active"}
	namespaceEventChan := make(chan metav1.WatchEvent)
	return NamespaceInMemoryStorage{
		namespaces:           core.NamespaceList{TypeMeta: metav1.TypeMeta{Kind: "NamespaceList", APIVersion: "v1"}, Items: []core.Namespace{namespace}},
		namespaceEventChan:   namespaceEventChan,
		namespaceBroadcaster: broadcast.NewBroadcastServer(context.TODO(), "NamespaceBroadcaster", namespaceEventChan),
	}
}
