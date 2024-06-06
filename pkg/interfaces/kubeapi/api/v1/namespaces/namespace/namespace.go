package namespace

import (
	"go-kube/pkg/interfaces/kubeapi/api/v1/namespaces/namespace/configmaps"
	"go-kube/pkg/interfaces/kubeapi/api/v1/namespaces/namespace/events"
	"go-kube/pkg/interfaces/kubeapi/api/v1/pods"
	"go-kube/pkg/storage"

	v1 "k8s.io/api/core/v1"
)

type NamespaceResource interface {
	Get() v1.Namespace
	Pods() pods.PodsResource
	Configmaps() configmaps.ConfigmapsResource
	Events() events.EventsResource
}

type NamespaceResourceImpl struct {
	namespaceName string
	storage       *storage.StorageContainer
}

func (impl NamespaceResourceImpl) Get() v1.Namespace {
	return impl.storage.Namespaces.GetNamespace(impl.namespaceName)
}

func (impl NamespaceResourceImpl) Pods() pods.PodsResource {
	// TODO [Support for multiple namespaces]: pass namespace name somehow if we want to support multiple namespaces
	return pods.NewPodsResource(impl.storage)
}

func (impl NamespaceResourceImpl) Configmaps() configmaps.ConfigmapsResource {
	return configmaps.NewConfigmapsResource(impl.namespaceName, impl.storage)
}

func (impl NamespaceResourceImpl) Events() events.EventsResource {
	return events.NewEventsResource(impl.storage)
}

func NewNamespaceResource(name string, storage *storage.StorageContainer) NamespaceResourceImpl {
	return NamespaceResourceImpl{
		namespaceName: name,
		storage:       storage,
	}
}
