package infrastructure

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"k8s.io/klog/v2"
	"net/http"
)

func HandleRequest[T any](supplier func() T) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.V(7).Infof("Req: %s%s?%s", r.Host, r.URL.Path, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		resourceList := supplier()
		json.NewEncoder(w).Encode(resourceList)
	}
}

func HandleRequestWithBody[B any, T any](supplier func(B) T) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.V(7).Infof("Req: %s%s?%s", r.Host, r.URL.Path, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		reqBody, _ := io.ReadAll(r.Body)
		var payload B
		err := json.Unmarshal(reqBody, &payload)
		if err != nil {
			klog.V(1).ErrorS(err, "There was an error decoding the json. err = %s", err)
			w.WriteHeader(500)
			return
		}
		resourceList := supplier(payload)
		json.NewEncoder(w).Encode(resourceList)
	}
}

func HandleRequestWithParamsAndBody[B any, T any](supplier func(map[string]string, B) T) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.V(7).Infof("Req: %s%s?%s", r.Host, r.URL.Path, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		reqBody, _ := io.ReadAll(r.Body)
		var payload B
		err := json.Unmarshal(reqBody, &payload)
		if err != nil {
			klog.V(1).ErrorS(err, "There was an error decoding the json. err = %s", err)
			w.WriteHeader(500)
			return
		}
		resourceList := supplier(mux.Vars(r), payload)
		json.NewEncoder(w).Encode(resourceList)
	}
}

func HandleRequestWithParams[T any](supplier func(map[string]string) T) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.V(7).Infof("Req: %s%s?%s", r.Host, r.URL.Path, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		resourceList := supplier(mux.Vars(r))
		json.NewEncoder(w).Encode(resourceList)
	}
}
