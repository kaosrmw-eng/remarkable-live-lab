package stream

import (
	"bytes"
	"github.com/owulveryck/goMarkableStream/internal/delta"
	"github.com/owulveryck/goMarkableStream/internal/remarkable"
	"io"
	"net/http"
	"sync"
	"time"
)

// Broadcast captures/encodes once, sharing immutable frames with bounded queues.
// A slow viewer disconnects and reconnects with a full frame, never broken deltas.
type Broadcast struct {
	mu        sync.Mutex
	clients   map[chan []byte]bool
	source    io.ReaderAt
	offset    int64
	threshold float64
	reset     bool
	running   bool
}

func NewBroadcast(source io.ReaderAt, offset int64, threshold float64) *Broadcast {
	return &Broadcast{clients: make(map[chan []byte]bool), source: source, offset: offset, threshold: threshold}
}
func (b *Broadcast) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.mu.Lock()
	if len(b.clients) >= 16 {
		b.mu.Unlock()
		http.Error(w, "Viewer capacity reached", 429)
		return
	}
	ch := make(chan []byte, 2)
	b.clients[ch] = true
	b.reset = true
	if !b.running {
		b.running = true
		go b.run()
	}
	b.mu.Unlock()
	defer func() { b.mu.Lock(); delete(b.clients, ch); b.mu.Unlock() }()
	w.Header().Set("Content-Type", "application/octet-stream")
	controller := http.NewResponseController(w)
	for {
		select {
		case <-r.Context().Done():
			return
		case frame, ok := <-ch:
			if !ok {
				return
			}
			controller.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if _, err := w.Write(frame); err != nil {
				return
			}
			if err := controller.Flush(); err != nil {
				return
			}
		}
	}
}
func (b *Broadcast) run() {
	encoder := delta.NewEncoder(b.threshold)
	defer encoder.ReleaseMemory()
	raw := make([]byte, remarkable.Config.SizeBytes)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		b.mu.Lock()
		if len(b.clients) == 0 {
			b.running = false
			b.mu.Unlock()
			return
		}
		if b.reset {
			encoder.Reset()
			b.reset = false
		}
		// Hold registration lock through encoding: a new client must start on a full frame.
		_, err := b.source.ReadAt(raw, b.offset)
		var encoded bytes.Buffer
		if err == nil {
			err = encoder.Encode(raw, &encoded)
		}
		if err != nil {
			for ch := range b.clients {
				close(ch)
				delete(b.clients, ch)
			}
			b.running = false
			b.mu.Unlock()
			return
		}
		if encoded.Len() > 0 {
			for ch := range b.clients {
				select {
				case ch <- encoded.Bytes():
				default:
					close(ch)
					delete(b.clients, ch)
				}
			}
		}
		b.mu.Unlock()
	}
}
