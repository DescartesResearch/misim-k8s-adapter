package v1beta1

import (
	"go-kube/pkg/interfaces/kubeapi/apis/cluster/v1beta1/machines"
	"go-kube/pkg/interfaces/kubeapi/apis/cluster/v1beta1/machinesets"
	"go-kube/pkg/interfaces/kubeapi/apis/cluster/v1beta1/namespaces"
	"go-kube/pkg/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type V1Beta1Resource interface {
	Get() metav1.APIResourceList
	Machines() machines.MachinesResource
	MachineSets() machinesets.MachineSetsResource
	Namespaces() namespaces.NamespacesResource
}

type V1Beta1ResourceImpl struct {
	storage *storage.StorageContainer
}

func (impl V1Beta1ResourceImpl) Get() metav1.APIResourceList {
	return metav1.APIResourceList{TypeMeta: metav1.TypeMeta{Kind: "APIResourceList", APIVersion: "v1"},
		GroupVersion: "cluster.x-k8s.io/v1beta1",
		APIResources: []metav1.APIResource{
			{
				Name:         "machines",
				SingularName: "machine",
				Namespaced:   true,
				Group:        "cluster.x-k8s.io",
				Version:      "v1beta1",
				Kind:         "Machine",
				Verbs:        []string{"delete", "deletecollection", "get", "list", "patch", "create", "update", "watch"},
				Categories:   []string{"cluster-api"},
			},
			{
				Name:         "machines/status",
				SingularName: "",
				Namespaced:   true,
				Group:        "cluster.x-k8s.io",
				Version:      "v1beta1",
				Kind:         "Machine",
				Verbs:        []string{"get", "patch", "update"},
			},
			{
				Name:         "clusters",
				SingularName: "cluster",
				Namespaced:   true,
				Group:        "cluster.x-k8s.io",
				Version:      "v1beta1",
				Kind:         "Cluster",
				Verbs:        []string{"delete", "deletecollection", "get", "list", "patch", "create", "update", "watch"},
				Categories:   []string{"cluster-api"},
			},
			{
				Name:         "clusters/status",
				SingularName: "",
				Namespaced:   true,
				Group:        "cluster.x-k8s.io",
				Version:      "v1beta1",
				Kind:         "Cluster",
				Verbs:        []string{"get", "patch", "update"},
			},
			{
				Name:         "machinesets",
				SingularName: "machineset",
				Namespaced:   true,
				Group:        "cluster.x-k8s.io",
				Version:      "v1beta1",
				Kind:         "MachineSet",
				Verbs:        []string{"delete", "deletecollection", "get", "list", "patch", "create", "update", "watch"},
				Categories:   []string{"cluster-api"},
			},
			{
				Name:         "machinesets/status",
				SingularName: "",
				Namespaced:   true,
				Group:        "cluster.x-k8s.io",
				Version:      "v1beta1",
				Kind:         "MachineSet",
				Verbs:        []string{"get", "patch", "update"},
			},
			{
				Name:         "machinesets/scale",
				SingularName: "",
				Namespaced:   true,
				Group:        "autoscaling",
				Version:      "v1",
				Kind:         "Scale",
				Verbs:        []string{"get", "patch", "update"},
			},
		},
	}
}

func (impl V1Beta1ResourceImpl) Machines() machines.MachinesResource {
	return machines.NewMachinesResource(impl.storage)
}

func (impl V1Beta1ResourceImpl) MachineSets() machinesets.MachineSetsResource {
	return machinesets.NewMachineSetsResource(impl.storage)
}

func (impl V1Beta1ResourceImpl) Namespaces() namespaces.NamespacesResource {
	return namespaces.NewNamespacesResource(impl.storage)
}

func NewV1Beta1Resource(storage *storage.StorageContainer) V1Beta1Resource {
	return V1Beta1ResourceImpl{storage: storage}
}
