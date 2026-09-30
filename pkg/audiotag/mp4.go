// SPDX-License-Identifier: Unlicense OR MIT

package audiotag

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"time"
)

// readMP4 reads iTunes metadata: moov/udta/meta/ilst, whose items hold
// their values in data boxes.
func readMP4(r io.ReaderAt, size int64) Info {
	var info Info
	at, n, ok := mp4Path(r, 0, size, "moov", "udta", "meta")
	if !ok {
		return info
	}
	// meta is a full box in MP4 files, a plain one in QuickTime's.
	var probe [8]byte
	if _, err := r.ReadAt(probe[:], at); err != nil {
		return info
	}
	if string(probe[4:8]) != "hdlr" {
		at, n = at+4, n-4
	}
	at, n, ok = mp4Path(r, at, n, "ilst")
	if !ok || n > maxTag {
		return info
	}
	ilst := make([]byte, n)
	if _, err := r.ReadAt(ilst, at); err != nil {
		return info
	}
	eachBox(ilst, func(kind string, item []byte) {
		eachBox(item, func(name string, data []byte) {
			// A data box: its type, its locale, the value.
			if name != "data" || len(data) < 8 {
				return
			}
			value := data[8:]
			switch kind {
			case "\xa9nam":
				info.Title = strings.ToValidUTF8(string(value), "�")
			case "\xa9ART":
				info.Performer = strings.ToValidUTF8(string(value), "�")
			case "covr":
				if info.Cover == nil && len(value) > 0 {
					info.Cover = value
				}
			}
		})
	})
	return info
}

// mp4Path finds the box at the end of a path of box types, among those
// between at and at+size, and returns where its content starts and its
// length.
func mp4Path(r io.ReaderAt, at, size int64, path ...string) (int64, int64, bool) {
	for _, name := range path {
		var ok bool
		if at, size, ok = mp4Box(r, at, size, name); !ok {
			return 0, 0, false
		}
	}
	return at, size, true
}

// mp4Box finds the box of type name among those between at and at+size.
func mp4Box(r io.ReaderAt, at, size int64, name string) (int64, int64, bool) {
	end := at + size
	for at+8 <= end {
		var h [16]byte
		if _, err := r.ReadAt(h[:8], at); err != nil {
			return 0, 0, false
		}
		length, headerSize := int64(binary.BigEndian.Uint32(h[:])), int64(8)
		switch length {
		case 0:
			length = end - at
		case 1:
			if _, err := r.ReadAt(h[8:16], at+8); err != nil {
				return 0, 0, false
			}
			length, headerSize = int64(binary.BigEndian.Uint64(h[8:])), 16
		}
		if length < headerSize || length > end-at {
			return 0, 0, false
		}
		if string(h[4:8]) == name {
			return at + headerSize, length - headerSize, true
		}
		at += length
	}
	return 0, 0, false
}

// eachBox calls fn with the type and the content of each box in b.
func eachBox(b []byte, fn func(kind string, body []byte)) {
	for len(b) >= 8 {
		n := int(binary.BigEndian.Uint32(b))
		if n < 8 || n > len(b) {
			return
		}
		fn(string(b[4:8]), b[8:n])
		b = b[n:]
	}
}

// adtsDuration adds up the frames of an AAC stream in ADTS, past an ID3v2
// tag: each holds 1024 samples a raw block.
func adtsDuration(r io.ReaderAt, size int64) (time.Duration, error) {
	start := int64(0)
	var head [10]byte
	if _, err := r.ReadAt(head[:], 0); err == nil {
		start = id3v2Size(head[:])
	}
	br := bufio.NewReaderSize(io.NewSectionReader(r, start, size-start), 64<<10)
	rates := [...]int64{96000, 88200, 64000, 48000, 44100, 32000, 24000, 22050, 16000, 12000, 11025, 8000, 7350}
	var samples, rate int64
	for {
		var h [7]byte
		if _, err := io.ReadFull(br, h[:]); err != nil {
			break
		}
		if h[0] != 0xff || h[1]&0xf6 != 0xf0 {
			// An ID3v1 tag, or junk, after the frames.
			break
		}
		index := int(h[2] >> 2 & 0x0f)
		length := int(h[3]&3)<<11 | int(h[4])<<3 | int(h[5])>>5
		if index >= len(rates) || length < 7 {
			break
		}
		rate = rates[index]
		samples += 1024 * int64(h[6]&3+1)
		if _, err := br.Discard(length - 7); err != nil {
			break
		}
	}
	if rate == 0 {
		return 0, errors.New("not an AAC file")
	}
	return time.Duration(samples) * time.Second / time.Duration(rate), nil
}
