package machine

import (
	"go-kube/pkg/storage"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

type MachineResource interface {
	Get() cluster.Machine
	Put(cluster.Machine) cluster.Machine
}

type MachineResourceImpl struct {
	machineName string
	storage     *storage.StorageContainer
}

func (impl MachineResourceImpl) Get() cluster.Machine {
	return impl.storage.Machines.GetMachine(impl.machineName)
}

func (impl MachineResourceImpl) Put(m cluster.Machine) cluster.Machine {
	return impl.storage.Machines.PutMachine(impl.machineName, m)
}

func NewMachineResource(name string, storage *storage.StorageContainer) MachineResource {
	return MachineResourceImpl{
		machineName: name,
		storage:     storage,
	}
}
