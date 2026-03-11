package control

import (
	"go-kube/pkg/misim"
	"go-kube/pkg/storage"
)

// NodeUpdatesResource handles all node related requests from the simulation.
type NodeUpdatesResource interface {
	Post(misim.NodeUpdateRequest)
	FailNodes(misim.NodeFailureRequest) misim.NodeFailureResponse
	MarkNodesNotReady(misim.NodeNotReadyRequest) misim.NodeNotReadyResponse
	MarkNodesNoExecute(misim.NodeNoExecuteRequest) misim.NodeNoExecuteResponse
}

type nodeUpdatesResourceImpl struct {
	storage    *storage.StorageContainer
	controller *NodeController
}

// Post handles a node update request.
func (impl nodeUpdatesResourceImpl) Post(u misim.NodeUpdateRequest) {
	if len(u.MachineSets) == 0 {
		// No cluster scaling
		impl.controller.UpdateNodes(u.AllNodes, u.Events)
	} else {
		// If the request contains machines set, we use only the machines
		impl.controller.InitMachinesNodes(u.AllNodes, u.Events, u.MachineSets, u.Machines)
	}
}

// FailNodes marks the requested nodes as failed.
func (impl nodeUpdatesResourceImpl) FailNodes(req misim.NodeFailureRequest) misim.NodeFailureResponse {
	return impl.controller.HandleNodeFailure(req.FailedPods)
}

// MarkNodesNotReady marks the requested nodes as not ready.
func (impl nodeUpdatesResourceImpl) MarkNodesNotReady(req misim.NodeNotReadyRequest) misim.NodeNotReadyResponse {
	return impl.controller.HandleMarkNodeNotReady(req.Nodes)
}

func (impl nodeUpdatesResourceImpl) MarkNodesNoExecute(req misim.NodeNoExecuteRequest) misim.NodeNoExecuteResponse {
	return impl.controller.HandleMarkNodeNoExecute(req.Nodes)
}

// NewNodeUpdateResource creates a new node update resource.
func NewNodeUpdateResource(storage *storage.StorageContainer) NodeUpdatesResource {
	controller := NewNodeController(storage)
	return nodeUpdatesResourceImpl{
		storage:    storage,
		controller: &controller,
	}
}
