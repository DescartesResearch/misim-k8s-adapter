package infrastructure

import (
	"encoding/json"
	"go-kube/internal/broadcast"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"net/http"
)

func HandleWatchableRequest[T any](supplier func() (T, *broadcast.BroadcastServer[metav1.WatchEvent])) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.V(7).Infof("Req: %s%s?%s", r.Host, r.URL.Path, r.URL.RawQuery)
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

			klog.V(6).Infof("Client started listening (%s)...", r.URL.Path)
			for {
				klog.V(6).Infof("Client waits for result (%s)...", r.URL.Path)
				select {
				case <-ctx.Done():
					klog.V(6).Infof("Client stopped listening (%s)", r.URL.Path)
					return
				case event := <-eventChannel:
					klog.V(6).Infof("Received event for client (%s) of type %s", r.URL.Path, event.Type)
					if err := enc.Encode(event); err != nil {
						klog.V(1).ErrorS(err, "unable to encode watch object %T: %v", event, err)
						// client disconnect.
						return
					}
					if len(eventChannel) == 0 {
						flusher.Flush()
						klog.V(6).Infof("Client flushed (%s)!", r.URL.Path)
						//return
					}
				}
			}
		} else {
			// if no watch we just list the resource
			err := json.NewEncoder(w).Encode(resourceList)
			if err != nil {
				klog.V(1).ErrorS(err, "unable to encode resource list, error is: %v", err)
				return
			}
		}
	}
}
