package machines

import (
	"go-kube/internal/broadcast"
	"go-kube/pkg/interfaces/kubeapi/apis/cluster/v1beta1/machines/machine"
	"go-kube/pkg/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

type MachinesResource interface {
	Get() (cluster.MachineList, *broadcast.BroadcastServer[metav1.WatchEvent])
	Machine(machineName string) machine.MachineResource
}

type MachinesResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl MachinesResourceImpl) Get() (cluster.MachineList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return impl.storage.Machines.GetMachines()
}

func (impl MachinesResourceImpl) Machine(machineName string) machine.MachineResource {
	return machine.NewMachineResource(machineName, impl.storage)
}

func NewMachinesResource(storage *storage.StorageContainer) MachinesResource {
	return MachinesResourceImpl{storage: storage}
}
