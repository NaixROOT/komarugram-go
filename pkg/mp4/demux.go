// SPDX-License-Identifier: Unlicense OR MIT

// Package mp4 takes apart the MP4 files Telegram uses for GIFs and animated
// avatars: H.264 video, without sound, looped.
//
// The container is parsed here, in Go, and only the compressed samples are
// handed to the decoder, as the webm package does for stickers. The file
// layout is attacker controlled: every count and offset is checked against
// the data before anything is allocated or read, and a malformed structure
// is an error rather than a panic.
package mp4

import (
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"time"
)

// MaxSamples bounds the frames of one file. A GIF is at most a minute of
// video at 60 fps; this leaves room for longer loops without letting a
// hostile count allocate much.
const MaxSamples = 1 << 16

// ErrNotMP4 is returned for data that is not an MP4 file at all.
var ErrNotMP4 = errors.New("mp4: not an MP4 file")

// ErrUnsupported is returned for an MP4 file without an H.264 video track.
var ErrUnsupported = errors.New("mp4: no H.264 video track")

// Sample is one coded picture in decoding order, as length-prefixed NAL units.
type Sample struct {
	Data     []byte
	Keyframe bool
}

// File is the H.264 video track of an MP4 file.
type File struct {
	Width, Height int
	// Config is the AVCDecoderConfigurationRecord (avcC): the parameter sets
	// and the length of the NAL unit prefixes of the samples.
	Config []byte
	// Samples are in decoding order, which is also the order to feed them.
	Samples []Sample
	// Times are when the pictures show, from the first one, in the order a
	// decoder puts them out: the k-th picture out shows at Times[k].
	Times []time.Duration
	// Duration is how long one loop lasts.
	Duration time.Duration
}

// FrameDuration is the average time a picture shows.
func (f *File) FrameDuration() time.Duration {
	if len(f.Samples) == 0 {
		return 0
	}
	return f.Duration / time.Duration(len(f.Samples))
}

// track is what is read of one trak box.
type track struct {
	video                bool
	timescale            uint32
	width, height        int
	codec                string
	config               []byte
	deltas, offsets      []uint32 // stts and ctts, expanded per sample
	sizes                []uint32
	chunkOffsets         []uint64
	chunkRuns            []chunkRun
	keyframes            []uint32
	haveOffsets, haveSSS bool
}

// chunkRun is an stsc entry: from chunk first on, chunks hold samples each.
type chunkRun struct{ first, samples uint32 }

// Demux parses data and returns its first H.264 video track.
func Demux(data []byte) (*File, error) {
	var moov []byte
	sawFtyp := false
	if err := walk(data, func(kind string, body []byte) error {
		switch kind {
		case "ftyp":
			sawFtyp = true
		case "moov":
			if moov == nil {
				moov = body
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if !sawFtyp && moov == nil {
		return nil, ErrNotMP4
	}
	if moov == nil {
		return nil, fmt.Errorf("mp4: no moov box")
	}
	var found *track
	err := walk(moov, func(kind string, body []byte) error {
		if kind != "trak" || found != nil {
			return nil
		}
		t := &track{}
		if err := parseTrak(body, t); err != nil {
			return err
		}
		if t.video && (t.codec == "avc1" || t.codec == "avc3") && t.config != nil {
			found = t
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, ErrUnsupported
	}
	return found.file(data)
}

// walk calls fn for every box in data, in order.
func walk(data []byte, fn func(kind string, body []byte) error) error {
	for len(data) > 0 {
		if len(data) < 8 {
			return fmt.Errorf("mp4: truncated box header")
		}
		size := uint64(binary.BigEndian.Uint32(data))
		kind := string(data[4:8])
		header := uint64(8)
		switch size {
		case 0:
			size = uint64(len(data))
		case 1:
			if len(data) < 16 {
				return fmt.Errorf("mp4: truncated %q header", kind)
			}
			size, header = binary.BigEndian.Uint64(data[8:]), 16
		}
		if size < header || size > uint64(len(data)) {
			return fmt.Errorf("mp4: %q box of %d bytes in %d", kind, size, len(data))
		}
		if err := fn(kind, data[header:size]); err != nil {
			return err
		}
		data = data[size:]
	}
	return nil
}

// fullBox strips the version and flags of a full box.
func fullBox(body []byte) (version byte, rest []byte, err error) {
	if len(body) < 4 {
		return 0, nil, fmt.Errorf("mp4: truncated full box")
	}
	return body[0], body[4:], nil
}

func parseTrak(body []byte, t *track) error {
	return walk(body, func(kind string, body []byte) error {
		if kind != "mdia" {
			return nil
		}
		return walk(body, func(kind string, body []byte) error {
			switch kind {
			case "mdhd":
				version, rest, err := fullBox(body)
				if err != nil {
					return err
				}
				// Creation and modification times come before the timescale.
				at := 8
				if version == 1 {
					at = 16
				}
				if len(rest) < at+4 {
					return fmt.Errorf("mp4: truncated mdhd")
				}
				t.timescale = binary.BigEndian.Uint32(rest[at:])
			case "hdlr":
				_, rest, err := fullBox(body)
				if err != nil {
					return err
				}
				if len(rest) < 8 {
					return fmt.Errorf("mp4: truncated hdlr")
				}
				t.video = string(rest[4:8]) == "vide"
			case "minf":
				return walk(body, func(kind string, body []byte) error {
					if kind != "stbl" {
						return nil
					}
					return walk(body, func(kind string, body []byte) error {
						return parseTable(kind, body, t)
					})
				})
			}
			return nil
		})
	})
}

// entries reads the entry count of a table and checks that count entries of
// size bytes each fit in what follows it.
func entries(rest []byte, size int) (uint32, []byte, error) {
	if len(rest) < 4 {
		return 0, nil, fmt.Errorf("mp4: truncated table")
	}
	n := binary.BigEndian.Uint32(rest)
	rest = rest[4:]
	if uint64(n)*uint64(size) > uint64(len(rest)) {
		return 0, nil, fmt.Errorf("mp4: %d entries of %d bytes in %d", n, size, len(rest))
	}
	return n, rest, nil
}

// expand appends count copies of value, within MaxSamples.
func expand(out []uint32, count, value uint32) ([]uint32, error) {
	if uint64(len(out))+uint64(count) > MaxSamples {
		return nil, fmt.Errorf("mp4: more than %d samples", MaxSamples)
	}
	for range count {
		out = append(out, value)
	}
	return out, nil
}

func parseTable(kind string, body []byte, t *track) error {
	switch kind {
	case "stsd":
		_, rest, err := fullBox(body)
		if err != nil {
			return err
		}
		if len(rest) < 4 {
			return fmt.Errorf("mp4: truncated stsd")
		}
		return walk(rest[4:], func(kind string, body []byte) error {
			if t.codec != "" {
				return nil
			}
			t.codec = kind
			if kind != "avc1" && kind != "avc3" {
				return nil
			}
			// A VisualSampleEntry: 8 bytes of SampleEntry, 16 reserved, the
			// size, 50 more bytes, then its own boxes.
			if len(body) < 78 {
				return fmt.Errorf("mp4: truncated %s entry", kind)
			}
			t.width = int(binary.BigEndian.Uint16(body[24:]))
			t.height = int(binary.BigEndian.Uint16(body[26:]))
			return walk(body[78:], func(kind string, body []byte) error {
				if kind == "avcC" && t.config == nil {
					if len(body) < 7 || body[0] != 1 {
						return fmt.Errorf("mp4: bad avcC")
					}
					t.config = body
				}
				return nil
			})
		})
	case "stts", "ctts":
		_, rest, err := fullBox(body)
		if err != nil {
			return err
		}
		n, rest, err := entries(rest, 8)
		if err != nil {
			return err
		}
		out := &t.deltas
		if kind == "ctts" {
			out = &t.offsets
			t.haveOffsets = true
		}
		for i := range n {
			count := binary.BigEndian.Uint32(rest[i*8:])
			// Composition offsets are read as signed, as version 1 has them
			// and as muxers write them in version 0 too.
			value := binary.BigEndian.Uint32(rest[i*8+4:])
			if *out, err = expand(*out, count, value); err != nil {
				return err
			}
		}
	case "stsz":
		_, rest, err := fullBox(body)
		if err != nil {
			return err
		}
		if len(rest) < 8 {
			return fmt.Errorf("mp4: truncated stsz")
		}
		fixed := binary.BigEndian.Uint32(rest)
		if fixed != 0 {
			t.sizes, err = expand(nil, binary.BigEndian.Uint32(rest[4:]), fixed)
			return err
		}
		n, rest, err := entries(rest[4:], 4)
		if err != nil {
			return err
		}
		if n > MaxSamples {
			return fmt.Errorf("mp4: more than %d samples", MaxSamples)
		}
		t.sizes = make([]uint32, n)
		for i := range n {
			t.sizes[i] = binary.BigEndian.Uint32(rest[i*4:])
		}
	case "stco", "co64":
		_, rest, err := fullBox(body)
		if err != nil {
			return err
		}
		size := 4
		if kind == "co64" {
			size = 8
		}
		n, rest, err := entries(rest, size)
		if err != nil {
			return err
		}
		if n > MaxSamples {
			return fmt.Errorf("mp4: more than %d chunks", MaxSamples)
		}
		t.chunkOffsets = make([]uint64, n)
		for i := range n {
			if size == 4 {
				t.chunkOffsets[i] = uint64(binary.BigEndian.Uint32(rest[i*4:]))
			} else {
				t.chunkOffsets[i] = binary.BigEndian.Uint64(rest[i*8:])
			}
		}
	case "stsc":
		_, rest, err := fullBox(body)
		if err != nil {
			return err
		}
		n, rest, err := entries(rest, 12)
		if err != nil {
			return err
		}
		if n > MaxSamples {
			return fmt.Errorf("mp4: more than %d chunk runs", MaxSamples)
		}
		t.chunkRuns = make([]chunkRun, n)
		for i := range n {
			t.chunkRuns[i] = chunkRun{binary.BigEndian.Uint32(rest[i*12:]), binary.BigEndian.Uint32(rest[i*12+4:])}
		}
	case "stss":
		_, rest, err := fullBox(body)
		if err != nil {
			return err
		}
		n, rest, err := entries(rest, 4)
		if err != nil {
			return err
		}
		if n > MaxSamples {
			return fmt.Errorf("mp4: more than %d keyframes", MaxSamples)
		}
		t.haveSSS = true
		t.keyframes = make([]uint32, n)
		for i := range n {
			t.keyframes[i] = binary.BigEndian.Uint32(rest[i*4:])
		}
	}
	return nil
}

// file lays the samples of t out over data.
func (t *track) file(data []byte) (*File, error) {
	n := len(t.sizes)
	if n == 0 {
		return nil, fmt.Errorf("mp4: no samples")
	}
	if t.timescale == 0 {
		return nil, fmt.Errorf("mp4: no timescale")
	}
	if len(t.deltas) < n || t.haveOffsets && len(t.offsets) < n {
		return nil, fmt.Errorf("mp4: timing covers %d of %d samples", len(t.deltas), n)
	}
	if t.width <= 0 || t.height <= 0 {
		return nil, fmt.Errorf("mp4: size %dx%d", t.width, t.height)
	}
	file := &File{Width: t.width, Height: t.height, Config: t.config, Samples: make([]Sample, 0, n)}

	// Samples lie in chunks, which stsc says how many each holds.
	if len(t.chunkRuns) == 0 || t.chunkRuns[0].first != 1 {
		return nil, fmt.Errorf("mp4: bad stsc")
	}
	sample := 0
	for i, run := range t.chunkRuns {
		last := uint32(len(t.chunkOffsets))
		if i+1 < len(t.chunkRuns) {
			last = t.chunkRuns[i+1].first - 1
		}
		if run.first == 0 || last < run.first-1 || last > uint32(len(t.chunkOffsets)) || run.samples == 0 {
			return nil, fmt.Errorf("mp4: bad stsc entry %d", i)
		}
		for chunk := run.first; chunk <= last && sample < n; chunk++ {
			offset := t.chunkOffsets[chunk-1]
			for range run.samples {
				if sample >= n {
					break
				}
				size := uint64(t.sizes[sample])
				if offset > uint64(len(data)) || size > uint64(len(data))-offset {
					return nil, fmt.Errorf("mp4: sample %d lies outside the file", sample)
				}
				file.Samples = append(file.Samples, Sample{Data: data[offset : offset+size], Keyframe: !t.haveSSS})
				offset += size
				sample++
			}
		}
	}
	if sample < n {
		return nil, fmt.Errorf("mp4: chunks hold %d of %d samples", sample, n)
	}
	for _, k := range t.keyframes {
		if k >= 1 && int(k) <= n {
			file.Samples[k-1].Keyframe = true
		}
	}
	if !file.Samples[0].Keyframe {
		return nil, fmt.Errorf("mp4: the first sample is not a keyframe")
	}

	// A decoder puts pictures out in presentation order: sorted
	// presentation times are the times of its output, in turn.
	pts := make([]int64, n)
	var dts, end int64
	for i := range n {
		pts[i] = dts
		if t.haveOffsets {
			pts[i] += int64(int32(t.offsets[i]))
		}
		dts += int64(t.deltas[i])
		end = max(end, pts[i]+int64(t.deltas[i]))
	}
	slices.Sort(pts)
	// Hostile tables may add up to more than time.Duration holds in
	// nanoseconds; floating point saturates instead of wrapping.
	scale := func(v int64) time.Duration {
		return time.Duration(float64(v) / float64(t.timescale) * float64(time.Second))
	}
	file.Times = make([]time.Duration, n)
	for i, p := range pts {
		file.Times[i] = scale(p - pts[0])
	}
	file.Duration = scale(end - pts[0])
	if file.Duration <= 0 {
		return nil, fmt.Errorf("mp4: no duration")
	}
	return file, nil
}
