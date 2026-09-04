package stream

import (
	"bytes"
	"github.com/owulveryck/goMarkableStream/internal/remarkable"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBroadcastEightViewers(t *testing.T) {
	b := NewBroadcast(bytes.NewReader(make([]byte, remarkable.Config.SizeBytes)), 0, .5)
	server := httptest.NewServer(b)
	defer server.Close()
	var responses []*http.Response
	defer func() {
		for _, r := range responses {
			r.Body.Close()
		}
	}()
	for i := 0; i < 8; i++ {
		r, err := http.Get(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		responses = append(responses, r)
		if r.StatusCode != 200 {
			t.Fatalf("viewer %d: %d", i, r.StatusCode)
		}
		if _, err = io.ReadFull(r.Body, make([]byte, 8)); err != nil {
			t.Fatal(err)
		}
	}
}
