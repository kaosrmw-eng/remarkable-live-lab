package main

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestCaptureUnavailableThenRecover(t *testing.T) {
	calls := 0
	s := &captureSource{open: func() (io.ReaderAt, int64, error) {
		calls++
		if calls == 1 {
			return nil, 0, errors.New("display not ready")
		}
		return bytes.NewReader([]byte{0, 1, 2, 3}), 1, nil
	}}
	if s.source != nil || calls != 0 {
		t.Fatal("eager initialization")
	}
	if s.prepare() == nil || s.errorText() == "" {
		t.Fatal("missing error")
	}
	if s.prepare() != nil || s.errorText() != "" {
		t.Fatal("retry failed")
	}
	var b [2]byte
	if _, err := s.ReadAt(b[:], 0); err != nil || b[0] != 1 || b[1] != 2 {
		t.Fatal("wrong relative offset")
	}
}
