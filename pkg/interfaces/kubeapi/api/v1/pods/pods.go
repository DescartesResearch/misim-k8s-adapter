package pods

import (
	"go-kube/internal/broadcast"
	"go-kube/pkg/interfaces/kubeapi/api/v1/pods/pod"
	"go-kube/pkg/storage"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodsResource interface {
	Get() (v1.PodList, *broadcast.BroadcastServer[metav1.WatchEvent])
	Pod(podName string) pod.PodResource
}

type PodsResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl PodsResourceImpl) Get() (v1.PodList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return impl.storage.Pods.GetPods()
}

func (impl PodsResourceImpl) Pod(podName string) pod.PodResource {
	return pod.NewPodResource(podName, impl.storage)
}

func NewPodsResource(storage *storage.StorageContainer) PodsResourceImpl {
	return PodsResourceImpl{
		storage: storage,
	}
}
