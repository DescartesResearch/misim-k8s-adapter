package autoscaling

import (
	v1 "go-kube/pkg/interfaces/kubeapi/apis/autoscaling/v1"
	"go-kube/pkg/storage"
)

type AutoscalingResource interface {
	V1() v1.V1Resource
}

type AutoscalingResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl AutoscalingResourceImpl) V1() v1.V1Resource {
	return v1.NewV1Resource(impl.storage)
}

func NewAutoscalingResource(storage *storage.StorageContainer) AutoscalingResource {
	return AutoscalingResourceImpl{
		storage: storage,
	}
}
