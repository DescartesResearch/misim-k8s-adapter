package control

import (
	"go-kube/pkg/misim"
	"go-kube/pkg/storage"
)

type PodUpdatesResource interface {
	Post(misim.PodsUpdateRequest) misim.PodsUpdateResponse
	FailPod(misim.PodFailureRequest)
}

type PodUpdatesResourceImpl struct {
	controller *PodController
	storage    *storage.StorageContainer
}

func (impl PodUpdatesResourceImpl) Post(u misim.PodsUpdateRequest) misim.PodsUpdateResponse {
	return impl.controller.UpdatePods(u.AllPods, u.Events, u.PodsToBePlaced, false)
}

func (impl PodUpdatesResourceImpl) FailPod(u misim.PodFailureRequest) {
	impl.controller.failPod(u.FailedPod)
}

func NewPodUpdateResource(storage *storage.StorageContainer) PodUpdatesResourceImpl {
	controller := NewPodController(storage)
	return PodUpdatesResourceImpl{
		storage:    storage,
		controller: &controller,
	}
}
