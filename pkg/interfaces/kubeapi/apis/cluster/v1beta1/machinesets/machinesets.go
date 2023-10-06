package machinesets

import (
	"go-kube/internal/broadcast"
	"go-kube/pkg/interfaces/kubeapi/apis/cluster/v1beta1/machinesets/machineset"
	"go-kube/pkg/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

type MachineSetsResource interface {
	Get() (cluster.MachineSetList, *broadcast.BroadcastServer[metav1.WatchEvent])
	MachineSet(machineSetName string) machineset.MachineSetResource
}

type MachineSetsResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl MachineSetsResourceImpl) Get() (cluster.MachineSetList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return impl.storage.MachineSets.GetMachineSets()
}

func (impl MachineSetsResourceImpl) MachineSet(name string) machineset.MachineSetResource {
	return machineset.NewMachineSetResource(name, impl.storage)
}

func NewMachineSetsResource(storage *storage.StorageContainer) MachineSetsResource {
	return MachineSetsResourceImpl{storage: storage}
}
