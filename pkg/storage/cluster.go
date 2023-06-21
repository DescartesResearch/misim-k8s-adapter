package storage

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube-rise/internal/broadcast"
	"net/http"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

type MachineSetStorage interface {
	// Stores a machineset list in the storage
	StoreMachineSets(ms cluster.MachineSetList, events []metav1.WatchEvent)
	// Returns the current machinesets
	GetMachineSets() (cluster.MachineSetList, *broadcast.BroadcastServer[metav1.WatchEvent])
	// Returns the current scale of the machineset
	GetMachineSetsScale(w http.ResponseWriter, r *http.Request)
	// Puts the current scale of the machineset
	PutMachineSetsScale(w http.ResponseWriter, r *http.Request)
	// Is a upscaling possible on any MachineSet
	IsUpscalingPossible() bool
	// Is a dowscaling possible on any MachineSet
	IsDownscalingPossible() bool
}

type MachineStorage interface {
	// Stores a machineset list in the storage
	StoreMachines(ms cluster.MachineList, events []metav1.WatchEvent)
	// Returns the current machinesets
	GetMachines() (cluster.MachineList, *broadcast.BroadcastServer[metav1.WatchEvent])
	// Gets a single machine
	GetMachine(w http.ResponseWriter, r *http.Request)
	// Puts a machine
	PutMachine(w http.ResponseWriter, r *http.Request)
	// Scales machines
	ScaleMachines(machineSet cluster.MachineSet, changedNodes []core.Node, amount int) ([]cluster.Machine, error)
}

type StatusConfigMapStorage interface {
	// Stores the status config map
	StoreStatusConfigMap(w http.ResponseWriter, r *http.Request)
	// Returns the current status config map
	GetStatusConfigMap() core.ConfigMap
}
