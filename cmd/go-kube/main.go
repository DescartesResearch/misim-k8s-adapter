package main

import (
	"flag"

	"go-kube/pkg/config"
	"go-kube/pkg/control"
	"go-kube/pkg/interfaces"
	"go-kube/pkg/storage"
	"go-kube/pkg/storage/inmemorystorage"

	"k8s.io/klog/v2"
)

func initStorages() storage.StorageContainer {
	podStorage := inmemorystorage.NewPodInMemoryStorage()
	nodeStorage := inmemorystorage.NewNodeInMemoryStorage()
	namespaceStorage := inmemorystorage.NewNamespaceInMemoryStorage()
	daemonSetStorage := inmemorystorage.NewDaemonSetInMemoryStorage()
	machineStorage := inmemorystorage.NewMachineInMemoryStorage()
	machineSetStorage := inmemorystorage.NewMachineSetInMemoryStorage(&nodeStorage, &machineStorage)
	statusConfigMapStorage := inmemorystorage.NewStatusMapInMemoryStorage()
	podIdStorage := inmemorystorage.NewIdInMemoryStorage()
	machineIdStorage := inmemorystorage.NewIdInMemoryStorage()
	adapterStateStorage := inmemorystorage.NewAdapterStateInMemoryStorage()
	eventStorage := inmemorystorage.NewEventInMemoryStorage()

	return storage.StorageContainer{
		Pods:            &podStorage,
		Nodes:           &nodeStorage,
		Namespaces:      &namespaceStorage,
		DaemonSets:      &daemonSetStorage,
		Machines:        &machineStorage,
		MachineSets:     &machineSetStorage,
		StatusConfigMap: &statusConfigMapStorage,
		PodIds:          &podIdStorage,
		MachineIds:      &machineIdStorage,
		AdapterState:    &adapterStateStorage,
		Events:          &eventStorage,
	}
}

func main() {
	klog.InitFlags(nil) // initializing the flags
	defer klog.Flush()  // flushes all pending log I/O
	config.InitConfigFlags()
	flag.Parse() // parses the command-line flags
	control.Init(config.Seed)
	storages := initStorages()
	app := interfaces.NewAdapterApplication(&storages)
	app.Start()
}
