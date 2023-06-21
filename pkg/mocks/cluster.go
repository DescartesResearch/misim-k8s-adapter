package mocks

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

func GetClusters() cluster.ClusterList {
	return cluster.ClusterList{TypeMeta: metav1.TypeMeta{Kind: "ClusterList", APIVersion: "cluster.x-k8s.io/v1beta1"}, Items: nil}
}
