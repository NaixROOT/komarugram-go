// SPDX-License-Identifier: Unlicense OR MIT

package sendfiles

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

// orientationOfFile reads the turn a JPEG file asks for in its Exif data:
// 1 for none, up to 8; 1 for a file that is not a JPEG or has none.
func orientationOfFile(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 1
	}
	defer f.Close()
	// The Exif segment comes first and is at most 64 KiB.
	head := make([]byte, 70<<10)
	n, _ := io.ReadFull(f, head)
	return orientationOf(head[:n])
}

// orientationOf reads the orientation tag of the Exif segment of a JPEG.
func orientationOf(data []byte) int {
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return 1
	}
	i := 2
	for i+4 <= len(data) {
		if data[i] != 0xff {
			return 1
		}
		marker := data[i+1]
		if marker == 0xff {
			i++
			continue
		}
		// Start of scan or end of image: no Exif after it.
		if marker == 0xda || marker == 0xd9 {
			return 1
		}
		size := int(binary.BigEndian.Uint16(data[i+2:]))
		if size < 2 || i+2+size > len(data) {
			return 1
		}
		segment := data[i+4 : i+2+size]
		if marker == 0xe1 && bytes.HasPrefix(segment, []byte("Exif\x00\x00")) {
			return tiffOrientation(segment[6:])
		}
		i += 2 + size
	}
	return 1
}

// tiffOrientation finds tag 0x0112 in the first directory of Exif's TIFF.
func tiffOrientation(t []byte) int {
	if len(t) < 8 {
		return 1
	}
	var order binary.ByteOrder
	switch string(t[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 1
	}
	if order.Uint16(t[2:]) != 42 {
		return 1
	}
	dir := int(order.Uint32(t[4:]))
	if dir < 8 || dir+2 > len(t) {
		return 1
	}
	count := int(order.Uint16(t[dir:]))
	for k := 0; k < count; k++ {
		entry := dir + 2 + 12*k
		if entry+12 > len(t) {
			return 1
		}
		if order.Uint16(t[entry:]) != 0x0112 {
			continue
		}
		// A SHORT: its value sits in the first two bytes of the value field.
		if v := int(order.Uint16(t[entry+8:])); v >= 1 && v <= 8 {
			return v
		}
		return 1
	}
	return 1
}
