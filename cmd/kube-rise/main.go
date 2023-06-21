package main

import (
	"kube-rise/internal/inmemorystorage"
	"kube-rise/pkg/server"
	"kube-rise/pkg/storage"
)

func initStorages() storage.StorageContainer {
	var podStorage = inmemorystorage.NewPodInMemoryStorage()
	var nodeStorage = inmemorystorage.NewNodeInMemoryStorage()
	var namespaceStorage = inmemorystorage.NewNamespaceInMemoryStorage()
	var daemonSetStorage = inmemorystorage.NewDaemonSetInMemoryStorage()
	var machineStorage = inmemorystorage.NewMachineInMemoryStorage()
	var machineSetStorage = inmemorystorage.NewMachineSetInMemoryStorage(&nodeStorage, &machineStorage)
	var statusConfigMapStorage = inmemorystorage.NewStatusMapInMemoryStorage()

	return storage.StorageContainer{
		Pods:            &podStorage,
		Nodes:           &nodeStorage,
		Namespaces:      &namespaceStorage,
		DaemonSets:      &daemonSetStorage,
		Machines:        &machineStorage,
		MachineSets:     &machineSetStorage,
		StatusConfigMap: &statusConfigMapStorage,
	}
}

func main() {
	var storages = initStorages()
	var app = server.NewAdapterApplication(&storages)
	app.Start()
}
