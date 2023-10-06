package namespace

import (
	"go-kube/pkg/interfaces/kubeapi/apis/cluster/v1beta1/machines"
	"go-kube/pkg/interfaces/kubeapi/apis/cluster/v1beta1/machinesets"
	"go-kube/pkg/storage"
)

type NamespaceResource interface {
	Machines() machines.MachinesResource
	MachineSets() machinesets.MachineSetsResource
}

type NamespaceResourceImpl struct {
	namespaceName string
	storage       *storage.StorageContainer
}

func (impl NamespaceResourceImpl) Machines() machines.MachinesResource {
	// TODO [Support for multiple namespaces]: pass namespace name somehow if we want to support multiple namespaces
	return machines.NewMachinesResource(impl.storage)
}

func (impl NamespaceResourceImpl) MachineSets() machinesets.MachineSetsResource {
	// TODO [Support for multiple namespaces]: pass namespace name somehow if we want to support multiple namespaces
	return machinesets.NewMachineSetsResource(impl.storage)
}

func NewNamespaceResource(name string, storage *storage.StorageContainer) NamespaceResource {
	return NamespaceResourceImpl{
		namespaceName: name,
		storage:       storage,
	}
}
