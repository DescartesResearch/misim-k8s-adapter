package pod

import (
	"go-kube/pkg/interfaces/kubeapi/api/v1/pods/pod/binding"
	"go-kube/pkg/interfaces/kubeapi/api/v1/pods/pod/status"
	"go-kube/pkg/storage"
)

type PodResource interface {
	Status() status.StatusResource
	Binding() binding.BindingResource
}

type PodResourceImpl struct {
	podName string
	storage *storage.StorageContainer
}

func (impl PodResourceImpl) Status() status.StatusResource {
	return status.NewStatusResource(impl.podName, impl.storage)
}

func (impl PodResourceImpl) Binding() binding.BindingResource {
	return binding.NewBindingResource(impl.podName, impl.storage)
}

func NewPodResource(podName string, storage *storage.StorageContainer) PodResourceImpl {
	return PodResourceImpl{
		podName: podName,
		storage: storage,
	}
}
