package main

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

var serviceShutdown = make(chan struct{})

// Only registered on the authenticated private mux. A successful process exit
// leaves Restart=on-failure stopped; the boot unit brings it back on restart.
func powerDownHandler(g *sharingGate, shutdown chan struct{}) http.HandlerFunc {
	var once sync.Once
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", 405)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && origin != "https://"+r.Host && origin != "http://"+r.Host {
			http.Error(w, "Origin rejected", 403)
			return
		}
		var command struct {
			Confirm bool `json:"confirm"`
		}
		if json.NewDecoder(io.LimitReader(r.Body, 1024)).Decode(&command) != nil || !command.Confirm {
			http.Error(w, "Explicit confirmation required", 400)
			return
		}
		once.Do(func() {
			g.fail("Sharing service powering down")
			// Allow the acknowledgement to reach the browser before closing listeners.
			time.AfterFunc(500*time.Millisecond, func() { close(shutdown) })
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"shutting_down": true})
	}
}
