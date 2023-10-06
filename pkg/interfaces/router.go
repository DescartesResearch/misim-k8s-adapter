package interfaces

import (
	"github.com/gorilla/mux"
	"go-kube/internal/infrastructure"
	"go-kube/pkg/interfaces/kubeapi"
	"go-kube/pkg/interfaces/simulation"
	"go-kube/pkg/storage"
	"io"
	autoscaling "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/protobuf"
	"k8s.io/klog/v2"
	"net/http"
	cluster "sigs.k8s.io/cluster-api/api/v1beta1"
)

type AdapterApplication struct {
	router *mux.Router
	kube2  kubeapi.KubeApi
	sim2   simulation.SimulationApi
}

func NewAdapterApplication(storageContainer *storage.StorageContainer) *AdapterApplication {
	var router = mux.NewRouter().StrictSlash(true)
	return &AdapterApplication{
		router: router,
		kube2:  kubeapi.NewKubeApi(storageContainer),
		sim2:   simulation.NewSimulationApi(storageContainer),
	}
}

func (app *AdapterApplication) Start() {
	app.registerRoutes()
	var port = "8000"
	klog.V(1).Info("Starting adapter on port ", port)
	err := http.ListenAndServe(":"+port, app.router)
	if err != nil {
		klog.V(1).ErrorS(err, "Error when calling http.ListenAndServe, error is: %v", err)
		return
	}
}

func (app *AdapterApplication) registerRoutes() {
	// Simulator API
	app.router.HandleFunc("/updateNodes", infrastructure.HandleRequestWithBody(app.sim2.NodeUpdates().Post)).Methods("POST")
	app.router.HandleFunc("/updatePods", infrastructure.HandleRequestWithBody(app.sim2.PodUpdates().Post)).Methods("POST")

	// Kubeserver API
	app.router.HandleFunc("/api", infrastructure.HandleRequest(app.kube2.Api().Get)).Methods("GET")
	app.router.HandleFunc("/api/v1", infrastructure.HandleRequest(app.kube2.Api().V1().Get)).Methods("GET")
	// app.router.HandleFunc("/api/v1/namespaces/kube-system/configmaps", infrastructure.DoNothing()).Methods("POST")
	app.router.HandleFunc("/api/v1/pods", infrastructure.HandleWatchableRequest(app.kube2.Api().V1().Pods().Get)).Methods("GET")
	app.router.HandleFunc("/api/v1/nodes", infrastructure.HandleWatchableRequest(app.kube2.Api().V1().Nodes().Get)).Methods("GET")
	app.router.HandleFunc("/api/v1/nodes/{nodeName}", infrastructure.HandleRequestWithParams(func(params map[string]string) v1.Node {
		return app.kube2.Api().V1().Nodes().Node(params["nodeName"]).Get()
	})).Methods("GET")
	app.router.HandleFunc("/api/v1/nodes/{nodeName}", infrastructure.HandleRequestWithParamsAndBody(func(params map[string]string, body v1.Node) v1.Node {
		return app.kube2.Api().V1().Nodes().Node(params["nodeName"]).Put(body)
	})).Methods("PUT")
	app.router.HandleFunc("/api/v1/namespaces", infrastructure.HandleWatchableRequest(app.kube2.Api().V1().Namespaces().Get)).Methods("GET")
	app.router.HandleFunc("/api/v1/namespaces/default/pods/{podName}/status", infrastructure.HandleRequestWithParamsAndBody(func(params map[string]string, body v1.PodStatusResult) v1.Pod {
		return app.kube2.Api().V1().Namespaces().Namespace("default").Pods().Pod(params["podName"]).Status().Patch(body)
	})).Methods("PATCH")

	// Special protobuf behavior for binding
	app.router.HandleFunc("/api/v1/namespaces/default/pods/{podName}/binding", func(w http.ResponseWriter, r *http.Request) {
		klog.V(7).Infof("Req: %s%s?%s", r.Host, r.URL.Path, r.URL.RawQuery)
		// Loop over header names
		w.Header().Set("Content-Type", "application/json")

		reqBody, _ := io.ReadAll(r.Body)
		// https://github.com/kubernetes/kubernetes/blob/61d455ed1173cd89a98442adf4623a29c5681c58/staging/src/k8s.io/apimachinery/pkg/test/runtime_serializer_protobuf_protobuf_test.go#L87
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(schema.GroupVersion{Version: "v1"}, &v1.Binding{})
		serializer := protobuf.NewSerializer(scheme, scheme)
		u := &v1.Binding{}
		err := runtime.DecodeInto(serializer, reqBody, u)
		if err != nil {
			klog.V(1).ErrorS(err, "There was an error decoding the protobuf. err = ", err)
			w.WriteHeader(500)
			return
		}

		// Get pod name as path parameter
		pathParams := mux.Vars(r)
		podName := pathParams["podName"]

		// We always asume this means it's binded
		app.kube2.Api().V1().Namespaces().Namespace("default").Pods().Pod(podName).Binding().Post(*u)
	}).Methods("POST")

	app.router.HandleFunc("/api/v1/namespaces/kube-system/configmaps/cluster-autoscaler-status", infrastructure.HandleRequest(app.kube2.Api().V1().Namespaces().Namespace("kube-system").Configmaps().ClusterAutoscalerStatus().Get)).Methods("GET")
	app.router.HandleFunc("/api/v1/namespaces/kube-system/configmaps/cluster-autoscaler-status", infrastructure.HandleRequestWithBody(app.kube2.Api().V1().Namespaces().Namespace("kube-system").Configmaps().ClusterAutoscalerStatus().Put)).Methods("PUT")
	app.router.HandleFunc("/api/v1/services", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/api/v1/persistentvolumes", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/api/v1/persistentvolumeclaims", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/api/v1/replicationcontrollers", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis", infrastructure.HandleRequest(app.kube2.Apis().Get)).Methods("GET")
	app.router.HandleFunc("/apis/apps", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/apps/v1/daemonsets", infrastructure.HandleWatchableRequest(app.kube2.Apis().Apps().V1().DaemonSets().Get)).Methods("GET")
	app.router.HandleFunc("/apis/apps/v1/replicasets", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/apps/v1/statefulsets", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/autoscaling/v1", infrastructure.HandleRequest(app.kube2.Apis().Autoscaling().V1().Get)).Methods("GET")
	app.router.HandleFunc("/apis/batch/v1/jobs", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/events.k8s.io/v1", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/events.k8s.io/v1/namespaces/default/events", infrastructure.UnsupportedResource()).Methods("GET", "POST")
	app.router.HandleFunc("/apis/policy/v1/poddisruptionbudgets", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/storage.k8s.io/v1/storageclasses", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/storage.k8s.io/v1/csidrivers", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/storage.k8s.io/v1/csinodes", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/storage.k8s.io/v1/csistoragecapacities", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/storage.k8s.io/v1beta1/csistoragecapacities", infrastructure.UnsupportedResource()).Methods("GET")
	// Clusterx API
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1", infrastructure.HandleRequest(app.kube2.Apis().Cluster().V1Beta1().Get)).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/clusters", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/machines", infrastructure.HandleWatchableRequest(app.kube2.Apis().Cluster().V1Beta1().Machines().Get)).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/machinesets", infrastructure.HandleWatchableRequest(app.kube2.Apis().Cluster().V1Beta1().MachineSets().Get)).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/machinedeployments", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/machinepools", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/namespaces/{namespace}/machines/{machineName}", infrastructure.HandleRequestWithParams(func(params map[string]string) cluster.Machine {
		return app.kube2.Apis().Cluster().V1Beta1().Namespaces().Namespace(params["namespace"]).Machines().Machine(params["machineName"]).Get()
	})).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/namespaces/{namespace}/machines/{machineName}", infrastructure.HandleRequestWithParamsAndBody(func(params map[string]string, body cluster.Machine) cluster.Machine {
		return app.kube2.Apis().Cluster().V1Beta1().Namespaces().Namespace(params["namespace"]).Machines().Machine(params["machineName"]).Put(body)
	})).Methods("PUT")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/namespaces/{namespace}/machinesets/{machinesetName}/scale", infrastructure.HandleRequestWithParams(func(params map[string]string) autoscaling.Scale {
		return app.kube2.Apis().Cluster().V1Beta1().Namespaces().Namespace(params["namespace"]).MachineSets().MachineSet(params["machinesetName"]).Scale().Get()
	})).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/namespaces/{namespace}/machinesets/{machinesetName}/scale", infrastructure.HandleRequestWithParamsAndBody(func(params map[string]string, body autoscaling.Scale) autoscaling.Scale {
		return app.kube2.Apis().Cluster().V1Beta1().Namespaces().Namespace(params["namespace"]).MachineSets().MachineSet(params["machinesetName"]).Scale().Put(body)
	})).Methods("PUT")

}
