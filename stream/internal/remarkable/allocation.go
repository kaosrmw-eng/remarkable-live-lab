package remarkable

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Locate a uniquely sized mmap allocation, independent of ordering relative to
// DRM mappings. Never scan arbitrary bytes: follow only valid page-sized chunks.
func locateFrameAllocation(maps string, mem io.ReaderAt, frameBytes int64) (int64, error) {
	expected := (frameBytes + 16 + 4095) &^ int64(4095)
	var candidates []int64
	for _, line := range strings.Split(maps, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 5 || fields[1] != "rw-p" {
			continue
		}
		bounds := strings.Split(fields[0], "-")
		if len(bounds) != 2 {
			continue
		}
		start, e1 := strconv.ParseInt(bounds[0], 16, 64)
		end, e2 := strconv.ParseInt(bounds[1], 16, 64)
		if e1 != nil || e2 != nil || start < 0 || end <= start || end-start < expected || end-start > 128<<20 {
			continue
		}
		for off, steps := start, 0; off <= end-16 && steps < 1024; steps++ {
			var header [16]byte
			if _, err := mem.ReadAt(header[:], off); err != nil {
				break
			}
			flags := binary.LittleEndian.Uint64(header[8:])
			size := int64(flags &^ 7)
			if flags&7 != 2 || binary.LittleEndian.Uint64(header[:8]) != 0 || size < 4096 || size%4096 != 0 || size > end-off {
				break
			}
			if size == expected && frameBytes <= size-16 {
				candidates = append(candidates, off+16)
			}
			off += size
		}
	}
	if len(candidates) != 1 {
		return 0, fmt.Errorf("expected one bounded frame allocation, found %d", len(candidates))
	}
	return candidates[0], nil
}
