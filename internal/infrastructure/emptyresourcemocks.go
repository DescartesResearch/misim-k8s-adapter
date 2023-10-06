package infrastructure

import (
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1"
	storage "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
	exp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
)

func GetEmptyResourceList(resourceType string) runtime.Object {
	switch resourceType {
	case "replicasets":
		return &apps.ReplicaSetList{TypeMeta: metav1.TypeMeta{Kind: "ReplicaSetList", APIVersion: "apps/v1"}, Items: nil}
	case "persistentvolumes":
		return &core.PersistentVolumeList{TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeList", APIVersion: "v1"}, Items: nil}
	case "statefulsets":
		return &apps.StatefulSetList{TypeMeta: metav1.TypeMeta{Kind: "StatefulSetList", APIVersion: "apps/v1"}, Items: nil}
	case "storageclasses":
		return &storage.StorageClassList{TypeMeta: metav1.TypeMeta{Kind: "StorageClassList", APIVersion: "storage.k8s.io/v1"}, Items: nil}
	case "csidrivers":
		return &storage.CSIDriverList{TypeMeta: metav1.TypeMeta{Kind: "CSIDriverList", APIVersion: "storage.k8s.io/v1"}, Items: nil}
	case "poddisruptionbudgets":
		return &policy.PodDisruptionBudgetList{TypeMeta: metav1.TypeMeta{Kind: "PodDisruptionBudgetList", APIVersion: "policy/v1"}, Items: nil}
	case "csinodes":
		return &storage.CSINodeList{TypeMeta: metav1.TypeMeta{Kind: "CSINodeList", APIVersion: "storage.k8s.io/v1"}, Items: nil}
	case "persistentvolumeclaims":
		return &core.PersistentVolumeClaimList{TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaimList", APIVersion: "v1"}, Items: nil}
	case "csistoragecapacities":
		return &storage.CSIStorageCapacityList{TypeMeta: metav1.TypeMeta{Kind: "CSIStorageCapacityList", APIVersion: "storage.k8s.io/v1beta1"}, Items: nil}
	case "services":
		return &core.ServiceList{TypeMeta: metav1.TypeMeta{Kind: "ServiceList", APIVersion: "v1"}, Items: nil}
	case "replicationcontrollers":
		return &core.ReplicationControllerList{TypeMeta: metav1.TypeMeta{Kind: "ReplicationControllerList", APIVersion: "v1"}, Items: nil}
	case "jobs":
		return &batch.JobList{TypeMeta: metav1.TypeMeta{Kind: "JobList", APIVersion: "batch/v1"}, Items: nil}
	case "machinedeployments":
		return &cluster.MachineDeploymentList{TypeMeta: metav1.TypeMeta{Kind: "MachineDeploymentList", APIVersion: "cluster.x-k8s.io/v1beta1"}, Items: nil}
	case "machinepools":
		return &exp.MachinePoolList{TypeMeta: metav1.TypeMeta{Kind: "MachinePoolList", APIVersion: "cluster.x-k8s.io/v1beta1"}, Items: nil}
	default:
		return nil
	}
}
