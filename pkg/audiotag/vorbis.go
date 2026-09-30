// SPDX-License-Identifier: Unlicense OR MIT

package audiotag

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"time"
)

// comments reads Vorbis comments, as FLAC and Ogg keep them: a vendor
// string, then KEY=value pairs. METADATA_BLOCK_PICTURE holds a cover as
// FLAC's picture block, in Base64.
func comments(b []byte, info *Info, coverType *int) {
	next := func() ([]byte, bool) {
		if len(b) < 4 {
			return nil, false
		}
		n := binary.LittleEndian.Uint32(b)
		if uint64(n) > uint64(len(b)-4) {
			return nil, false
		}
		s := b[4 : 4+n]
		b = b[4+n:]
		return s, true
	}
	if _, ok := next(); !ok {
		return
	}
	if len(b) < 4 {
		return
	}
	count := binary.LittleEndian.Uint32(b)
	b = b[4:]
	var titles, performers []string
	for range count {
		c, ok := next()
		if !ok {
			break
		}
		key, value, ok := strings.Cut(string(c), "=")
		if !ok {
			continue
		}
		value = strings.TrimSpace(strings.ToValidUTF8(value, "�"))
		switch strings.ToUpper(key) {
		case "TITLE":
			if value != "" {
				titles = append(titles, value)
			}
		case "ARTIST":
			if value != "" {
				performers = append(performers, value)
			}
		case "METADATA_BLOCK_PICTURE":
			if data, err := base64.StdEncoding.DecodeString(value); err == nil {
				flacPicture(data, info, coverType)
			}
		}
	}
	if len(titles) > 0 {
		info.Title = strings.Join(titles, ", ")
	}
	if len(performers) > 0 {
		info.Performer = strings.Join(performers, ", ")
	}
}

// flacPicture reads FLAC's picture block: its type, MIME type, description,
// size and colors, and the image. A front cover takes the place of any
// other picture.
func flacPicture(b []byte, info *Info, coverType *int) {
	field := func() ([]byte, bool) {
		if len(b) < 4 {
			return nil, false
		}
		n := binary.BigEndian.Uint32(b)
		if uint64(n) > uint64(len(b)-4) {
			return nil, false
		}
		s := b[4 : 4+n]
		b = b[4+n:]
		return s, true
	}
	if len(b) < 4 {
		return
	}
	kind := int(binary.BigEndian.Uint32(b))
	b = b[4:]
	if _, ok := field(); !ok { // MIME type
		return
	}
	if _, ok := field(); !ok { // description
		return
	}
	if len(b) < 16 {
		return
	}
	b = b[16:] // width, height, depth, colors
	data, ok := field()
	if !ok || len(data) == 0 {
		return
	}
	if *coverType != 3 && (kind == 3 || *coverType < 0) {
		info.Cover, *coverType = data, kind
	}
}

// readFLAC reads a FLAC file's metadata blocks: STREAMINFO for how long it
// plays, the comments and the pictures. An ID3v2 tag may come before it.
func readFLAC(r io.ReaderAt, size int64) (Info, error) {
	var head [10]byte
	if _, err := r.ReadAt(head[:], 0); err != nil {
		return Info{}, errors.New("not a FLAC file")
	}
	at := id3v2Size(head[:])
	var magic [4]byte
	if _, err := r.ReadAt(magic[:], at); err != nil || string(magic[:]) != "fLaC" {
		return Info{}, errors.New("not a FLAC file")
	}
	at += 4
	var info Info
	coverType := -1
	for {
		var h [4]byte
		if _, err := r.ReadAt(h[:], at); err != nil {
			return Info{}, errors.New("the FLAC file ends in its metadata")
		}
		last, kind := h[0]&0x80 != 0, h[0]&0x7f
		n := int64(h[1])<<16 | int64(h[2])<<8 | int64(h[3])
		at += 4
		if at+n > size {
			return Info{}, errors.New("the FLAC file ends in its metadata")
		}
		switch kind {
		case 0, 4, 6: // STREAMINFO, VORBIS_COMMENT, PICTURE
			block := make([]byte, n)
			if _, err := r.ReadAt(block, at); err != nil {
				return Info{}, err
			}
			switch kind {
			case 0:
				if len(block) < 18 {
					return Info{}, errors.New("no STREAMINFO")
				}
				// 20 bits of rate, 3 of channels, 5 of depth, 36 of samples.
				v := binary.BigEndian.Uint64(block[10:])
				rate := int64(v >> 44)
				samples := int64(v & (1<<36 - 1))
				if rate > 0 {
					info.Duration = time.Duration(samples) * time.Second / time.Duration(rate)
				}
			case 4:
				comments(block, &info, &coverType)
			case 6:
				flacPicture(block, &info, &coverType)
			}
		}
		at += n
		if last {
			return info, nil
		}
	}
}

// readOgg reads an Ogg file of Vorbis or Opus: the identification header
// for the rate, the comment header, and the position of the last page.
func readOgg(r io.ReaderAt, size int64) (Info, error) {
	packets, err := oggPackets(r, size, 2)
	if err != nil {
		return Info{}, err
	}
	var rate, skip int64
	var tags []byte
	id := packets[0]
	switch {
	case bytes.HasPrefix(id, []byte("OpusHead")) && len(id) >= 12:
		rate, skip = 48000, int64(binary.LittleEndian.Uint16(id[10:]))
		tags, _ = bytes.CutPrefix(packets[1], []byte("OpusTags"))
	case bytes.HasPrefix(id, []byte("\x01vorbis")) && len(id) >= 16:
		rate = int64(binary.LittleEndian.Uint32(id[12:]))
		tags, _ = bytes.CutPrefix(packets[1], []byte("\x03vorbis"))
	default:
		return Info{}, errors.New("the Ogg file holds neither Vorbis nor Opus")
	}
	if rate <= 0 {
		return Info{}, errors.New("no sample rate")
	}
	var info Info
	coverType := -1
	comments(tags, &info, &coverType)
	granule, err := lastGranule(r, size)
	if err != nil {
		return Info{}, err
	}
	info.Duration = time.Duration(max(0, granule-skip)) * time.Second / time.Duration(rate)
	return info, nil
}

// oggPackets reads the first n packets of an Ogg file's first stream.
func oggPackets(r io.ReaderAt, size int64, n int) ([][]byte, error) {
	var packets [][]byte
	var packet []byte
	var serial uint32
	at, total := int64(0), 0
	for len(packets) < n {
		var h [27]byte
		if _, err := r.ReadAt(h[:], at); err != nil || string(h[:4]) != "OggS" {
			return nil, errors.New("not an Ogg file")
		}
		if at == 0 {
			serial = binary.LittleEndian.Uint32(h[14:])
		}
		segments := make([]byte, h[26])
		if _, err := r.ReadAt(segments, at+27); err != nil {
			return nil, errors.New("the Ogg file ends in a page")
		}
		body := int64(0)
		for _, s := range segments {
			body += int64(s)
		}
		start := at + 27 + int64(len(segments))
		at = start + body
		if at > size {
			return nil, errors.New("the Ogg file ends in a page")
		}
		if binary.LittleEndian.Uint32(h[14:]) != serial {
			continue
		}
		data := make([]byte, body)
		if _, err := r.ReadAt(data, start); err != nil {
			return nil, err
		}
		for _, s := range segments {
			packet = append(packet, data[:s]...)
			data = data[s:]
			total += int(s)
			if total > maxTag {
				return nil, errors.New("the Ogg headers are too large")
			}
			// A segment shorter than 255 bytes ends a packet.
			if s < 255 {
				packets = append(packets, packet)
				packet = nil
				if len(packets) == n {
					break
				}
			}
		}
	}
	return packets, nil
}

// lastGranule is the position of the last page on which a packet ends.
func lastGranule(r io.ReaderAt, size int64) (int64, error) {
	// A page is at most 65307 bytes, so the last one starts in the tail.
	tailSize := min(size, 1<<17)
	tail := make([]byte, tailSize)
	if _, err := r.ReadAt(tail, size-tailSize); err != nil && err != io.EOF {
		return 0, err
	}
	for end := len(tail); ; {
		i := bytes.LastIndex(tail[:end], []byte("OggS"))
		if i < 0 || i+14 > len(tail) {
			return 0, errors.New("no Ogg page with a position")
		}
		// -1 marks a page on which no packet ends.
		if granule := int64(binary.LittleEndian.Uint64(tail[i+6:])); granule != -1 {
			return granule, nil
		}
		end = i
	}
}
