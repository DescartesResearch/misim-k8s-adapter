package storage

import (
	"go-kube/internal/broadcast"
	v1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

type MachineSetStorage interface {
	// Stores a machineset list in the storage
	StoreMachineSets(ms cluster.MachineSetList, events []metav1.WatchEvent)
	// Returns the current machinesets
	GetMachineSets() (cluster.MachineSetList, *broadcast.BroadcastServer[metav1.WatchEvent])
	// Finds machineset by name
	GetMachineSet(machineSetName string) cluster.MachineSet
	PutMachineSet(machineSetName string, machineSet cluster.MachineSet) cluster.MachineSet
	// Is a upscaling possible on any MachineSet
	IsUpscalingPossible() bool
	// Is a dowscaling possible on any MachineSet
	IsDownscalingPossible() bool
	// Get scale
	GetMachineSetsScale(machineSetName string) v1.Scale
}
