package main

import (
	"github.com/owulveryck/goMarkableStream/internal/pubsub"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublicViewerCannotControl(t *testing.T) {
	sharing.stop("test")
	handler := publicViewer(pubsub.NewPubSub())
	for _, path := range []string{"/sharing", "/control", "/login", "/password", "/funnel", "/debug/pprof/", "/private-control.html", "/auth.js"} {
		for _, method := range []string{"GET", "POST"} {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(method, path, nil)
			r.Header.Set("Authorization", "Bearer fabricated")
			handler.ServeHTTP(w, r)
			want := 404
			if method == "POST" {
				want = 405
			}
			if w.Code != want {
				t.Errorf("%s %s got %d want %d", method, path, w.Code, want)
			}
		}
	}
	for _, path := range []string{"/stream", "/screenshot"} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != 423 {
			t.Errorf("stopped %s: %d", path, w.Code)
		}
	}
	for _, path := range []string{"/", "/state", "/public-viewer.js", "/worker_stream_processing.js", "/lib/fzstd.min.js"} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 {
			t.Errorf("public asset %s: %d", path, w.Code)
		}
	}
}
