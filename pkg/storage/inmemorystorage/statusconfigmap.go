package inmemorystorage

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatusConfigMapInMemoryStorage struct {
	statusConfigMap core.ConfigMap
}

func (s *StatusConfigMapInMemoryStorage) GetStatusConfigMap() core.ConfigMap {
	return s.statusConfigMap
}

func (s *StatusConfigMapInMemoryStorage) StoreStatusConfigMap(configMap core.ConfigMap) {
	s.statusConfigMap = configMap
}

func NewStatusMapInMemoryStorage() StatusConfigMapInMemoryStorage {
	return StatusConfigMapInMemoryStorage{
		statusConfigMap: core.ConfigMap{TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{Name: "cluster-autoscaler-status", Namespace: "kube-system"}},
	}
}
