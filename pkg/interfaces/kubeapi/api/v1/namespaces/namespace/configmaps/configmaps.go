package configmaps

import (
	clusterautoscalerstatus "go-kube/pkg/interfaces/kubeapi/api/v1/namespaces/namespace/configmaps/cluster-autoscaler-status"
	"go-kube/pkg/storage"
)

type ConfigmapsResource interface {
	ClusterAutoscalerStatus() clusterautoscalerstatus.ClusterAutoscalerStatusResource
}

type ConfigmapsResourceImpl struct {
	namespaceName string
	storage       *storage.StorageContainer
}

func (impl ConfigmapsResourceImpl) ClusterAutoscalerStatus() clusterautoscalerstatus.ClusterAutoscalerStatusResource {
	return clusterautoscalerstatus.NewClusterAutoscalerStatusResource(impl.namespaceName, impl.storage)
}

func NewConfigmapsResource(name string, storage *storage.StorageContainer) ConfigmapsResourceImpl {
	return ConfigmapsResourceImpl{
		namespaceName: name,
		storage:       storage,
	}
}
