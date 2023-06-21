package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"kube-rise/pkg/control"
	"kube-rise/pkg/server/debugging"
	"kube-rise/pkg/server/infrastructure"
	"kube-rise/pkg/server/kubeapi"
	"kube-rise/pkg/server/simulation"
	"kube-rise/pkg/storage"
	"net/http"
)

type KubeAPI struct {
	core    *kubeapi.CoreKubeAPIServer
	meta    *kubeapi.MetaKubeAPIServer
	apps    *kubeapi.AppsKubeAPIServer
	cluster *kubeapi.ClusterKubeAPIServer
}

type AdapterApplication struct {
	router    *mux.Router
	sim       *simulation.SimServer
	kube      *KubeAPI
	debugging *debugging.DebugServer
}

func NewAdapterApplication(storageContainer *storage.StorageContainer) *AdapterApplication {
	var router = mux.NewRouter().StrictSlash(true)
	var kubeUpdateController = control.NewKubeUpdateController(storageContainer)
	return &AdapterApplication{
		router: router,
		sim:    simulation.NewSimServer(kubeUpdateController),
		kube: &KubeAPI{
			core:    kubeapi.NewCoreKubeAPIServer(storageContainer),
			meta:    kubeapi.NewMetaKubeAPIServer(),
			apps:    kubeapi.NewAppsKubeAPIServer(storageContainer),
			cluster: kubeapi.NewClusterKubeAPIServer(storageContainer),
		},
		debugging: debugging.NewDebugServer(storageContainer),
	}
}

func (app *AdapterApplication) Start() {
	app.registerRoutes()
	var port = "8000"
	fmt.Println("Starting server on port %v", port)
	err := http.ListenAndServe(":"+port, app.router)
	if err != nil {
		fmt.Printf("Error when calling http.ListenAndServe, error is: %v", err)
		return
	}
}

func (app *AdapterApplication) registerRoutes() {
	// Simulator API
	app.router.HandleFunc("/updateNodes", app.sim.UpdateNodes).Methods("POST")
	app.router.HandleFunc("/updatePods", app.sim.UpdatePods).Methods("POST")

	// Kubeserver API
	// Currently with function
	app.router.HandleFunc("/api/v1/namespaces/default/pods/{podName}/status", app.sim.UpdateStatus).Methods("PATCH")
	app.router.HandleFunc("/api/v1/namespaces/default/pods/{podName}/binding", app.sim.UpdateBinding).Methods("POST")
	app.router.HandleFunc("/api/v1/pods", app.kube.core.GetPods).Methods("GET")
	app.router.HandleFunc("/api/v1/nodes", app.kube.core.GetNodes).Methods("GET")
	app.router.HandleFunc("/api/v1/namespaces", app.kube.core.GetNamespaces).Methods("GET")

	// Other Kubernetes API mocks
	app.router.HandleFunc("/apis/apps/v1/replicasets", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/api/v1/persistentvolumes", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/apps/v1/statefulsets", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/storage.k8s.io/v1/storageclasses", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/storage.k8s.io/v1/csidrivers", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/policy/v1/poddisruptionbudgets", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/storage.k8s.io/v1/csinodes", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/api/v1/persistentvolumeclaims", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/storage.k8s.io/v1/csistoragecapacities", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/api/v1/services", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/api/v1/replicationcontrollers", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/events.k8s.io/v1", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/events.k8s.io/v1/namespaces/default/events", infrastructure.UnsupportedResource()).Methods("GET", "POST")
	app.router.HandleFunc("/apis/apps", infrastructure.UnsupportedResource()).Methods("GET")

	// Additional dependencies for cluster-autoscaler
	app.router.HandleFunc("/api", app.kube.meta.GetApi).Methods("GET")
	app.router.HandleFunc("/api/v1", app.kube.meta.GetV1Api).Methods("GET")
	app.router.HandleFunc("/apis/autoscaling/v1", app.kube.meta.GetAutoscalingApi).Methods("GET")
	app.router.HandleFunc("/apis", app.kube.meta.GetApis).Methods("GET")
	// app.router.HandleFunc("/api/v1/namespaces/kube-system/configmaps", infrastructure.DoNothing()).Methods("POST")
	app.router.HandleFunc("/apis/apps/v1/daemonsets", app.kube.apps.GetDaemonSets).Methods("GET")
	app.router.HandleFunc("/apis/batch/v1/jobs", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/api/v1/namespaces/kube-system/configmaps/cluster-autoscaler-status", app.kube.cluster.GetStatusConfigMap).Methods("GET")
	app.router.HandleFunc("/api/v1/namespaces/kube-system/configmaps/cluster-autoscaler-status", app.kube.cluster.PutStatusConfigMap).Methods("PUT")
	app.router.HandleFunc("/api/v1/nodes/{nodeName}", app.kube.core.PutNode).Methods("PUT")
	app.router.HandleFunc("/api/v1/nodes/{nodeName}", app.kube.core.GetNode).Methods("GET")

	// Clusterx API
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1", app.kube.meta.GetClusterAPIDescription).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/clusters", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/machines", app.kube.cluster.GetMachines).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/namespaces/{namespace}/machines/{machineName}", app.kube.cluster.GetMachine).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/namespaces/{namespace}/machines/{machineName}", app.kube.cluster.PutMachine).Methods("PUT")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/machinesets", app.kube.cluster.GetMachineSets).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/namespaces/{namespace}/machinesets/{machinesetName}/scale", app.kube.cluster.GetMachineSetsScale).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/namespaces/{namespace}/machinesets/{machinesetName}/scale", app.kube.cluster.PutMachineSetsScale).Methods("PUT")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/machinepools", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/machinedeployments", infrastructure.UnsupportedResource()).Methods("GET")

	// For debugging
	app.router.HandleFunc("/debugging", app.debugging.InitTestValues)
	app.router.HandleFunc("/debugPods", app.debugging.InitTestPods)
}
