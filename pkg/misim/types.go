// Package misim contains type definitions used for the communication with the simulation.
package misim

import (
	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

// BindingInformation represents the information about a successfully bound pod.
type BindingInformation struct {
	// The name of the pod that was bound.
	Pod string
	// The name of the node the pod was bound to.
	Node string
}

// BindingFailureInformation represents the information about a pod that could
// not be bound to a node.
type BindingFailureInformation struct {
	// The name of the pod that could not be bound.
	Pod string
	// The message specifying the reason why the pod could not be bound.
	Message string
}

// NodeUpdateRequest represents a request sent by the simulation when one or more
// nodes have been updated.
type NodeUpdateRequest struct {
	// All nodes the should be scheduled on the machines
	AllNodes    v1.NodeList
	Events      []metav1.WatchEvent
	MachineSets []cluster.MachineSet
	// Machines available for the nodes
	Machines []cluster.Machine
}

// PodFailureRequest represents a request sent by the simulation when a pod failed.
type PodFailureRequest struct {
	// The name of the failed pod
	FailedPod string `json:"failedPod"`
}

// NodeFailureRequest represents a request sent by the simulation when one or
// more nodes fail.
type NodeFailureRequest struct {
	// FailedPods is the list of pod names that were scheduled on the nodes that failed.
	FailedPods []string `json:"failedPods"`
}

// NodeFailureResponse represents the response to send to the simulation after
// processing a node failure. It includes information about when to mark a node
// as "NotReady" and when to evict pods as a result of the failure.
type NodeFailureResponse struct {
	// NodeMonitorGracePeriodSeconds is the duration (in seconds) after which the
	// failed nodes will be marked as "NotReady".
	NodeMonitorGracePeriodSeconds int `json:"nodeMonitorGracePeriodSeconds"`
	// NoExecuteTaintDelaySeconds represents the delay in seconds after which a NoExecute
	// taint should be added to the node.
	NoExecuteTaintDelaySeconds int `json:"noExecuteTaintDelaySeconds"`
	// PodEvictionEvents maps the delay (in seconds) after a node is marked "NotReady"
	// to the list of pods that should be evicted at that time. Evictions are determined
	// based on the pods' tolerations. An eviction delay of `-1` indicates to never evict the pod.
	PodEvictionEvents map[int64][]string `json:"podEvictionEvents"`
}

// NodeNotReadyRequest represents a request sent by the simulation when one or
// more nodes have been marked as not ready. Currently, this can only happen when the
// node(s) previously failed and the `NodeMonitorGracePeriodSeconds` has elapsed.
type NodeNotReadyRequest struct {
	// Nodes is the list of node names to mark as "NotReady".
	Nodes []string `json:"nodes"`
}

// NodeNotReadyResponse represents the response to send to the simulation after
// marking nodes as "NotReady". It contains the updated Kubernetes representations
// of the affected nodes, machines, and machine sets .
type NodeNotReadyResponse struct {
	// Nodes is the list of updated Kubernetes node representations.
	Nodes []v1.Node `json:"nodes"`
	// Machines is the list of updated Kubernetes machine representations.
	Machines []cluster.Machine `json:"machines"`
	// MachineSets is the list of updated Kubernetes node representations.
	MachineSets []cluster.MachineSet `json:"machineSets"`
}

// NodeNoExecuteRequest represents a request sent by the simulation when one or
// more nodes have been marked as not ready and a NoExecute taint should be added.
// Currently, this can only happen when the node(s) previously failed and the
// nodes have been marked as `NotReady` after `NodeMonitorGracePeriodSeconds` seconds
// and the `NodeMonitorPeriodSeconds` delay has elapsed..
type NodeNoExecuteRequest struct {
	// Nodes is the list of node names to add the "NoExecute" taint to.
	Nodes []string `json:"nodes"`
}

// NodeNoExecuteResponse represent the response to send to the simulation after adding
// the "NoExecute" taints to the requested nodes have been added. It contains the updated
// Kubernetes representations of the affected nodes.
type NodeNoExecuteResponse struct {
	// Nodes is the list of updated Kubernetes node representations.
	Nodes []v1.Node `json:"nodes"`
}

// PodsUpdateRequest represents the request sent by the simulation when one or
// more pods have been updated.
type PodsUpdateRequest struct {
	// All pods in the simulation
	AllPods v1.PodList
	Events  []metav1.WatchEvent
	// Pods that still have to be placed
	PodsToBePlaced v1.PodList
}

// PodsUpdateResponse represents the respond to send to the simulation after
// handling a PodsUpdateRequest. It contains the information about bindings and
// failures from the kubescheduler.
type PodsUpdateResponse struct {
	Failed       []BindingFailureInformation
	Binded       []BindingInformation
	NewNodes     []v1.Node
	DeletedNodes []v1.Node
}

// EventsResponse represents the response to send to the simulation after
// an event request
type EventsResponse struct{ eventsv1.EventList }
