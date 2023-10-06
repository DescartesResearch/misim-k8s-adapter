package status

import (
	"go-kube/pkg/control"
	"go-kube/pkg/storage"
	v1 "k8s.io/api/core/v1"
)

type StatusResource interface {
	Patch(status v1.PodStatusResult) v1.Pod
}

type StatusResourceImpl struct {
	podName          string
	storageContainer *storage.StorageContainer
}

func (impl StatusResourceImpl) Patch(status v1.PodStatusResult) v1.Pod {
	controller := control.NewPodController(impl.storageContainer)
	// We always assume this means it's failed
	controller.FailedPod(impl.podName, status.Status)
	return impl.storageContainer.Pods.GetPod(impl.podName)
}

func NewStatusResource(podName string, storage *storage.StorageContainer) StatusResourceImpl {
	return StatusResourceImpl{
		podName:          podName,
		storageContainer: storage,
	}
}
