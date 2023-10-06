package binding

import (
	"go-kube/pkg/control"
	"go-kube/pkg/storage"
	v1 "k8s.io/api/core/v1"
)

type BindingResource interface {
	Post(v1.Binding)
}

type BindingResourceImpl struct {
	podName string
	storage *storage.StorageContainer
}

func (impl BindingResourceImpl) Post(binding v1.Binding) {
	controller := control.NewPodController(impl.storage)
	controller.BindPod(impl.podName, binding.Target.Name)
}

func NewBindingResource(podName string, storage *storage.StorageContainer) BindingResourceImpl {
	return BindingResourceImpl{
		podName: podName,
		storage: storage,
	}
}
