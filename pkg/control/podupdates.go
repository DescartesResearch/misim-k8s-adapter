package control

import (
	"go-kube/pkg/misim"
	"go-kube/pkg/storage"
)

type PodUpdatesResource interface {
	Post(misim.PodsUpdateRequest) misim.PodsUpdateResponse
}

type PodUpdatesResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl PodUpdatesResourceImpl) Post(u misim.PodsUpdateRequest) misim.PodsUpdateResponse {
	controller := NewPodController(impl.storage)
	return controller.UpdatePods(u.AllPods, u.Events, u.PodsToBePlaced, false)
}

func NewPodUpdateResource(storage *storage.StorageContainer) PodUpdatesResourceImpl {
	return PodUpdatesResourceImpl{
		storage: storage,
	}
}
