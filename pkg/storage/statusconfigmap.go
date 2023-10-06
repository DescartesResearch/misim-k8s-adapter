package storage

import (
	core "k8s.io/api/core/v1"
)

type StatusConfigMapStorage interface {
	// Stores the status config map
	StoreStatusConfigMap(core.ConfigMap)
	// Returns the current status config map
	GetStatusConfigMap() core.ConfigMap
}
