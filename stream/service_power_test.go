package main

import (
	"github.com/owulveryck/goMarkableStream/internal/pubsub"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPowerDownRequiresConfirmationAndOrigin(t *testing.T) {
	g := newSharingGate()
	done := make(chan struct{})
	h := powerDownHandler(g, done)
	for _, tc := range []struct {
		method, body, origin string
		code                 int
	}{
		{"GET", "", "", 405}, {"POST", "{}", "", 400}, {"POST", `{"confirm":true}`, "https://untrusted.example", 403},
	} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(tc.method, "https://owner.example/service-power", strings.NewReader(tc.body))
		r.Header.Set("Origin", tc.origin)
		h(w, r)
		if w.Code != tc.code {
			t.Fatalf("got %d expected %d", w.Code, tc.code)
		}
	}
	select {
	case <-done:
		t.Fatal("unauthorized shutdown")
	default:
	}
}
func TestPowerDownLatchesOffAndExitsOnce(t *testing.T) {
	g := newSharingGate()
	g.start()
	done := make(chan struct{})
	h := powerDownHandler(g, done)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		h(w, httptest.NewRequest("POST", "https://owner.example/service-power", strings.NewReader(`{"confirm":true}`)))
		if w.Code != 200 {
			t.Fatal(w.Code)
		}
	}
	g.start()
	if on, _ := g.state(); on {
		t.Fatal("power-down allowed capture to restart")
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown not requested")
	}
}
func TestPublicCannotPowerDown(t *testing.T) {
	h := publicViewer(pubsub.NewPubSub())
	for _, method := range []string{"GET", "POST"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(method, "/service-power", strings.NewReader(`{"confirm":true}`)))
		want := 404
		if method == "POST" {
			want = 405
		}
		if w.Code != want {
			t.Fatal(w.Code)
		}
	}
}
