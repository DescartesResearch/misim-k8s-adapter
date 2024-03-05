package main

import (
	"flag"
	"go-kube/pkg/interfaces"
	"go-kube/pkg/storage"
	"go-kube/pkg/storage/inmemorystorage"

	"k8s.io/klog/v2"
)

func initStorages() storage.StorageContainer {
	var podStorage = inmemorystorage.NewPodInMemoryStorage()
	var nodeStorage = inmemorystorage.NewNodeInMemoryStorage()
	var namespaceStorage = inmemorystorage.NewNamespaceInMemoryStorage()
	var daemonSetStorage = inmemorystorage.NewDaemonSetInMemoryStorage()
	var machineStorage = inmemorystorage.NewMachineInMemoryStorage()
	var machineSetStorage = inmemorystorage.NewMachineSetInMemoryStorage(&nodeStorage, &machineStorage)
	var statusConfigMapStorage = inmemorystorage.NewStatusMapInMemoryStorage()
	var podIdStorage = inmemorystorage.NewIdInMemoryStorage()
	var machineIdStorage = inmemorystorage.NewIdInMemoryStorage()
	var adapterStateStorage = inmemorystorage.NewAdapterStateInMemoryStorage()
	var eventStorage = inmemorystorage.NewEventInMemoryStorage()

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
	flag.Parse()        // parses the command-line flags
	var storages = initStorages()
	var app = interfaces.NewAdapterApplication(&storages)
	app.Start()
}
