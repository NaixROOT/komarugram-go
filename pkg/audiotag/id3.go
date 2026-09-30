// SPDX-License-Identifier: Unlicense OR MIT

package audiotag

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
	"unicode/utf16"
)

// readID3 reads the ID3v2 tag at the start of a file, and the ID3v1 tag at
// its end for what the first does not say.
func readID3(r io.ReaderAt, size int64) Info {
	var info Info
	if tag := id3v2At(r, size, 0); tag != nil {
		info = parseID3v2(tag)
	}
	if info.Title == "" && info.Performer == "" && size >= 128 {
		tail := make([]byte, 128)
		if _, err := r.ReadAt(tail, size-128); err == nil && string(tail[:3]) == "TAG" {
			info.Title = latin1(trimNull(tail[3:33]))
			info.Performer = latin1(trimNull(tail[33:63]))
		}
	}
	return info
}

// id3v2Size is the length of the ID3v2 tag at the start of head, header
// and footer included, or 0 when there is none.
func id3v2Size(head []byte) int64 {
	if len(head) < 10 || string(head[:3]) != "ID3" || head[3] < 2 || head[3] > 4 {
		return 0
	}
	n := 10 + int64(synchsafe(head[6:10]))
	if head[5]&0x10 != 0 {
		n += 10
	}
	return n
}

// id3v2At reads the whole ID3v2 tag at offset at, nil when there is none
// or it is too large.
func id3v2At(r io.ReaderAt, size, at int64) []byte {
	var head [10]byte
	if _, err := r.ReadAt(head[:], at); err != nil {
		return nil
	}
	n := id3v2Size(head[:])
	if n == 0 || n > maxTag || at+n > size {
		return nil
	}
	tag := make([]byte, n)
	if _, err := r.ReadAt(tag, at); err != nil {
		return nil
	}
	return tag
}

func synchsafe(b []byte) uint32 {
	return uint32(b[0]&0x7f)<<21 | uint32(b[1]&0x7f)<<14 | uint32(b[2]&0x7f)<<7 | uint32(b[3]&0x7f)
}

// parseID3v2 reads the title, the performer and the front cover of an
// ID3v2.2, 2.3 or 2.4 tag.
func parseID3v2(tag []byte) Info {
	version, flags := tag[3], tag[5]
	body := tag[10 : 10+synchsafe(tag[6:10])]
	if version < 4 && flags&0x80 != 0 {
		// Before 2.4, unsynchronisation is of the whole tag.
		body = resync(body)
	}
	if flags&0x40 != 0 && version >= 3 && len(body) >= 4 {
		// The extended header: its size leaves itself out in 2.3.
		n := int(binary.BigEndian.Uint32(body))
		if version == 3 {
			n += 4
		} else {
			n = int(synchsafe(body))
		}
		if n > len(body) {
			return Info{}
		}
		body = body[n:]
	}
	idSize, headerSize := 4, 10
	if version == 2 {
		idSize, headerSize = 3, 6
	}
	var info Info
	coverType := -1
	for len(body) >= headerSize && body[0] != 0 {
		id := string(body[:idSize])
		var n int
		var formatFlags byte
		switch version {
		case 2:
			n = int(body[3])<<16 | int(body[4])<<8 | int(body[5])
		case 3:
			n = int(binary.BigEndian.Uint32(body[4:]))
			formatFlags = body[9]
		case 4:
			n = int(synchsafe(body[4:8]))
			formatFlags = body[9]
		}
		if n < 0 || headerSize+n > len(body) {
			break
		}
		data := body[headerSize : headerSize+n]
		body = body[headerSize+n:]
		if data = frameData(version, flags, formatFlags, data); data == nil {
			continue
		}
		switch id {
		case "TIT2", "TT2":
			info.Title = id3Text(data)
		case "TPE1", "TP1":
			info.Performer = id3Text(data)
		case "APIC", "PIC":
			picture, kind, ok := id3Picture(data, version == 2)
			// The front cover is the one; any other does while there is none.
			if ok && coverType != 3 && (kind == 3 || coverType < 0) {
				info.Cover, coverType = picture, kind
			}
		}
	}
	return info
}

// frameData undoes what a frame's flags did to its data; nil for a frame
// compressed or encrypted, which is passed over.
func frameData(version, tagFlags, flags byte, data []byte) []byte {
	switch version {
	case 3:
		if flags&0xc0 != 0 {
			return nil
		}
		if flags&0x20 != 0 && len(data) > 0 {
			data = data[1:] // the group
		}
	case 4:
		if flags&0x0c != 0 {
			return nil
		}
		if flags&0x40 != 0 && len(data) > 0 {
			data = data[1:] // the group
		}
		if flags&0x01 != 0 {
			if len(data) < 4 {
				return nil
			}
			data = data[4:] // the data length
		}
		if flags&0x02 != 0 || tagFlags&0x80 != 0 {
			data = resync(data)
		}
	}
	return data
}

// resync undoes unsynchronisation: a zero byte put after each 0xff.
func resync(b []byte) []byte {
	if !bytes.Contains(b, []byte{0xff, 0}) {
		return b
	}
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		out = append(out, b[i])
		if b[i] == 0xff && i+1 < len(b) && b[i+1] == 0 {
			i++
		}
	}
	return out
}

// id3Text reads a text frame: its encoding, then its values, which 2.4
// separates with zeroes.
func id3Text(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	var values []string
	for _, v := range strings.Split(decodeText(data[0], data[1:]), "\x00") {
		if v = strings.TrimSpace(v); v != "" {
			values = append(values, v)
		}
	}
	return strings.Join(values, ", ")
}

// id3Picture reads an attached picture: its encoding, its format (a MIME
// type, or three letters in 2.2), its type, a description, the image.
func id3Picture(data []byte, v22 bool) (picture []byte, kind int, ok bool) {
	if len(data) < 2 {
		return nil, 0, false
	}
	encoding := data[0]
	rest := data[1:]
	if v22 {
		if len(rest) < 4 {
			return nil, 0, false
		}
		rest = rest[3:]
	} else {
		end := bytes.IndexByte(rest, 0)
		if end < 0 {
			return nil, 0, false
		}
		rest = rest[end+1:]
	}
	if len(rest) < 1 {
		return nil, 0, false
	}
	kind = int(rest[0])
	_, rest, ok = cutText(encoding, rest[1:])
	if !ok || len(rest) == 0 {
		return nil, 0, false
	}
	return rest, kind, true
}

// cutText splits a string in encoding, ended by its zero, from what
// follows.
func cutText(encoding byte, b []byte) (text, rest []byte, ok bool) {
	if encoding == 1 || encoding == 2 {
		for i := 0; i+1 < len(b); i += 2 {
			if b[i] == 0 && b[i+1] == 0 {
				return b[:i], b[i+2:], true
			}
		}
		return nil, nil, false
	}
	end := bytes.IndexByte(b, 0)
	if end < 0 {
		return nil, nil, false
	}
	return b[:end], b[end+1:], true
}

// decodeText decodes text in an ID3 encoding: 0 is ISO-8859-1, 1 UTF-16
// after a byte order mark, 2 UTF-16 big-endian, 3 UTF-8.
func decodeText(encoding byte, b []byte) string {
	switch encoding {
	case 1, 2:
		var order binary.ByteOrder = binary.BigEndian
		if encoding == 1 && len(b) >= 2 {
			switch {
			case b[0] == 0xff && b[1] == 0xfe:
				order, b = binary.LittleEndian, b[2:]
			case b[0] == 0xfe && b[1] == 0xff:
				b = b[2:]
			}
		}
		units := make([]uint16, 0, len(b)/2)
		for i := 0; i+1 < len(b); i += 2 {
			u := order.Uint16(b[i:])
			// Each value of a list may bring its own byte order mark.
			if u == 0xfeff {
				continue
			}
			if u == 0xfffe {
				if order == binary.ByteOrder(binary.BigEndian) {
					order = binary.LittleEndian
				} else {
					order = binary.BigEndian
				}
				continue
			}
			units = append(units, u)
		}
		return strings.TrimRight(string(utf16.Decode(units)), "\x00")
	case 3:
		return strings.TrimRight(strings.ToValidUTF8(string(b), "�"), "\x00")
	}
	return strings.TrimRight(latin1(b), "\x00")
}

func latin1(b []byte) string {
	r := make([]rune, len(b))
	for i, c := range b {
		r[i] = rune(c)
	}
	return string(r)
}

func trimNull(b []byte) []byte {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		b = b[:i]
	}
	return bytes.TrimRight(b, " ")
}
