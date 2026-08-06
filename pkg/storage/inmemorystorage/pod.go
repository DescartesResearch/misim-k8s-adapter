package inmemorystorage

import (
	"context"
	"sync"

	"go-kube/internal/broadcast"
	"go-kube/pkg/misim"
	storage2 "go-kube/pkg/storage"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

type PodInMemoryStorage struct {
	mu sync.Mutex

	pods           core.PodList
	podEventChan   chan metav1.WatchEvent
	podBroadcaster *broadcast.BroadcastServer[metav1.WatchEvent]
	nextResourceId int

	failedPodBuffer   InMemBuffer[misim.BindingFailureInformation]
	bindedPodBuffer   InMemBuffer[misim.BindingInformation]
	podsToBePlaced    InMemBuffer[core.Pod]
	podsUpdateChannel InMemChannelWrapper[misim.PodsUpdateResponse]
}

func (s *PodInMemoryStorage) BeginTransaction() {
	s.mu.Lock()
}

func (s *PodInMemoryStorage) EndTransaction() {
	s.mu.Unlock()
}

func (s *PodInMemoryStorage) GetPods() (core.PodList, *broadcast.BroadcastServer[metav1.WatchEvent]) {
	return s.pods, s.podBroadcaster
}

func (s *PodInMemoryStorage) StorePods(pods core.PodList, events []metav1.WatchEvent) {
	s.pods = pods

	// First only modified/deleted events
	for _, e := range events {
		klog.V(3).Infof("EVENT TYPE: %s", e.Type)
		if e.Type == "MODIFIED" || e.Type == "DELETED" {
			s.podEventChan <- e
		}
	}
	// Second only added events
	for _, e := range events {
		if e.Type == "ADDED" {
			s.podEventChan <- e
		}
	}
}

func (s *PodInMemoryStorage) DeletePods(events []metav1.WatchEvent) {
	s.pods = core.PodList{}
	for _, e := range events {
		e.Type = "DELETED"
		s.podEventChan <- e
	}
}

func (s *PodInMemoryStorage) GetPod(podName string) core.Pod {
	var u core.Pod
	for _, element := range s.pods.Items {
		if element.Name == podName {
			klog.V(8).Infof("Found pod %s", podName)
			u = element
		}
	}
	return u
}

func (s *PodInMemoryStorage) DeletePod(podName string) core.Pod {
	var index int
	var deletedPod core.Pod
	for i, pod := range s.pods.Items {
		if pod.Name == podName {
			index = i
			deletedPod = pod
			break
		}
	}

	s.pods.Items[index] = s.pods.Items[len(s.pods.Items)-1]
	s.pods.Items = s.pods.Items[:len(s.pods.Items)-1]

	podDeleteEvent := metav1.WatchEvent{Type: "DELETED", Object: runtime.RawExtension{Object: &deletedPod}}
	s.podEventChan <- podDeleteEvent
	return deletedPod
}

func (s *PodInMemoryStorage) UpdatePod(podName string, newValues core.Pod) {
	// Find index, and replace
	var index int = -1
	for i, element := range s.pods.Items {
		if element.Name == podName {
			klog.V(8).Info("Found pod %s", podName)
			index = i
			break
		}
	}
	if index != -1 {
		// Found in list => update
		s.pods.Items[index] = newValues

		// Fire modified watch event
		s.podEventChan <- metav1.WatchEvent{
			Type:   "MODIFIED",
			Object: runtime.RawExtension{Object: &s.pods.Items[index]},
		}
	}
	// else do nothing
}

func (s *PodInMemoryStorage) FailedPodBuffer() storage2.Buffer[misim.BindingFailureInformation] {
	return &s.failedPodBuffer
}

func (s *PodInMemoryStorage) BindedPodBuffer() storage2.Buffer[misim.BindingInformation] {
	return &s.bindedPodBuffer
}

func (s *PodInMemoryStorage) PodsToBePlaced() storage2.Buffer[core.Pod] {
	return &s.podsToBePlaced
}

func (s *PodInMemoryStorage) PodsUpdateChannel() storage2.ChannelWrapper[misim.PodsUpdateResponse] {
	return &s.podsUpdateChannel
}

func NewPodInMemoryStorage() PodInMemoryStorage {
	podEventChan := make(chan metav1.WatchEvent, 500)
	return PodInMemoryStorage{
		pods:           core.PodList{TypeMeta: metav1.TypeMeta{Kind: "PodList", APIVersion: "v1"}, Items: nil},
		podEventChan:   podEventChan,
		podBroadcaster: broadcast.NewBroadcastServer(context.TODO(), "PodBroadcaster", podEventChan),
		nextResourceId: 1,

		failedPodBuffer:   NewInMemBuffer[misim.BindingFailureInformation](),
		bindedPodBuffer:   NewInMemBuffer[misim.BindingInformation](),
		podsToBePlaced:    NewInMemBuffer[core.Pod](),
		podsUpdateChannel: NewInMemChannelWrapper[misim.PodsUpdateResponse](),
	}
}
