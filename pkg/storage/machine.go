package storage

import (
	"go-kube/internal/broadcast"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

type MachineStorage interface {
	// Stores a machineset list in the storage
	StoreMachines(ms cluster.MachineList, events []metav1.WatchEvent)
	// Returns the current machinesets
	GetMachines() (cluster.MachineList, *broadcast.BroadcastServer[metav1.WatchEvent])
	// Gets a single machine
	// GetMachine(w http.ResponseWriter, r *http.Request)
	GetMachine(machineName string) cluster.Machine
	// Deletes the machine
	DeleteMachine(machineName string) cluster.Machine
	AddMachine(cluster.Machine)
	// Puts a machine
	// PutMachine(w http.ResponseWriter, r *http.Request)
	PutMachine(machineName string, machine cluster.Machine) cluster.Machine
	IncrementMachineCount()
	GetMachineCount() int
}
