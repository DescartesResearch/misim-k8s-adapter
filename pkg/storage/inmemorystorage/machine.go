package inmemorystorage

import (
	"context"
	"go-kube/internal/broadcast"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

type MachineInMemoryStorage struct {
	machines           cluster.MachineList
	machineEventChan   chan metav1.WatchEvent
	machineBroadcaster *broadcast.BroadcastServer[metav1.WatchEvent]
	machineCount       int
}

func (s *MachineInMemoryStorage) GetMachines() (cluster.MachineList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return s.machines, s.machineBroadcaster
}

func (s *MachineInMemoryStorage) StoreMachines(ms cluster.MachineList, events []metav1.WatchEvent) {
	s.machines = ms
	for _, n := range events {
		s.machineEventChan <- n
		if n.Type == "ADDED" {
			s.IncrementMachineCount()
		}
	}
}

func (s *MachineInMemoryStorage) GetMachine(machineName string) cluster.Machine {
	var machineRef cluster.Machine
	for _, ms := range s.machines.Items {
		if ms.Name == machineName {
			machineRef = ms
			break
		}
	}
	return machineRef
}

func (s *MachineInMemoryStorage) PutMachine(machineName string, u cluster.Machine) cluster.Machine {
	indexForReplacement := -1
	for index, machine := range s.machines.Items {
		if machine.Name == machineName {
			indexForReplacement = index
			break
		}
	}
	s.machines.Items[indexForReplacement] = u
	return u
}

func (s *MachineInMemoryStorage) AddMachine(machine cluster.Machine) {
	s.machines.Items = append(s.machines.Items, machine)
	// Fire watch event
	machineAddEvent := metav1.WatchEvent{Type: "ADDED", Object: runtime.RawExtension{Object: &machine}}
	s.machineEventChan <- machineAddEvent
}

func (s *MachineInMemoryStorage) DeleteMachine(machineName string) cluster.Machine {
	var index int
	var deletedMachine cluster.Machine
	for i, machine := range s.machines.Items {
		if machine.Name == machineName {
			index = i
			deletedMachine = machine
			break
		}
	}
	s.machines.Items[index] = s.machines.Items[len(s.machines.Items)-1]
	s.machines.Items = s.machines.Items[:len(s.machines.Items)-1]
	// Fire deleted event
	s.machineEventChan <- metav1.WatchEvent{Type: "DELETED", Object: runtime.RawExtension{Object: &s.machines.Items[index]}}
	return deletedMachine
}

func (s *MachineInMemoryStorage) IncrementMachineCount() {
	s.machineCount = s.machineCount + 1
	klog.V(4).Infof("Incremented machine count to %d", s.machineCount)
}

func (s *MachineInMemoryStorage) GetMachineCount() int {
	return s.machineCount
}

func NewMachineInMemoryStorage() MachineInMemoryStorage {
	machineEventChan := make(chan metav1.WatchEvent, 500)
	return MachineInMemoryStorage{
		machines:           cluster.MachineList{TypeMeta: metav1.TypeMeta{Kind: "MachineList", APIVersion: "cluster.x-k8s.io/v1beta1"}, Items: nil},
		machineEventChan:   machineEventChan,
		machineBroadcaster: broadcast.NewBroadcastServer(context.TODO(), "MachineBroadcaster", machineEventChan),
		machineCount:       0,
	}
}
