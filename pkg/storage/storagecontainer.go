package storage

type StorageContainer struct {
	Pods            PodStorage
	Nodes           NodeStorage
	Namespaces      NamespaceStorage
	DaemonSets      DaemonSetStorage
	Machines        MachineStorage
	MachineSets     MachineSetStorage
	StatusConfigMap StatusConfigMapStorage
	PodIds          IdStorage
	MachineIds      IdStorage
	AdapterState    AdapterStateStorage
}
