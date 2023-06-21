package mocks

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetApiVersions() metav1.APIVersions {
	return metav1.APIVersions{
		Versions: []string{"v1"},
	}
}

func GetV1Api() metav1.APIResourceList {
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

func GetApis() metav1.APIGroupList {
	return metav1.APIGroupList{
		Groups: []metav1.APIGroup{
			{
				Name: "cluster.x-k8s.io",
				PreferredVersion: metav1.GroupVersionForDiscovery{
					GroupVersion: "cluster.x-k8s.io/v1beta1",
					Version:      "v1beta1",
				},
				Versions: []metav1.GroupVersionForDiscovery{
					{
						GroupVersion: "cluster.x-k8s.io/v1beta1",
						Version:      "v1beta1",
					},
				},
			},
			{
				Name: "autoscaling",
				Versions: []metav1.GroupVersionForDiscovery{
					{
						GroupVersion: "autoscaling/v1",
						Version:      "v1",
					},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{
					GroupVersion: "autoscaling/v1",
					Version:      "v1",
				},
			},
		},
	}
}

func GetClusterAPIDescription() metav1.APIResourceList {
	// TODO Maybe delete cluster endpoint
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

func GetAutoscalingApi() metav1.APIResourceList {
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
