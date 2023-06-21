package entity

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Update request from the simulation for pods
type PodsUpdateRequest struct {
	AllPods        v1.PodList
	Events         []metav1.WatchEvent
	PodsToBePlaced v1.PodList
}

// Response of the adapter to a PodsUpdateRequest from the simulation
// with the information about bindings and failures from the kubescheduler
type PodsUpdateResponse struct {
	Failed       []BindingFailureInformation
	Binded       []BindingInformation
	NewNodes     []v1.Node
	DeletedNodes []v1.Node
}
