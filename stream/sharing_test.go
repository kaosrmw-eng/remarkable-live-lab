package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSharingTimerAndStaleTimer(t *testing.T) {
	g := newSharingGate()
	g.startFor(20 * time.Millisecond)
	time.Sleep(60 * time.Millisecond)
	if on, _ := g.state(); on {
		t.Fatal("timer failed")
	}
	g.startFor(30 * time.Millisecond)
	g.stop("test")
	g.startFor(0)
	time.Sleep(70 * time.Millisecond)
	if on, _ := g.state(); !on {
		t.Fatal("old timer stopped new permanent session")
	}
	g.stop("done")
}

func TestSharingDefaultOffAndCancellation(t *testing.T) {
	g := newSharingGate()
	r := httptest.NewRequest("GET", "/stream", nil)
	w := httptest.NewRecorder()
	called := false
	g.wrap(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).ServeHTTP(w, r)
	if called || w.Code != 423 {
		t.Fatal("default capture not blocked")
	}
	g.start()
	entered := make(chan struct{})
	done := make(chan struct{})
	go func() {
		g.wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-r.Context().Done() })).ServeHTTP(httptest.NewRecorder(), r)
		close(done)
	}()
	<-entered
	g.stop("test")
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("stop did not cancel active stream")
	}
	on, _ := g.state()
	if on {
		t.Fatal("still enabled")
	}
}
