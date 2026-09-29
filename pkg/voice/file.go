// SPDX-License-Identifier: Unlicense OR MIT

package voice

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileExtensions are the files Telegram plays as voice messages, as the Bot
// API's sendVoice takes them: Opus in OGG, MP3 and M4A.
var FileExtensions = []string{".ogg", ".oga", ".opus", ".mp3", ".m4a"}

// ErrFileFormat is a file that is not one of FileExtensions' formats.
var ErrFileFormat = errors.New("a voice message is Opus in OGG, MP3 or M4A")

// FileMIME is the type a voice message file is sent with, by its name, or
// "" for a file that cannot be one.
func FileMIME(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ogg", ".oga", ".opus":
		return "audio/ogg"
	case ".mp3":
		return "audio/mpeg"
	case ".m4a":
		return "audio/mp4"
	}
	return ""
}

// FileDuration reads how long the voice message file at path plays, from
// its headers, without decoding it.
func FileDuration(path string) (time.Duration, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return 0, err
	}
	var d time.Duration
	switch FileMIME(path) {
	case "audio/ogg":
		d, err = oggOpusDuration(f, info.Size())
	case "audio/mpeg":
		d, err = mp3Duration(f)
	case "audio/mp4":
		d, err = mp4Duration(f, info.Size())
	default:
		return 0, ErrFileFormat
	}
	if err != nil {
		return 0, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s: no sound", filepath.Base(path))
	}
	return d, nil
}

// oggOpusDuration is the granule position of the last page, in 48 kHz
// samples, less the pre-skip the OpusHead packet states.
func oggOpusDuration(r io.ReaderAt, size int64) (time.Duration, error) {
	head := make([]byte, min(size, 4096))
	if _, err := r.ReadAt(head, 0); err != nil && err != io.EOF {
		return 0, err
	}
	if !bytes.HasPrefix(head, []byte("OggS")) {
		return 0, errors.New("not an OGG file")
	}
	at := bytes.Index(head, []byte("OpusHead"))
	if at < 0 || at+12 > len(head) {
		return 0, errors.New("the OGG file does not hold Opus")
	}
	preSkip := int64(binary.LittleEndian.Uint16(head[at+10:]))
	// A page is at most 65307 bytes, so the last one starts in the tail.
	tailSize := min(size, 1<<17)
	tail := make([]byte, tailSize)
	if _, err := r.ReadAt(tail, size-tailSize); err != nil && err != io.EOF {
		return 0, err
	}
	for end := len(tail); ; {
		i := bytes.LastIndex(tail[:end], []byte("OggS"))
		if i < 0 || i+14 > len(tail) {
			return 0, errors.New("no OGG page with a position")
		}
		granule := int64(binary.LittleEndian.Uint64(tail[i+6:]))
		// -1 marks a page on which no packet ends.
		if granule != -1 {
			return time.Duration(max(0, granule-preSkip)) * time.Second / 48000, nil
		}
		end = i
	}
}

// mp3Duration adds up the samples of every MPEG audio layer III frame,
// past an ID3v2 tag. Walking the frames is exact for variable bit rates,
// whatever header the encoder wrote or left out.
func mp3Duration(r io.ReadSeeker) (time.Duration, error) {
	br := bufio.NewReaderSize(r, 64<<10)
	var header [10]byte
	if _, err := io.ReadFull(br, header[:]); err != nil {
		return 0, errors.New("not an MP3 file")
	}
	var start int64
	if string(header[:3]) == "ID3" {
		start = 10 + (int64(header[6]&0x7f)<<21 | int64(header[7]&0x7f)<<14 | int64(header[8]&0x7f)<<7 | int64(header[9]&0x7f))
		if header[5]&0x10 != 0 {
			start += 10 // footer
		}
	}
	if _, err := r.Seek(start, io.SeekStart); err != nil {
		return 0, err
	}
	br.Reset(r)
	var samples, rate int64
	frames := 0
	for {
		var h [4]byte
		if _, err := io.ReadFull(br, h[:]); err != nil {
			break
		}
		length, n, sampleRate, ok := mp3Frame(h)
		if !ok {
			if frames == 0 {
				return 0, errors.New("not an MP3 file")
			}
			// An ID3v1 tag, or junk, after the frames.
			break
		}
		body := make([]byte, length-4)
		if frames > 0 {
			body = nil
			if _, err := br.Discard(length - 4); err != nil {
				break
			}
		} else if _, err := io.ReadFull(br, body); err != nil {
			break
		}
		frames++
		rate = int64(sampleRate)
		// The first frame may be the encoder's Xing or Info header: a
		// silent frame that says how many frames follow, and in its LAME
		// tag the samples the encoder added before and after the sound.
		if count, trim, ok := xingHeader(body); ok {
			if count > 0 {
				return time.Duration(max(0, count*int64(n)-trim)) * time.Second / time.Duration(rate), nil
			}
			samples -= trim
			continue
		}
		samples += int64(n)
	}
	if rate == 0 {
		return 0, errors.New("not an MP3 file")
	}
	return time.Duration(samples) * time.Second / time.Duration(rate), nil
}

// xingHeader reads the Xing or Info header in the body of a first frame:
// the frames that follow, 0 when it does not say, and the samples of
// padding its LAME tag states.
func xingHeader(body []byte) (frames, trim int64, ok bool) {
	at := bytes.Index(body, []byte("Xing"))
	if at < 0 {
		at = bytes.Index(body, []byte("Info"))
	}
	// The header follows the side information, at most 32 bytes in.
	if at < 0 || at > 40 || at+8 > len(body) {
		return 0, 0, false
	}
	flags := binary.BigEndian.Uint32(body[at+4:])
	next := at + 8
	if flags&1 != 0 && next+4 <= len(body) {
		frames = int64(binary.BigEndian.Uint32(body[next:]))
		next += 4
	}
	for bit, size := range map[uint32]int{2: 4, 4: 100, 8: 4} {
		if flags&bit != 0 {
			next += size
		}
	}
	// The LAME tag: a 9-byte encoder name, then at byte 21 the encoder
	// delay and the padding, 12 bits each: samples that are not sound.
	if next+24 <= len(body) && isName(body[next:next+4]) {
		d := body[next+21:]
		delay := int64(d[0])<<4 | int64(d[1])>>4
		padding := int64(d[1]&0x0f)<<8 | int64(d[2])
		trim = delay + padding
	}
	return frames, trim, true
}

func isName(b []byte) bool {
	for _, c := range b {
		if (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') {
			return false
		}
	}
	return true
}

var (
	mp3Bitrates = [2][16]int{
		// MPEG-1 layer III, then MPEG-2 and 2.5, in kbit/s.
		{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 0},
		{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0},
	}
	mp3Rates = [3][3]int{{44100, 48000, 32000}, {22050, 24000, 16000}, {11025, 12000, 8000}}
)

// mp3Frame reads a layer III frame header: the frame's length in bytes,
// its samples and sample rate.
func mp3Frame(h [4]byte) (length, samples, rate int, ok bool) {
	if h[0] != 0xff || h[1]&0xe0 != 0xe0 {
		return 0, 0, 0, false
	}
	version := h[1] >> 3 & 3 // 3 MPEG-1, 2 MPEG-2, 0 MPEG-2.5
	layer := h[1] >> 1 & 3   // 1 is layer III
	bitrateIndex := h[2] >> 4
	rateIndex := h[2] >> 2 & 3
	if version == 1 || layer != 1 || bitrateIndex == 0 || bitrateIndex == 15 || rateIndex == 3 {
		return 0, 0, 0, false
	}
	padding := int(h[2] >> 1 & 1)
	v := 0 // index into mp3Rates
	switch version {
	case 2:
		v = 1
	case 0:
		v = 2
	}
	rate = mp3Rates[v][rateIndex]
	if v == 0 {
		return 144000*mp3Bitrates[0][bitrateIndex]/rate + padding, 1152, rate, true
	}
	return 72000*mp3Bitrates[1][bitrateIndex]/rate + padding, 576, rate, true
}

// mp4Duration is the duration the movie header, moov/mvhd, states.
func mp4Duration(r io.ReaderAt, size int64) (time.Duration, error) {
	moov, moovSize, ok := mp4Box(r, 0, size, "moov")
	if !ok {
		return 0, errors.New("not an M4A file")
	}
	mvhd, mvhdSize, ok := mp4Box(r, moov, moovSize, "mvhd")
	if !ok || mvhdSize < 32 {
		return 0, errors.New("not an M4A file")
	}
	b := make([]byte, 32)
	if _, err := r.ReadAt(b, mvhd); err != nil {
		return 0, err
	}
	var scale, duration uint64
	if b[0] == 1 {
		scale = uint64(binary.BigEndian.Uint32(b[20:]))
		duration = binary.BigEndian.Uint64(b[24:])
	} else {
		scale = uint64(binary.BigEndian.Uint32(b[12:]))
		duration = uint64(binary.BigEndian.Uint32(b[16:]))
	}
	if scale == 0 {
		return 0, errors.New("no time scale")
	}
	return time.Duration(duration * uint64(time.Second) / scale), nil
}

// mp4Box finds the box of type name among those between at and at+size,
// and returns where its content starts and its length.
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
		if length < headerSize || at+length > end {
			return 0, 0, false
		}
		if string(h[4:8]) == name {
			return at + headerSize, length - headerSize, true
		}
		at += length
	}
	return 0, 0, false
}
