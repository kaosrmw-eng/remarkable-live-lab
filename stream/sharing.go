package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
)

// Capture is opt-in and never survives a process restart.
type sharingGate struct {
	mu         sync.Mutex
	enabled    bool
	fault      bool
	rotation   int
	reason     string
	generation context.Context
	cancel     context.CancelFunc
	deadline   time.Time
}

var sharing = newSharingGate()

func (g *sharingGate) rotationValue() int { g.mu.Lock(); defer g.mu.Unlock(); return g.rotation }

func newSharingGate() *sharingGate {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return &sharingGate{reason: "Not sharing", generation: ctx, cancel: cancel}
}
func (g *sharingGate) stop(reason string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.enabled = false
	g.reason = reason
	g.cancel()
}
func (g *sharingGate) fail(reason string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.fault = true
	g.enabled = false
	g.reason = reason
	g.cancel()
}
func (g *sharingGate) start() {
	g.startFor(0)
}
func (g *sharingGate) startFor(duration time.Duration) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.enabled || g.fault {
		return
	}
	g.generation, g.cancel = context.WithCancel(context.Background())
	g.enabled = true
	g.reason = "Sharing publicly"
	g.deadline = time.Time{}
	if duration > 0 {
		g.deadline = time.Now().Add(duration)
		session := g.generation
		go func() {
			timer := time.NewTimer(duration)
			defer timer.Stop()
			select {
			case <-session.Done():
				return
			case <-timer.C:
			}
			g.mu.Lock()
			defer g.mu.Unlock()
			if g.generation == session {
				g.enabled = false
				g.reason = "Sharing timer expired"
				g.cancel()
			}
		}()
	}
}
func (g *sharingGate) state() (bool, string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.enabled, g.reason
}
func (g *sharingGate) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.mu.Lock()
		on, session := g.enabled, g.generation
		g.mu.Unlock()
		if !on {
			http.Error(w, "Not sharing", http.StatusLocked)
			return
		}
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		go func() {
			select {
			case <-session.Done():
				cancel()
			case <-ctx.Done():
			}
		}()
		next.ServeHTTP(guardedResponse{ResponseWriter: w, session: session}, r.WithContext(ctx))
	})
}

type guardedResponse struct {
	http.ResponseWriter
	session context.Context
}

func (w guardedResponse) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w guardedResponse) Write(p []byte) (int, error) {
	checkSuspend()
	if w.session.Err() != nil {
		return 0, errors.New("sharing stopped")
	}
	return w.ResponseWriter.Write(p)
}
func (w guardedResponse) Flush() {
	if w.session.Err() == nil {
		if f, ok := w.ResponseWriter.(http.Flusher); ok {
			f.Flush()
		}
	}
}

type guardedCapture struct{ source io.ReaderAt }

func (g guardedCapture) ReadAt(p []byte, off int64) (int, error) {
	checkSuspend()
	sharing.mu.Lock()
	if !sharing.enabled {
		sharing.mu.Unlock()
		return 0, errors.New("capture stopped")
	}
	n, err := g.source.ReadAt(p, off)
	sharing.mu.Unlock()
	if err != nil {
		sharing.stop("Capture unavailable; try starting again")
	}
	return n, err
}
func sharingControl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodPost {
		// Only same-origin browser commands; auth middleware still requires JWT.
		if origin := r.Header.Get("Origin"); origin != "" && origin != "https://"+r.Host && origin != "http://"+r.Host {
			http.Error(w, "Origin rejected", 403)
			return
		}
		var command struct {
			Action    string `json:"action"`
			Rotation  int    `json:"rotation"`
			Minutes   int    `json:"minutes"`
			Permanent bool   `json:"permanent"`
		}
		if json.NewDecoder(io.LimitReader(r.Body, 1024)).Decode(&command) != nil {
			http.Error(w, "Invalid command", 400)
			return
		}
		switch command.Action {
		case "start":
			if !command.Permanent && (command.Minutes < 5 || command.Minutes > 60 || command.Minutes%5 != 0) {
				http.Error(w, "Choose 5–60 minutes in five-minute steps", 400)
				return
			}
			checkSuspend()
			captureStartMu.Lock()
			defer captureStartMu.Unlock()
			if on, _ := sharing.state(); !on {
				if err := capture.prepare(); err != nil {
					sharing.stop("Capture unavailable — retry Start sharing")
					http.Error(w, "Capture unavailable: "+err.Error(), 503)
					return
				}
			}
			duration := time.Duration(command.Minutes) * time.Minute
			if command.Permanent {
				duration = 0
			}
			sharing.startFor(duration)
		case "stop":
			sharing.stop("Stopped by you")
		case "rotate":
			if command.Rotation != 0 && command.Rotation != 90 && command.Rotation != 180 && command.Rotation != 270 {
				http.Error(w, "Invalid rotation", 400)
				return
			}
			sharing.mu.Lock()
			sharing.rotation = command.Rotation
			sharing.mu.Unlock()
		default:
			http.Error(w, "Invalid action", 400)
			return
		}
	} else if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 405)
		return
	}
	checkSuspend()
	on, reason := sharing.state()
	w.Header().Set("Content-Type", "application/json")
	sharing.mu.Lock()
	var expires any
	if on && !sharing.deadline.IsZero() {
		expires = sharing.deadline.UTC().Format(time.RFC3339)
	}
	sharing.mu.Unlock()
	json.NewEncoder(w).Encode(map[string]any{"sharing": on, "reason": reason, "rotation": sharing.rotationValue(), "expires_at": expires, "capture_error": capture.errorText()})
}
