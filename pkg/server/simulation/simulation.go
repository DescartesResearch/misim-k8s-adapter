package simulation

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/protobuf"
	"kube-rise/pkg/control"
	"kube-rise/pkg/entity"
	"net/http"
)

type SimServer struct {
	kubeupdatecontroller control.KubeUpdateController
}

func NewSimServer(kuc control.KubeUpdateController) *SimServer {
	return &SimServer{kubeupdatecontroller: kuc}
}

func (s *SimServer) UpdateNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reqBody, _ := io.ReadAll(r.Body)
	// fmt.Println(string(reqBody))
	var u entity.NodeUpdateRequest
	err := json.Unmarshal(reqBody, &u)
	if err != nil {
		fmt.Printf("There was an error decoding the json. err = %s", err)
		w.WriteHeader(500)
		return
	}

	fmt.Println("Update nodes")
	// fmt.Println(u)

	var response entity.NodeUpdateResponse
	if u.MachineSets == nil || len(u.MachineSets) == 0 {
		// No cluster scaling
		response = s.kubeupdatecontroller.UpdateNodes(u.AllNodes, u.Events)
	} else {
		response = s.kubeupdatecontroller.InitMachinesNodes(u.AllNodes, u.Events, u.MachineSets, u.Machines)
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Printf("Unable to encode response for node update, error is: %v", err)
		return
	}
}

func (s *SimServer) UpdatePods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reqBody, _ := io.ReadAll(r.Body)

	// fmt.Println(string(reqBody))

	var u entity.PodsUpdateRequest
	err := json.Unmarshal(reqBody, &u)
	if err != nil {
		fmt.Printf("There was an error decoding the json. err = %s", err)
		w.WriteHeader(500)
		return
	}

	fmt.Println("Update pods")
	// fmt.Println(u)

	var responseChan = s.kubeupdatecontroller.UpdatePods(u.AllPods, u.Events, u.PodsToBePlaced, false)
	var finalResponse entity.PodsUpdateResponse
	if len(u.PodsToBePlaced.Items) > 0 {
		// Immediately wait for the response from the channel
		select {
		case response := <-responseChan:
			finalResponse = response
			// Should not get stuck with the new scheduling logic
			// case <-time.After(1000 * time.Second):
			// fmt.Println("Scheduler took to long using default response")
			// finalResponse = s.kubeupdatecontroller.CreateDefaultResponse()
		}
	} else {
		finalResponse = s.kubeupdatecontroller.CreateDefaultResponse()
	}
	// s.kubeupdatecontroller.UpdatePods(u.AllPods, u.Events, u.PodsToBePlaced, true)
	err = json.NewEncoder(w).Encode(finalResponse)
	if err != nil {
		fmt.Printf("Unable to encode response for pod update, error is: %v", err)
		return
	}
}

func (s *SimServer) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Req: %s %s %s\n", r.Host, r.URL.Path, r.URL.RawQuery)
	// Loop over header names
	// for name, values := range r.Header {
	// Loop over all values for the name.
	// for _, value := range values {
	// fmt.Println(name, value)
	// }
	// }
	w.Header().Set("Content-Type", "application/json")

	reqBody, _ := io.ReadAll(r.Body)
	// fmt.Println(string(reqBody))
	var u v1.PodStatusResult
	err := json.Unmarshal(reqBody, &u)
	// fmt.Println(u)
	if err != nil {
		fmt.Printf("There was an error decoding the json. err = %s", err)
		w.WriteHeader(500)
		return
	}

	// Get pod name as path parameter
	pathParams := mux.Vars(r)
	podName := pathParams["podName"]

	// We always asume this means it's failed
	finalResponse := s.kubeupdatecontroller.Failed(u.Status, podName)

	err = json.NewEncoder(w).Encode(finalResponse)
	if err != nil {
		fmt.Printf("Unable to encode response for pod status update, error is: %v", err)
		return
	}
}

func (s *SimServer) UpdateBinding(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Req: %s %s %s\n", r.Host, r.URL.Path, r.URL.RawQuery)
	// Loop over header names
	// for name, values := range r.Header {
	// Loop over all values for the name.
	// for _, value := range values {
	// fmt.Println(name, value)
	// }
	// }
	w.Header().Set("Content-Type", "application/json")

	reqBody, _ := io.ReadAll(r.Body)
	// https://github.com/kubernetes/kubernetes/blob/61d455ed1173cd89a98442adf4623a29c5681c58/staging/src/k8s.io/apimachinery/pkg/test/runtime_serializer_protobuf_protobuf_test.go#L87
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(schema.GroupVersion{Version: "v1"}, &v1.Binding{})
	serializer := protobuf.NewSerializer(scheme, scheme)
	u := &v1.Binding{}
	err := runtime.DecodeInto(serializer, reqBody, u)
	if err != nil {
		fmt.Printf("There was an error decoding the protobuf. err = %s", err)
		w.WriteHeader(500)
		return
	}

	// Get pod name as path parameter
	pathParams := mux.Vars(r)
	podName := pathParams["podName"]

	// We always asume this means it's binded
	s.kubeupdatecontroller.Binded(*u, podName)
}
