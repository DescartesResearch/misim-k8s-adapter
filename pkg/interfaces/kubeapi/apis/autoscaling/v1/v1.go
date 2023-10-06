package v1

import (
	"go-kube/pkg/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type V1Resource interface {
	Get() metav1.APIResourceList
}

type V1ResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl V1ResourceImpl) Get() metav1.APIResourceList {
	return metav1.APIResourceList{TypeMeta: metav1.TypeMeta{Kind: "APIResourceList", APIVersion: "v1"},
		GroupVersion: "autoscaling/v1",
		APIResources: []metav1.APIResource{
			{
				Name:         "horizontalpodautoscalers",
				SingularName: "",
				Namespaced:   true,
				Kind:         "HorizontalPodAutoscaler",
				Verbs:        []string{"delete", "deletecollection", "get", "list", "patch", "create", "update", "watch"},
				Categories:   []string{"all"},
			},
			{
				Name:         "horizontalpodautoscalers/status",
				SingularName: "",
				Namespaced:   true,
				Kind:         "HorizontalPodAutoscaler",
				Verbs:        []string{"get", "patch", "update"},
			},
		},
	}
}

func NewV1Resource(storage *storage.StorageContainer) V1Resource {
	return V1ResourceImpl{storage: storage}
}
