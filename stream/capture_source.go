package main

import (
	"fmt"
	"github.com/owulveryck/goMarkableStream/internal/remarkable"
	"io"
	"sync"
)

// The network service has no startup dependency on the display process.
type captureSource struct {
	mu        sync.Mutex
	source    io.ReaderAt
	base      int64
	lastError string
	open      func() (io.ReaderAt, int64, error)
}

var capture = &captureSource{open: remarkable.GetFileAndPointer}
var captureStartMu sync.Mutex

func (s *captureSource) prepare() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if closer, ok := s.source.(io.Closer); ok {
		closer.Close()
	}
	s.source = nil
	reader, base, err := s.open()
	if err != nil {
		s.lastError = err.Error()
		return err
	}
	s.source = reader
	s.base = base
	s.lastError = ""
	return nil
}
func (s *captureSource) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.source == nil {
		return 0, fmt.Errorf("capture unavailable")
	}
	n, err := s.source.ReadAt(p, s.base+off)
	if err != nil {
		s.lastError = err.Error()
	}
	return n, err
}
func (s *captureSource) errorText() string { s.mu.Lock(); defer s.mu.Unlock(); return s.lastError }
