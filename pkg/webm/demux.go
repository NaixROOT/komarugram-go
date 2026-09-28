// SPDX-License-Identifier: Unlicense OR MIT

// Package webm takes apart the WebM files Telegram uses for video stickers.
//
// The container is parsed here, in Go, and only the compressed VP9 frames are
// handed to the decoder. That split matters: the file layout is attacker
// controlled, and a demuxer written in a memory-safe language cannot be talked
// into reading outside its own slices.
//
// Sticker transparency lives in the container rather than in the VP9 bitstream:
// the alpha channel is a second VP9 stream carried in each block's
// BlockAdditional element, which is why Frame has two payloads.
package webm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

// Matroska element ids used here. Ids keep their length marker, which is what
// makes them unique.
const (
	idSegment         = 0x18538067
	idInfo            = 0x1549A966
	idTimecodeScale   = 0x2AD7B1
	idTracks          = 0x1654AE6B
	idTrackEntry      = 0xAE
	idTrackNumber     = 0xD7
	idTrackType       = 0x83
	idCodecID         = 0x86
	idDefaultDuration = 0x23E383
	idVideo           = 0xE0
	idPixelWidth      = 0xB0
	idPixelHeight     = 0xBA
	idCluster         = 0x1F43B675
	idTimecode        = 0xE7
	idSimpleBlock     = 0xA3
	idBlockGroup      = 0xA0
	idBlock           = 0xA1
	idBlockAdditions  = 0x75A1
	idBlockMore       = 0xA6
	idBlockAddID      = 0xEE
	idBlockAdditional = 0xA5

	trackTypeVideo = 1
)

// ErrNotWebM is returned for data that is not a Matroska file at all.
var ErrNotWebM = errors.New("webm: not a Matroska file")

// Frame is one coded picture: the VP9 packet, plus the packet holding its
// alpha channel when the sticker has transparency.
type Frame struct {
	Data      []byte
	Alpha     []byte
	Timestamp time.Duration
}

// File is the video track of a WebM file, already split into frames.
type File struct {
	Width, Height int
	CodecID       string
	FrameDuration time.Duration
	Frames        []Frame
}

// HasAlpha reports whether any frame carries an alpha channel.
func (f *File) HasAlpha() bool {
	for i := range f.Frames {
		if len(f.Frames[i].Alpha) > 0 {
			return true
		}
	}
	return false
}

// Duration is how long one loop of the file lasts.
func (f *File) Duration() time.Duration {
	return time.Duration(len(f.Frames)) * f.FrameDuration
}

// Demux parses data and returns the frames of its first video track. The input
// is never trusted: every read is bounds checked and every malformed structure
// is an error rather than a panic.
func Demux(data []byte) (*File, error) {
	if len(data) < 4 || binary.BigEndian.Uint32(data) != 0x1A45DFA3 {
		return nil, ErrNotWebM
	}

	file := &File{}
	timecodeScale := uint64(1000000) // Matroska's default: one millisecond
	videoTrack := int64(-1)
	var clusterTime uint64

	top := &reader{data: data}
	// The EBML header is skipped: only the segment holds anything of interest.
	if err := top.walk(func(r *reader, id uint32, body []byte) error {
		if id != idSegment {
			return nil
		}
		return (&reader{data: body}).walk(func(r *reader, id uint32, body []byte) error {
			switch id {
			case idInfo:
				return (&reader{data: body}).walk(func(r *reader, id uint32, body []byte) error {
					if id == idTimecodeScale {
						if scale := beUint(body); scale > 0 {
							timecodeScale = scale
						}
					}
					return nil
				})

			case idTracks:
				return (&reader{data: body}).walk(func(r *reader, id uint32, body []byte) error {
					if id != idTrackEntry {
						return nil
					}
					return parseTrackEntry(body, file, &videoTrack)
				})

			case idCluster:
				return (&reader{data: body}).walk(func(r *reader, id uint32, body []byte) error {
					switch id {
					case idTimecode:
						clusterTime = beUint(body)
					case idSimpleBlock:
						return appendBlock(file, videoTrack, timecodeScale, clusterTime, body, nil)
					case idBlockGroup:
						return parseBlockGroup(file, videoTrack, timecodeScale, clusterTime, body)
					}
					return nil
				})
			}
			return nil
		})
	}); err != nil {
		return nil, err
	}

	if videoTrack < 0 {
		return nil, errors.New("webm: no video track")
	}
	if len(file.Frames) == 0 {
		return nil, errors.New("webm: no frames")
	}
	if file.FrameDuration <= 0 && len(file.Frames) > 1 {
		// Fall back to the spacing of the first two frames.
		file.FrameDuration = file.Frames[1].Timestamp - file.Frames[0].Timestamp
	}
	if file.FrameDuration <= 0 {
		file.FrameDuration = time.Second / 30
	}
	return file, nil
}

func parseTrackEntry(body []byte, file *File, videoTrack *int64) error {
	var (
		number   int64
		isVideo  bool
		codec    string
		duration time.Duration
		width    int
		height   int
	)
	err := (&reader{data: body}).walk(func(r *reader, id uint32, body []byte) error {
		switch id {
		case idTrackNumber:
			number = int64(beUint(body))
		case idTrackType:
			isVideo = beUint(body) == trackTypeVideo
		case idCodecID:
			codec = string(trimZero(body))
		case idDefaultDuration:
			duration = time.Duration(beUint(body))
		case idVideo:
			return (&reader{data: body}).walk(func(r *reader, id uint32, body []byte) error {
				switch id {
				case idPixelWidth:
					width = int(beUint(body))
				case idPixelHeight:
					height = int(beUint(body))
				}
				return nil
			})
		}
		return nil
	})
	if err != nil || !isVideo || *videoTrack >= 0 {
		return err
	}
	*videoTrack = number
	file.CodecID = codec
	file.FrameDuration = duration
	file.Width, file.Height = width, height
	return nil
}

func parseBlockGroup(file *File, videoTrack int64, scale, clusterTime uint64, body []byte) error {
	var block, alpha []byte
	err := (&reader{data: body}).walk(func(r *reader, id uint32, body []byte) error {
		switch id {
		case idBlock:
			block = body
		case idBlockAdditions:
			return (&reader{data: body}).walk(func(r *reader, id uint32, body []byte) error {
				if id != idBlockMore {
					return nil
				}
				var addID uint64 = 1
				var payload []byte
				if err := (&reader{data: body}).walk(func(r *reader, id uint32, body []byte) error {
					switch id {
					case idBlockAddID:
						addID = beUint(body)
					case idBlockAdditional:
						payload = body
					}
					return nil
				}); err != nil {
					return err
				}
				// Additional id 1 is the alpha channel for VP8/VP9 in WebM.
				if addID == 1 {
					alpha = payload
				}
				return nil
			})
		}
		return nil
	})
	if err != nil || block == nil {
		return err
	}
	return appendBlock(file, videoTrack, scale, clusterTime, block, alpha)
}

// appendBlock unpacks a Block or SimpleBlock header and keeps the payload.
func appendBlock(file *File, videoTrack int64, scale, clusterTime uint64, block, alpha []byte) error {
	r := &reader{data: block}
	track, err := r.vint(false)
	if err != nil {
		return err
	}
	if int64(track) != videoTrack {
		return nil
	}
	if r.offset+3 > len(block) {
		return fmt.Errorf("webm: truncated block header")
	}
	relative := int16(binary.BigEndian.Uint16(block[r.offset:]))
	flags := block[r.offset+2]
	payload := block[r.offset+3:]

	if lacing := (flags >> 1) & 0x3; lacing != 0 {
		// Lacing packs several frames into one block. Video tracks do not use
		// it, and guessing would risk handing the decoder a bogus packet.
		return fmt.Errorf("webm: laced video block (lacing %d) is not supported", lacing)
	}

	timestamp := time.Duration(clusterTime+uint64(int64(relative))) * time.Duration(scale)
	file.Frames = append(file.Frames, Frame{
		Data:      payload,
		Alpha:     alpha,
		Timestamp: timestamp,
	})
	return nil
}

// reader walks a sequence of EBML elements inside one slice.
type reader struct {
	data   []byte
	offset int
}

// walk calls visit for each element. Children are passed as sub-slices of the
// original buffer, so nothing is copied.
func (r *reader) walk(visit func(r *reader, id uint32, body []byte) error) error {
	for r.offset < len(r.data) {
		id, err := r.elementID()
		if err != nil {
			return err
		}
		size, unknown, err := r.size()
		if err != nil {
			return err
		}
		if unknown {
			// An element of unknown size runs to the end of its parent, which
			// is how streamed files are written.
			size = uint64(len(r.data) - r.offset)
		}
		if size > uint64(len(r.data)-r.offset) {
			return fmt.Errorf("webm: element 0x%X claims %d bytes, %d left", id, size, len(r.data)-r.offset)
		}
		body := r.data[r.offset : r.offset+int(size)]
		r.offset += int(size)
		if err := visit(r, id, body); err != nil {
			return err
		}
	}
	return nil
}

// elementID reads a variable-length id, marker bits included.
func (r *reader) elementID() (uint32, error) {
	if r.offset >= len(r.data) {
		return 0, errors.New("webm: truncated element id")
	}
	first := r.data[r.offset]
	length := 0
	switch {
	case first&0x80 != 0:
		length = 1
	case first&0x40 != 0:
		length = 2
	case first&0x20 != 0:
		length = 3
	case first&0x10 != 0:
		length = 4
	default:
		return 0, fmt.Errorf("webm: invalid element id 0x%02X", first)
	}
	if r.offset+length > len(r.data) {
		return 0, errors.New("webm: truncated element id")
	}
	var id uint32
	for _, b := range r.data[r.offset : r.offset+length] {
		id = id<<8 | uint32(b)
	}
	r.offset += length
	return id, nil
}

// size reads a variable-length size, reporting the "unknown size" encoding.
func (r *reader) size() (value uint64, unknown bool, err error) {
	start := r.offset
	value, err = r.vint(false)
	if err != nil {
		return 0, false, err
	}
	length := r.offset - start
	// All data bits set means the size is unknown.
	if value == uint64(1)<<(7*length)-1 {
		return 0, true, nil
	}
	return value, false, nil
}

// vint reads a variable-length integer, dropping the marker bit unless
// keepMarker asks for the raw value.
func (r *reader) vint(keepMarker bool) (uint64, error) {
	if r.offset >= len(r.data) {
		return 0, errors.New("webm: truncated integer")
	}
	first := r.data[r.offset]
	if first == 0 {
		return 0, errors.New("webm: invalid integer")
	}
	length := 1
	for mask := byte(0x80); first&mask == 0; mask >>= 1 {
		length++
	}
	if r.offset+length > len(r.data) {
		return 0, errors.New("webm: truncated integer")
	}
	value := uint64(first)
	if !keepMarker {
		value = uint64(first) & (0xFF >> length)
	}
	for _, b := range r.data[r.offset+1 : r.offset+length] {
		value = value<<8 | uint64(b)
	}
	r.offset += length
	return value, nil
}

// beUint reads a big-endian unsigned integer of any width, as Matroska stores it.
func beUint(body []byte) uint64 {
	var value uint64
	for _, b := range body {
		value = value<<8 | uint64(b)
	}
	return value
}

func trimZero(body []byte) []byte {
	for i, b := range body {
		if b == 0 {
			return body[:i]
		}
	}
	return body
}
