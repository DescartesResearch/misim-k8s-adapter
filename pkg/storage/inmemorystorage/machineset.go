package inmemorystorage

import (
	"context"
	"go-kube/internal/broadcast"
	"strconv"

	v1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

type MachineSetsInMemoryStorage struct {
	machineSets           cluster.MachineSetList
	machineSetsEventChan  chan metav1.WatchEvent
	nodeStorage           *NodeInMemoryStorage
	machineStorage        *MachineInMemoryStorage
	machineSetBroadcaster *broadcast.BroadcastServer[metav1.WatchEvent]
}

func (s *MachineSetsInMemoryStorage) GetMachineSets() (cluster.MachineSetList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return s.machineSets, s.machineSetBroadcaster
}

func (s *MachineSetsInMemoryStorage) StoreMachineSets(ms cluster.MachineSetList, events []metav1.WatchEvent) {
	s.machineSets = ms
	for _, e := range events {
		s.machineSetsEventChan <- e
	}
}

func (s *MachineSetsInMemoryStorage) GetMachineSet(machineSetName string) cluster.MachineSet {
	var machineSet cluster.MachineSet
	for _, set := range s.machineSets.Items {
		if set.Name == machineSetName {
			machineSet = set
			break
		}
	}
	return machineSet
}

func (s *MachineSetsInMemoryStorage) PutMachineSet(machineSetName string, machineSet cluster.MachineSet) cluster.MachineSet {
	index := -1
	for i, set := range s.machineSets.Items {
		if set.Name == machineSetName {
			index = i
			break
		}
	}
	s.machineSets.Items[index] = machineSet
	// Fire MODIFIED event
	s.machineSetsEventChan <- metav1.WatchEvent{Type: "MODIFIED", Object: runtime.RawExtension{Object: &machineSet}}
	return machineSet
}

func (s *MachineSetsInMemoryStorage) GetMachineSetsScale(machineSetName string) v1.Scale {
	var machineSetRef cluster.MachineSet
	for _, ms := range s.machineSets.Items {
		if ms.Name == machineSetName {
			machineSetRef = ms
			break
		}
	}

	result := v1.Scale{TypeMeta: metav1.TypeMeta{APIVersion: "autoscaling/v1", Kind: "Scale"},
		ObjectMeta: metav1.ObjectMeta{Name: machineSetName},
		Spec:       v1.ScaleSpec{Replicas: *machineSetRef.Spec.Replicas},
		Status:     v1.ScaleStatus{Replicas: *machineSetRef.Spec.Replicas}}

	return result
}

func (s *MachineSetsInMemoryStorage) IsUpscalingPossible() bool {
	for _, ms := range s.machineSets.Items {
		maxSize, _ := strconv.Atoi(ms.Annotations["cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size"])
		if *ms.Spec.Replicas < int32(maxSize) {
			return true
		}
	}
	return false
}

func (s *MachineSetsInMemoryStorage) IsDownscalingPossible() bool {
	for _, ms := range s.machineSets.Items {
		minSize, _ := strconv.Atoi(ms.Annotations["cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size"])
		if *ms.Spec.Replicas > int32(minSize) {
			return true
		}
	}
	return false
}

func NewMachineSetInMemoryStorage(nodeStorage *NodeInMemoryStorage, machineStorage *MachineInMemoryStorage) MachineSetsInMemoryStorage {
	machineSetsEventChan := make(chan metav1.WatchEvent, 500)
	return MachineSetsInMemoryStorage{
		machineSets:           cluster.MachineSetList{TypeMeta: metav1.TypeMeta{Kind: "MachineSetList", APIVersion: "cluster-x.k8s.io/v1beta1"}, Items: nil},
		machineSetsEventChan:  machineSetsEventChan,
		nodeStorage:           nodeStorage,
		machineStorage:        machineStorage,
		machineSetBroadcaster: broadcast.NewBroadcastServer(context.TODO(), "MachineSetBroadcaster", machineSetsEventChan),
	}
}
