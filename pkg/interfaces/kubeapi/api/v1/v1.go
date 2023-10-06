package v1

import (
	"go-kube/pkg/interfaces/kubeapi/api/v1/namespaces"
	"go-kube/pkg/interfaces/kubeapi/api/v1/nodes"
	"go-kube/pkg/interfaces/kubeapi/api/v1/pods"
	"go-kube/pkg/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type V1Resource interface {
	Get() metav1.APIResourceList
	Nodes() nodes.NodesResource
	Pods() pods.PodsResource
	Namespaces() namespaces.NamespacesResource
}

type V1ResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl V1ResourceImpl) Nodes() nodes.NodesResource {
	return nodes.NewNodesResource(impl.storage)
}

func (impl V1ResourceImpl) Pods() pods.PodsResource {
	return pods.NewPodsResource(impl.storage)
}

func (impl V1ResourceImpl) Namespaces() namespaces.NamespacesResource {
	return namespaces.NewNamespacesResource(impl.storage)
}

func (impl V1ResourceImpl) Get() metav1.APIResourceList {
	return metav1.APIResourceList{TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "APIResourceList"},
		GroupVersion: "v1",
		APIResources: []metav1.APIResource{
			{
				Name:         "namespaces",
				SingularName: "",
				Namespaced:   false,
				Kind:         "Namespace",
				Verbs:        []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			},
			{
				Name:         "nodes",
				SingularName: "",
				Namespaced:   false,
				Kind:         "Node",
				Verbs:        []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"},
			},
			{
				Name:         "nodes/status",
				SingularName: "",
				Namespaced:   false,
				Kind:         "Node",
				Verbs:        []string{"get", "patch", "update"},
			},
			{
				Name:         "persistentvolumeclaims",
				SingularName: "",
				Namespaced:   true,
				Kind:         "PersistentVolumeClaim",
				Verbs:        []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"},
			},
			{
				Name:         "persistentvolumes",
				SingularName: "",
				Namespaced:   false,
				Kind:         "PersistentVolume",
				Verbs:        []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"},
			},
			{
				Name:         "pods",
				SingularName: "",
				Namespaced:   true,
				Kind:         "Pod",
				Verbs:        []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"},
				Categories:   []string{"all"},
			},
			{
				Name:         "pods/binding",
				SingularName: "",
				Namespaced:   true,
				Kind:         "Binding",
				Verbs:        []string{"create"},
			},
			{
				Name:         "pods/eviction",
				SingularName: "",
				Namespaced:   true,
				Group:        "policy",
				Version:      "v1",
				Kind:         "Eviction",
				Verbs:        []string{"create"},
			},
			{
				Name:         "pods/status",
				SingularName: "",
				Namespaced:   true,
				Kind:         "Pod",
				Verbs:        []string{"get", "patch", "update"},
			},
			{
				Name:         "replicationcontrollers",
				SingularName: "",
				Namespaced:   true,
				Kind:         "ReplicationController",
				Verbs:        []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"},
				Categories:   []string{"all"},
			},
			{
				Name:         "services",
				SingularName: "",
				Namespaced:   true,
				Kind:         "Service",
				Verbs:        []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"},
				Categories:   []string{"all"},
			},
		},
	}
}

func NewV1Resource(storage *storage.StorageContainer) V1ResourceImpl {
	return V1ResourceImpl{
		storage: storage,
	}
}
