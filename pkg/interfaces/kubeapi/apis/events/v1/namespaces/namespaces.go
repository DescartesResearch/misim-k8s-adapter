package namespaces

import (
	"go-kube/internal/broadcast"
	"go-kube/pkg/interfaces/kubeapi/apis/events/v1/namespaces/namespace"
	"go-kube/pkg/storage"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NamespacesResource interface {
	Get() (v1.NamespaceList, *broadcast.BroadcastServer[metav1.WatchEvent])
	Namespace(namespaceName string) namespace.NamespaceResource
}

type NamespacesResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl NamespacesResourceImpl) Get() (v1.NamespaceList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return impl.storage.Namespaces.GetNamespaces()
}

func (impl NamespacesResourceImpl) Namespace(namespaceName string) namespace.NamespaceResource {
	return namespace.NewNamespaceResource(namespaceName, impl.storage)
}

func NewNamespacesResource(storage *storage.StorageContainer) NamespacesResourceImpl {
	return NamespacesResourceImpl{
		storage: storage,
	}
}
