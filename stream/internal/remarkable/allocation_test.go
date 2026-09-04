package remarkable

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestAllocationMovedAndPayloadOffset(t *testing.T) {
	mem := make([]byte, 0x6000)
	binary.LittleEndian.PutUint64(mem[0x1008:], 0x1002)
	binary.LittleEndian.PutUint64(mem[0x2008:], 0x3002)
	maps := "1000-5000 rw-p 00000000 00:00 0\n8000-9000 rw-s 00000000 00:06 109 /dev/dri/card0\n"
	ptr, err := locateFrameAllocation(maps, bytes.NewReader(mem), 0x2000)
	if err != nil || ptr != 0x2010 {
		t.Fatalf("%x %v", ptr, err)
	}
	binary.LittleEndian.PutUint64(mem[0x2008:], 0x9002)
	if _, err = locateFrameAllocation(maps, bytes.NewReader(mem), 0x2000); err == nil {
		t.Fatal("accepted out of bounds")
	}
}
func TestAllocationAmbiguityRejected(t *testing.T) {
	mem := make([]byte, 0x6000)
	binary.LittleEndian.PutUint64(mem[8:], 0x3002)
	binary.LittleEndian.PutUint64(mem[0x3008:], 0x3002)
	if _, err := locateFrameAllocation("0-6000 rw-p 00000000 00:00 0", bytes.NewReader(mem), 0x2000); err == nil {
		t.Fatal("ambiguous buffer accepted")
	}
}
