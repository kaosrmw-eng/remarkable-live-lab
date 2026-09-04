//go:build linux

package main

import (
	"encoding/binary"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"sync"
	"time"
)

var clockMu sync.Mutex
var priorSleep int64

func checkSuspend() {
	var boot, mono unix.Timespec
	if unix.ClockGettime(unix.CLOCK_BOOTTIME, &boot) != nil || unix.ClockGettime(unix.CLOCK_MONOTONIC, &mono) != nil {
		sharing.fail("Sleep guard unavailable")
		return
	}
	slept := boot.Nano() - mono.Nano()
	clockMu.Lock()
	changed := priorSleep != 0 && slept-priorSleep > int64(20*time.Millisecond)
	priorSleep = slept
	clockMu.Unlock()
	if changed {
		sharing.stop("Stopped after sleep")
	}
}
func startSharingGuards() {
	checkSuspend()
	// Suspend is checked before every capture, frame write and control/state
	// request. No periodic idle wakeup is necessary. Input reads below block.
	for _, path := range []string{"/dev/input/event0", "/dev/input/event1"} {
		go func(path string) {
			f, e := os.Open(path)
			if e != nil {
				sharing.fail("Power guard unavailable")
				return
			}
			defer f.Close()
			b := make([]byte, 24) // ARM64 Linux input_event, no exclusive grab
			for {
				if _, e = io.ReadFull(f, b); e != nil {
					sharing.fail("Power guard disconnected")
					return
				}
				typ := binary.LittleEndian.Uint16(b[16:18])
				code := binary.LittleEndian.Uint16(b[18:20])
				value := binary.LittleEndian.Uint32(b[20:24])
				if (typ == 1 && code == 116 && value == 1) || (typ == 5 && value == 1) {
					sharing.stop("Stopped by power button or cover")
				}
			}
		}(path)
	}
}
