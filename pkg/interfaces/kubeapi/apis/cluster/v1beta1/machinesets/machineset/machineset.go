package machineset

import (
	"go-kube/pkg/interfaces/kubeapi/apis/cluster/v1beta1/machinesets/machineset/scale"
	"go-kube/pkg/storage"
)

type MachineSetResource interface {
	Scale() scale.ScaleResource
}

type MachineSetResourceImpl struct {
	machineSetName string
	storage        *storage.StorageContainer
}

func (impl MachineSetResourceImpl) Scale() scale.ScaleResource {
	return scale.NewScaleResource(impl.machineSetName, impl.storage)
}

func NewMachineSetResource(name string, storage *storage.StorageContainer) MachineSetResource {
	return MachineSetResourceImpl{
		machineSetName: name,
		storage:        storage,
	}
}
