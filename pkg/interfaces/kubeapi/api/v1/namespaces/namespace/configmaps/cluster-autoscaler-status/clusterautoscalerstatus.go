package clusterautoscalerstatus

import (
	"go-kube/pkg/storage"
	core "k8s.io/api/core/v1"
)

type ClusterAutoscalerStatusResource interface {
	Get() core.ConfigMap
	Put(core.ConfigMap) core.ConfigMap
}

type ClusterAutoscalerStatusResourceImpl struct {
	namespaceName string
	storage       *storage.StorageContainer
}

func (impl ClusterAutoscalerStatusResourceImpl) Get() core.ConfigMap {
	return impl.storage.StatusConfigMap.GetStatusConfigMap()
}

func (impl ClusterAutoscalerStatusResourceImpl) Put(configMap core.ConfigMap) core.ConfigMap {
	impl.storage.StatusConfigMap.StoreStatusConfigMap(configMap)
	return configMap
}

func NewClusterAutoscalerStatusResource(namespace string, storage *storage.StorageContainer) ClusterAutoscalerStatusResourceImpl {
	return ClusterAutoscalerStatusResourceImpl{
		namespaceName: namespace,
		storage:       storage,
	}
}
