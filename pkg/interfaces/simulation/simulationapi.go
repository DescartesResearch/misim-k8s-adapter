package simulation

import (
	"go-kube/pkg/control"
	"go-kube/pkg/storage"
)

type SimulationApi interface {
	NodeUpdates() control.NodeUpdatesResource
	PodUpdates() control.PodUpdatesResource
	Events() control.EventsResource
}

type SimulationApiImpl struct {
	storage *storage.StorageContainer
}

func (impl SimulationApiImpl) NodeUpdates() control.NodeUpdatesResource {
	return control.NewNodeUpdateResource(impl.storage)
}

func (impl SimulationApiImpl) PodUpdates() control.PodUpdatesResource {
	return control.NewPodUpdateResource(impl.storage)
}

func (impl SimulationApiImpl) Events() control.EventsResource {
	return control.NewEventsResource(impl.storage)
}

func NewSimulationApi(storage *storage.StorageContainer) SimulationApiImpl {
	return SimulationApiImpl{storage: storage}
}
