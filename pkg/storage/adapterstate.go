package storage

type IdStorage interface {
	GetNextId() int
	StoreNextId(id int)
}

type AdapterStateStorage interface {
	StoreClusterAutoscalerActive(active bool)
	IsClusterAutoscalerActive() bool
	StoreClusterAutoscalingDone(done bool)
	IsClusterAutoscalingDone() bool
}
