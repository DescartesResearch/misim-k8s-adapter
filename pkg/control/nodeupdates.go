package control

import (
	"go-kube/pkg/misim"
	"go-kube/pkg/storage"
)

type NodeUpdatesResource interface {
	Post(misim.NodeUpdateRequest) misim.NodeUpdateResponse
}

type NodeUpdatesResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl NodeUpdatesResourceImpl) Post(u misim.NodeUpdateRequest) misim.NodeUpdateResponse {
	controller := NewNodeController(impl.storage)

	if u.MachineSets == nil || len(u.MachineSets) == 0 {
		// No cluster scaling
		return controller.UpdateNodes(u.AllNodes, u.Events)
	} else {
		// If the request contains machines set, we use only the machines
		return controller.InitMachinesNodes(u.AllNodes, u.Events, u.MachineSets, u.Machines)
	}
}

func NewNodeUpdateResource(storage *storage.StorageContainer) NodeUpdatesResourceImpl {
	return NodeUpdatesResourceImpl{
		storage: storage,
	}
}
