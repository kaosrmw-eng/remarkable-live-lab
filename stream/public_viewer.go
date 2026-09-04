package main

import (
	"encoding/json"
	"github.com/owulveryck/goMarkableStream/internal/pubsub"
	"github.com/owulveryck/goMarkableStream/internal/stream"
	"net/http"
)

// Separate read-only server: never forward to the private mux.
func publicViewer(bus *pubsub.PubSub) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		files := map[string]string{"/": "client/public-viewer.html", "/public-viewer.js": "client/public-viewer.js", "/worker_stream_processing.js": "client/worker_stream_processing.js", "/lib/fzstd.min.js": "client/lib/fzstd.min.js"}
		path, ok := files[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		b, e := assetsFS.ReadFile(path)
		if e != nil {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else {
			w.Header().Set("Content-Type", "application/javascript")
		}
		w.Write(b)
	})
	mux.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		checkSuspend()
		on, _ := sharing.state()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"sharing": on, "rotation": sharing.rotationValue(), "width": 960, "height": 1696})
	})
	mux.Handle("/stream", sharing.wrap(stream.NewBroadcast(file, pointerAddr, c.DeltaThreshold)))
	mux.Handle("/screenshot", sharing.wrap(stream.NewScreenshotHandler(file, pointerAddr)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, "Read-only viewer", 405)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
func controlPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "frame-ancestors 'none'")
	b, e := assetsFS.ReadFile("client/private-control.html")
	if e != nil {
		http.Error(w, "Unavailable", 500)
		return
	}
	w.Write(b)
}
