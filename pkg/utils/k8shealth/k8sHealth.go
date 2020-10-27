package k8shealth

import (
	"net/http"

	"k8s.io/klog"
)

// IsReady set the ready probe as ready or not
var IsReady = false

// IsHealth set the health probe as ready or not
var IsHealth = true

// HealthHandlers exposes /health and /ready endpoints for k8s probes
func HealthHandlers() {

	// create a new mux server
	server := http.NewServeMux()
	server.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if IsHealth {
			w.WriteHeader(200)
			w.Write([]byte("ok"))
		} else {
			w.WriteHeader(500)
			w.Write([]byte("dead"))
		}
	})
	server.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if IsReady {
			w.WriteHeader(200)
			w.Write([]byte("ok"))
		} else {
			w.WriteHeader(423)
			w.Write([]byte("not-ready"))
		}
	})

	klog.V(4).Info("k8s probes listenig on port 8080")
	// start an http server using the mux server
	klog.Fatal(http.ListenAndServe(":8080", server))
}
