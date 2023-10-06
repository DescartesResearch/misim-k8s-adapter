package infrastructure

import (
	"encoding/json"
	"k8s.io/klog/v2"
	"net/http"
	"strings"
)

// If query parameter "watch" is added writes empty
// Writes {"metadata": null, "items": null} to the response
func UnsupportedResource() Endpoint {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.V(7).Infof("Req: %s%s?%s", r.Host, r.URL.Path, r.URL.RawQuery)
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
					klog.V(6).Info("Client stopped listening")
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
				z := map[string]*string{"metadata": nil, "items": nil}
				err = json.NewEncoder(w).Encode(z)
				klog.V(6).ErrorS(err, "unseen type %s\n", resourceType[len(resourceType)-1])
			} else {
				err = json.NewEncoder(w).Encode(y)
			}
			if err != nil {
				klog.V(1).ErrorS(err, "unable to encode empty resource list, error is: %v", err)
				return
			}
		}
	}
}
