package apis

import (
	"go-kube/pkg/interfaces/kubeapi/apis/apps"
	"go-kube/pkg/interfaces/kubeapi/apis/autoscaling"
	"go-kube/pkg/interfaces/kubeapi/apis/cluster"
	"go-kube/pkg/interfaces/kubeapi/apis/events"
	"go-kube/pkg/storage"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ApisResource interface {
	Get() meta.APIGroupList
	Apps() apps.AppsResource
	Autoscaling() autoscaling.AutoscalingResource
	Cluster() cluster.ClusterResource
	Events() events.EventsResource
}

type ApisResourceImpl struct {
	storage *storage.StorageContainer
}

func (api ApisResourceImpl) Get() meta.APIGroupList {
	return meta.APIGroupList{
		Groups: []meta.APIGroup{
			{
				Name: "cluster.x-k8s.io",
				PreferredVersion: meta.GroupVersionForDiscovery{
					GroupVersion: "cluster.x-k8s.io/v1beta1",
					Version:      "v1beta1",
				},
				Versions: []meta.GroupVersionForDiscovery{
					{
						GroupVersion: "cluster.x-k8s.io/v1beta1",
						Version:      "v1beta1",
					},
				},
			},
			{
				Name: "autoscaling",
				Versions: []meta.GroupVersionForDiscovery{
					{
						GroupVersion: "autoscaling/v1",
						Version:      "v1",
					},
				},
				PreferredVersion: meta.GroupVersionForDiscovery{
					GroupVersion: "autoscaling/v1",
					Version:      "v1",
				},
			},
		},
	}
}

func (api ApisResourceImpl) Apps() apps.AppsResource {
	return apps.NewAppsResource(api.storage)
}

func (api ApisResourceImpl) Autoscaling() autoscaling.AutoscalingResource {
	return autoscaling.NewAutoscalingResource(api.storage)
}

func (api ApisResourceImpl) Cluster() cluster.ClusterResource {
	return cluster.NewClusterResource(api.storage)
}

func (api ApisResourceImpl) Events() events.EventsResource {
	return events.NewEventsResource(api.storage)
}

func NewApisResource(storage *storage.StorageContainer) ApisResource {
	return ApisResourceImpl{
		storage: storage,
	}
}
