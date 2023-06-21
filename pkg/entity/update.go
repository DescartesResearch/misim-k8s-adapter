package entity

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

// Update request from the simulation for nodes
type NodeUpdateRequest struct {
	AllNodes    v1.NodeList
	Events      []metav1.WatchEvent
	MachineSets []cluster.MachineSet
	Machines    []cluster.Machine
}

// Response of the adapter to a NodeUpdateRequest from the simulation
type NodeUpdateResponse struct {
	Data v1.NodeList `json:"Updated NodeList with"`
}
