package inmemorystorage

type IdInMemoryStorage struct {
	nextId int
}

func (s *IdInMemoryStorage) GetNextId() int {
	return s.nextId
}

func (s *IdInMemoryStorage) StoreNextId(id int) {
	s.nextId = id
}

func NewIdInMemoryStorage() IdInMemoryStorage {
	return IdInMemoryStorage{
		nextId: 1,
	}
}

type AdapterStateInMemoryStorage struct {
	clusterAutoscalerActive bool
	clusterAutoscalingDone  bool
}

func (s *AdapterStateInMemoryStorage) StoreClusterAutoscalerActive(active bool) {
	s.clusterAutoscalerActive = active
}

func (s *AdapterStateInMemoryStorage) IsClusterAutoscalerActive() bool {
	return s.clusterAutoscalerActive
}

func (s *AdapterStateInMemoryStorage) StoreClusterAutoscalingDone(done bool) {
	s.clusterAutoscalingDone = done
}

func (s *AdapterStateInMemoryStorage) IsClusterAutoscalingDone() bool {
	return s.clusterAutoscalingDone
}

func NewAdapterStateInMemoryStorage() AdapterStateInMemoryStorage {
	return AdapterStateInMemoryStorage{
		clusterAutoscalerActive: false,
		clusterAutoscalingDone:  true,
	}
}
