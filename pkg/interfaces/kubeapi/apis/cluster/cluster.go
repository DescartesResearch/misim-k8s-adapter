package cluster

import (
	"go-kube/pkg/interfaces/kubeapi/apis/cluster/v1beta1"
	"go-kube/pkg/storage"
)

type ClusterResource interface {
	V1Beta1() v1beta1.V1Beta1Resource
}

type ClusterResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl ClusterResourceImpl) V1Beta1() v1beta1.V1Beta1Resource {
	return v1beta1.NewV1Beta1Resource(impl.storage)
}

func NewClusterResource(storage *storage.StorageContainer) ClusterResource {
	return ClusterResourceImpl{
		storage: storage,
	}
}
