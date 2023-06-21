package infrastructure

import (
	"encoding/json"
	"fmt"
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1"
	storage "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kube-rise/internal/broadcast"
	"net/http"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
	exp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"strings"
)

type Endpoint func(w http.ResponseWriter, r *http.Request)

func HandleRequest[T any](supplier func() T) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Req: %s %s %s\n", r.Host, r.URL.Path, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		resourceList := supplier()
		json.NewEncoder(w).Encode(resourceList)
	}
}

func DoNothing() Endpoint {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Req: %s %s %s\n", r.Host, r.URL.Path, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
	}
}

func HandleWatchableRequest[T any](supplier func() (T, *broadcast.BroadcastServer[metav1.WatchEvent])) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Req: %s %s %s\n", r.Host, r.URL.Path, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		resourceList, broadcastServer := supplier()
		if r.URL.Query().Get("watch") != "" {
			// watch initiated HTTP streaming answers
			// Sources: https://gist.github.com/vmarmol/b967b29917a34d9307ce
			// https://github.com/kubernetes/kubernetes/blob/828495bcc013b77bb63bcb64111e094e455715bb/staging/src/k8s.io/apiserver/pkg/endpoints/handlers/watch.go#L181
			// https://stackoverflow.com/questions/54890809/how-to-use-request-context-instead-of-closenotifier
			ctx := r.Context()
			flusher, ok := w.(http.Flusher)
			if !ok {
				http.NotFound(w, r)
				return
			}

			// Send the initial headers saying we're gonna stream the response.
			w.Header().Set("Transfer-Encoding", "chunked")
			w.WriteHeader(http.StatusOK)
			flusher.Flush()

			enc := json.NewEncoder(w)

			eventChannel := broadcastServer.Subscribe()
			defer broadcastServer.CancelSubscription(eventChannel)

			for {
				select {
				case <-ctx.Done():
					fmt.Println("Client stopped listening")
					return
				case event := <-eventChannel:
					if err := enc.Encode(event); err != nil {
						fmt.Printf("unable to encode watch object %T: %v", event, err)
						// client disconnect.
						return
					}
					if len(eventChannel) == 0 {
						flusher.Flush()
					}
				}
			}
		} else {
			// if no watch we just list the resource
			err := json.NewEncoder(w).Encode(resourceList)
			if err != nil {
				fmt.Printf("unable to encode resource list, error is: %v", err)
				return
			}
		}
	}
}

func UnsupportedResource() Endpoint {
	return func(w http.ResponseWriter, r *http.Request) {
		// fmt.Printf("Req: %s %s %s\n", r.Host, r.URL.Path, r.URL.RawQuery)
		// w.Header().Set("Retry-After", "9999")
		// w.WriteHeader(410)
		// json.NewEncoder(w).Encode(errors.NewResourceExpired("resource is unsupported by this adapter").ErrStatus)
		fmt.Printf("Req: %s %s %s\n", r.Host, r.URL.Path, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("watch") != "" {
			ctx := r.Context()
			flusher, ok := w.(http.Flusher)
			if !ok {
				http.NotFound(w, r)
				return
			}

			// Send the initial headers saying we're gonna stream the response.
			w.Header().Set("Transfer-Encoding", "chunked")
			w.WriteHeader(http.StatusOK)
			flusher.Flush()

			for {
				select {
				case <-ctx.Done():
					fmt.Println("Client stopped listening")
					return
				}
			}
		} else {
			// if no watch we just list the resource
			// just return nothing here, to *string datatype enables us to use nil
			// y := map[string]*string{"metadata": nil, "items": nil}
			resourceType := strings.Split(r.URL.Path, "/")

			y := GetEmptyResourceList(resourceType[len(resourceType)-1])
			var err error
			if y == nil {
				fmt.Printf("unseen type %s\n", resourceType[len(resourceType)-1])
				z := map[string]*string{"metadata": nil, "items": nil}
				err = json.NewEncoder(w).Encode(z)
			} else {
				err = json.NewEncoder(w).Encode(y)
			}
			if err != nil {
				fmt.Printf("unable to encode empty resource list, error is: %v", err)
				return
			}
		}
	}
}

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
		return &storage.CSIStorageCapacityList{TypeMeta: metav1.TypeMeta{Kind: "CSIStorageCapacityList", APIVersion: "storage.k8s.io/v1"}, Items: nil}
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
