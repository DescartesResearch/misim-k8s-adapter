package misim

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

// Information about a successfully binded pod
type BindingInformation struct {
	Pod  string
	Node string
}

// Information about a failed binding for a pod
type BindingFailureInformation struct {
	Pod     string
	Message string
}

// Update request from the simulation for nodes
type NodeUpdateRequest struct {
	// All nodes the should be scheduled on the machines
	AllNodes    v1.NodeList
	Events      []metav1.WatchEvent
	MachineSets []cluster.MachineSet
	// Machines available for the nodes
	Machines []cluster.Machine
}

// Response of the adapter to a NodeUpdateRequest from the simulation
type NodeUpdateResponse struct {
	Data v1.NodeList `json:"Updated NodeList with"`
}

// Update request from the simulation for pods
type PodsUpdateRequest struct {
	// All pods in the simulation
	AllPods v1.PodList
	Events  []metav1.WatchEvent
	// Pods that still have to be placed
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
