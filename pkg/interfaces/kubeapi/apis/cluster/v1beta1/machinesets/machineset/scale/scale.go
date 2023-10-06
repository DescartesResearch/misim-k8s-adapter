package scale

import (
	"go-kube/pkg/control"
	"go-kube/pkg/storage"
	v1 "k8s.io/api/autoscaling/v1"
)

type ScaleResource interface {
	Get() v1.Scale
	Put(v1.Scale) v1.Scale
}

type ScaleResourceImpl struct {
	machineSetName string
	storage        *storage.StorageContainer
}

func (impl ScaleResourceImpl) Get() v1.Scale {
	return impl.storage.MachineSets.GetMachineSetsScale(impl.machineSetName)
}

func (impl ScaleResourceImpl) Put(m v1.Scale) v1.Scale {
	controller := control.NewScaleController(impl.storage)
	controller.ScaleMachineSet(impl.machineSetName, m)
	return impl.storage.MachineSets.GetMachineSetsScale(impl.machineSetName)
}

func NewScaleResource(name string, storage *storage.StorageContainer) ScaleResource {
	return ScaleResourceImpl{
		machineSetName: name,
		storage:        storage,
	}
}
