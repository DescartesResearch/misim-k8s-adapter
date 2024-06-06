package interfaces

import (
	"go-kube/internal/infrastructure"
	"go-kube/pkg/interfaces/kubeapi"
	"go-kube/pkg/interfaces/simulation"
	"go-kube/pkg/storage"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	autoscaling "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/protobuf"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
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
	app.router.HandleFunc("/updateNodes", infrastructure.HandleRequestWithJSONBody(app.sim2.NodeUpdates().Post)).Methods("POST")
	app.router.HandleFunc("/updatePods", infrastructure.HandleRequestWithJSONBody(app.sim2.PodUpdates().Post)).Methods("POST")
	app.router.HandleFunc("/getEventsApiEvents", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		eventList := app.sim2.Events().GetEventsApiEvents()
		eventsv1GV := eventsv1.SchemeGroupVersion
		eventsV1Codec := scheme.Codecs.CodecForVersions(scheme.Codecs.LegacyCodec(eventsv1GV), scheme.Codecs.UniversalDecoder(eventsv1GV), eventsv1GV, eventsv1GV)
		encodedEventList, err := runtime.Encode(eventsV1Codec, &eventList)
		if err != nil {
			klog.V(1).ErrorS(err, "There was an error encoding the events. err = ", err)
			w.WriteHeader(500)
			return
		}
		w.Write(encodedEventList)
	}).Methods("GET")

	app.router.HandleFunc("/getCoreApiEvents", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		eventList := app.sim2.Events().GetCoreApiEvents()
		corev1GV := v1.SchemeGroupVersion
		coreV1Codec := scheme.Codecs.CodecForVersions(scheme.Codecs.LegacyCodec(corev1GV), scheme.Codecs.UniversalDecoder(corev1GV), corev1GV, corev1GV)
		encodedEventList, err := runtime.Encode(coreV1Codec, &eventList)
		if err != nil {
			klog.V(1).ErrorS(err, "There was an error encoding the events. err = ", err)
			w.WriteHeader(500)
			return
		}
		w.Write(encodedEventList)
	}).Methods("GET")
	// Kubeserver API
	app.router.HandleFunc("/api", infrastructure.HandleJSONRequest(app.kube2.Api().Get)).Methods("GET")
	app.router.HandleFunc("/api/v1", infrastructure.HandleJSONRequest(app.kube2.Api().V1().Get)).Methods("GET")
	// app.router.HandleFunc("/api/v1/namespaces/kube-system/configmaps", infrastructure.DoNothing()).Methods("POST")
	app.router.HandleFunc("/api/v1/pods", infrastructure.HandleWatchableRequest(app.kube2.Api().V1().Pods().Get)).Methods("GET")
	app.router.HandleFunc("/api/v1/nodes", infrastructure.HandleWatchableRequest(app.kube2.Api().V1().Nodes().Get)).Methods("GET")
	app.router.HandleFunc("/api/v1/nodes/{nodeName}", infrastructure.HandleRequestWithParams(func(params map[string]string) v1.Node {
		return app.kube2.Api().V1().Nodes().Node(params["nodeName"]).Get()
	})).Methods("GET")
	app.router.HandleFunc("/api/v1/nodes/{nodeName}", infrastructure.HandleRequestWithParamsAndJSONBody(func(params map[string]string, body v1.Node) v1.Node {
		return app.kube2.Api().V1().Nodes().Node(params["nodeName"]).Put(body)
	})).Methods("PUT")
	app.router.HandleFunc("/api/v1/namespaces", infrastructure.HandleWatchableRequest(app.kube2.Api().V1().Namespaces().Get)).Methods("GET")
	app.router.HandleFunc("/api/v1/namespaces/default/pods/{podName}/status", infrastructure.HandleRequestWithParamsAndJSONBody(func(params map[string]string, body v1.PodStatusResult) v1.Pod {
		return app.kube2.Api().V1().Namespaces().Namespace("default").Pods().Pod(params["podName"]).Status().Patch(body)
	})).Methods("PATCH")
	app.router.HandleFunc("/api/v1/namespaces/default/pods/{podName}/binding", func(w http.ResponseWriter, r *http.Request) {
		klog.V(7).Infof("Req: %s%s?%s", r.Host, r.URL.Path, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		reqBody, _ := io.ReadAll(r.Body)
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
		pathParams := mux.Vars(r)
		podName := pathParams["podName"]

		// We always asume this means it's binded
		app.kube2.Api().V1().Namespaces().Namespace("default").Pods().Pod(podName).Binding().Post(*u)
	}).Methods("POST")

	app.router.HandleFunc("/api/v1/namespaces/kube-system/configmaps/cluster-autoscaler-status", infrastructure.HandleJSONRequest(app.kube2.Api().V1().Namespaces().Namespace("kube-system").Configmaps().ClusterAutoscalerStatus().Get)).Methods("GET")
	app.router.HandleFunc("/api/v1/namespaces/kube-system/configmaps/cluster-autoscaler-status", infrastructure.HandleRequestWithJSONBody(app.kube2.Api().V1().Namespaces().Namespace("kube-system").Configmaps().ClusterAutoscalerStatus().Put)).Methods("PUT")
	app.router.HandleFunc("/api/v1/services", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/api/v1/persistentvolumes", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/api/v1/persistentvolumeclaims", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/api/v1/replicationcontrollers", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis", infrastructure.HandleJSONRequest(app.kube2.Apis().Get)).Methods("GET")
	app.router.HandleFunc("/apis/apps", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/apps/v1/daemonsets", infrastructure.HandleWatchableRequest(app.kube2.Apis().Apps().V1().DaemonSets().Get)).Methods("GET")
	app.router.HandleFunc("/apis/apps/v1/replicasets", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/apps/v1/statefulsets", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/autoscaling/v1", infrastructure.HandleJSONRequest(app.kube2.Apis().Autoscaling().V1().Get)).Methods("GET")
	app.router.HandleFunc("/apis/batch/v1/jobs", infrastructure.UnsupportedResource()).Methods("GET")

	app.router.HandleFunc("/api/v1/namespaces/{namespace}/events", infrastructure.HandleRequestWithParamsAndJSONBody(
		func(params map[string]string, body v1.Event) v1.Event {
			return app.kube2.Api().V1().Namespaces().Namespace(params["namespace"]).Events().Post(body)
		})).Methods("POST")

	app.router.HandleFunc("/apis/events.k8s.io/v1", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/events.k8s.io/v1/namespaces/{namespace}/events", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/events.k8s.io/v1/namespaces/{namespace}/events", func(w http.ResponseWriter, r *http.Request) {
		klog.V(7).Infof("Req: %s%s?%s", r.Host, r.URL.Path, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		reqBody, _ := io.ReadAll(r.Body)
		u := &eventsv1.Event{}
		err := runtime.DecodeInto(scheme.Codecs.UniversalDecoder(), reqBody, u)
		if err != nil {
			klog.V(1).ErrorS(err, "There was an error decoding the protobuf. err = ", err)
			w.WriteHeader(500)
			return
		}
		pathParams := mux.Vars(r)
		namespace := pathParams["namespace"]
		klog.V(7).Infof("Received events API event: %v", u)
		app.kube2.Apis().Events().V1().Namespaces().Namespace(namespace).Events().Post(*u)
		w.Write([]byte("{}"))
	}).Methods("POST", "PUT", "PATCH")

	app.router.HandleFunc("/apis/policy/v1/poddisruptionbudgets", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/storage.k8s.io/v1/storageclasses", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/storage.k8s.io/v1/csidrivers", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/storage.k8s.io/v1/csinodes", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/storage.k8s.io/v1/csistoragecapacities", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/storage.k8s.io/v1beta1/csistoragecapacities", infrastructure.UnsupportedResource()).Methods("GET")
	// Clusterx API
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1", infrastructure.HandleJSONRequest(app.kube2.Apis().Cluster().V1Beta1().Get)).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/clusters", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/machines", infrastructure.HandleWatchableRequest(app.kube2.Apis().Cluster().V1Beta1().Machines().Get)).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/machinesets", infrastructure.HandleWatchableRequest(app.kube2.Apis().Cluster().V1Beta1().MachineSets().Get)).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/machinedeployments", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/machinepools", infrastructure.UnsupportedResource()).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/namespaces/{namespace}/machines/{machineName}", infrastructure.HandleRequestWithParams(func(params map[string]string) cluster.Machine {
		return app.kube2.Apis().Cluster().V1Beta1().Namespaces().Namespace(params["namespace"]).Machines().Machine(params["machineName"]).Get()
	})).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/namespaces/{namespace}/machines/{machineName}", infrastructure.HandleRequestWithParamsAndJSONBody(func(params map[string]string, body cluster.Machine) cluster.Machine {
		return app.kube2.Apis().Cluster().V1Beta1().Namespaces().Namespace(params["namespace"]).Machines().Machine(params["machineName"]).Put(body)
	})).Methods("PUT")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/namespaces/{namespace}/machinesets/{machinesetName}/scale", infrastructure.HandleRequestWithParams(func(params map[string]string) autoscaling.Scale {
		return app.kube2.Apis().Cluster().V1Beta1().Namespaces().Namespace(params["namespace"]).MachineSets().MachineSet(params["machinesetName"]).Scale().Get()
	})).Methods("GET")
	app.router.HandleFunc("/apis/cluster.x-k8s.io/v1beta1/namespaces/{namespace}/machinesets/{machinesetName}/scale", infrastructure.HandleRequestWithParamsAndJSONBody(func(params map[string]string, body autoscaling.Scale) autoscaling.Scale {
		return app.kube2.Apis().Cluster().V1Beta1().Namespaces().Namespace(params["namespace"]).MachineSets().MachineSet(params["machinesetName"]).Scale().Put(body)
	})).Methods("PUT")

}
