// SPDX-License-Identifier: Unlicense OR MIT

package audio

import (
	"errors"
	"io"
	"math"
	"sync/atomic"
)

// The stretch joins pieces of the sound, stretchSeq long, each started a
// tempo times longer step after the one before it, and moved within
// stretchSeek to where it continues the last one best; stretchOverlap of
// each is mixed with the end of the one before.
const (
	stretchSeq     = Rate * 40 / 1000
	stretchSeek    = Rate * 15 / 1000
	stretchOverlap = Rate * 8 / 1000
	stretchStep    = stretchSeq - stretchOverlap
)

// Stretched is a Source played faster or slower at the pitch it has, as a
// player's speed button does: pieces of the sound are left out, or
// repeated, where that is heard least (overlap-add with a search for the
// best place, as in SoundTouch). Both channels are cut at the same places,
// found in their mean, so that the stereo image keeps still. Its positions
// are those of the source.
//
// SetTempo may be called from any goroutine; the rest is for the one that
// reads.
type Stretched struct {
	src Source
	// tempo is the speed, in thousandths.
	tempo atomic.Int64
	// in are the frames read from the source and not left behind yet, from
	// frame base of it on, and at is where the next piece starts in them.
	in   []Frame
	base int64
	at   float64
	// tail is what followed the last piece, for the next one to continue.
	tail []Frame
	out  []Frame
	// drain is what was read ahead of the last piece when the speed went
	// back to the source's own: it is put out before the source is read, which
	// goes on from where it is, as a move of the source would cost a long MP3
	// the time to decode it to that place.
	drain []Frame
	// window and mean are the mean of the channels of the frames the best
	// place is looked for in, and of the tail.
	window, mean []float32
	// stretching is whether the last Read stretched; eof, whether the source
	// ended.
	stretching, eof bool
}

// Stretch returns src at its own speed, until SetTempo changes it.
func Stretch(src Source) *Stretched {
	s := &Stretched{src: src}
	s.tempo.Store(1000)
	return s
}

// SetTempo sets the speed: 1 is the source's own, 2 twice as fast. It is
// kept between 0.5 and 3.
func (s *Stretched) SetTempo(tempo float64) {
	s.tempo.Store(int64(math.Round(min(max(tempo, 0.5), 3) * 1000)))
}

// Tempo is the speed set.
func (s *Stretched) Tempo() float64 { return float64(s.tempo.Load()) / 1000 }

// reset forgets what was read ahead: the source is where the sound goes on.
func (s *Stretched) reset() {
	s.in, s.out, s.tail, s.at, s.eof = append(s.in[:0], s.drain...), s.out[:0], nil, 0, false
	s.base = s.src.Position() - int64(len(s.drain))
	s.drain = nil
}

func (s *Stretched) Read(pcm []Frame) (int, error) {
	tempo := s.tempo.Load()
	if tempo == 1000 {
		if s.stretching {
			// Back to the source's own speed, from where the sound is: what
			// came out of the last piece first, then what was read ahead.
			s.stretching = false
			s.drain = append(append(s.drain[:0], s.out...), s.in[min(int(s.at), len(s.in)):]...)
			s.in, s.out, s.tail = s.in[:0], s.out[:0], nil
		}
		if len(s.drain) > 0 {
			n := copy(pcm, s.drain)
			s.drain = s.drain[n:]
			return n, nil
		}
		return s.src.Read(pcm)
	}
	if !s.stretching {
		s.stretching = true
		s.reset()
	}
	for len(s.out) == 0 {
		if err := s.piece(float64(tempo) / 1000); err != nil {
			return 0, err
		}
	}
	n := copy(pcm, s.out)
	s.out = s.out[:copy(s.out, s.out[n:])]
	return n, nil
}

// piece puts the next piece of the sound into s.out, or returns io.EOF
// once the source has ended and all of it is out.
func (s *Stretched) piece(tempo float64) error {
	start := int(s.at)
	need := start + stretchSeek + stretchSeq
	for len(s.in) < need && !s.eof {
		old := len(s.in)
		s.in = append(s.in, make([]Frame, need-old)...)
		n, err := s.src.Read(s.in[old:])
		s.in = s.in[:old+n]
		if errors.Is(err, io.EOF) {
			s.eof = true
		} else if err != nil {
			return err
		} else if n == 0 {
			return io.ErrNoProgress
		}
	}
	if len(s.in) < need {
		// The end of the source: what is left goes out as it is.
		rest := s.in[min(start, len(s.in)):]
		s.out = append(s.out, rest...)
		s.base += int64(len(s.in))
		s.in, s.at, s.tail = s.in[:0], 0, nil
		if len(rest) == 0 {
			return io.EOF
		}
		return nil
	}
	if s.tail == nil {
		s.out = append(s.out, s.in[start:start+stretchStep]...)
		s.tail = append([]Frame(nil), s.in[start+stretchStep:start+stretchSeq]...)
	} else {
		from := start + s.bestOffset(s.in[start:start+stretchSeek+stretchOverlap])
		for i, old := range s.tail {
			var mixed Frame
			for c := range mixed {
				mixed[c] = int16((int(old[c])*(stretchOverlap-i) + int(s.in[from+i][c])*i) / stretchOverlap)
			}
			s.out = append(s.out, mixed)
		}
		s.out = append(s.out, s.in[from+stretchOverlap:from+stretchStep]...)
		copy(s.tail, s.in[from+stretchStep:from+stretchSeq])
	}
	s.at += tempo * stretchStep
	if drop := min(int(s.at), len(s.in)); drop > 0 {
		s.in = s.in[:copy(s.in, s.in[drop:])]
		s.at -= float64(drop)
		s.base += int64(drop)
	}
	return nil
}

// bestOffset is where in window, which is stretchSeek longer than the tail,
// the sound is most like the tail: the offset of the greatest normalized
// correlation of their means.
func (s *Stretched) bestOffset(window []Frame) int {
	s.window = means(s.window[:0], window)
	s.mean = means(s.mean[:0], s.tail)
	best, bestScore := 0, math.Inf(-1)
	for k := 0; k+len(s.mean) <= len(s.window); k++ {
		var corr, norm float32
		for i, t := range s.mean {
			v := s.window[k+i]
			corr += v * t
			norm += v * v
		}
		if score := float64(corr) / math.Sqrt(float64(norm)+1); score > bestScore {
			best, bestScore = k, score
		}
	}
	return best
}

// means appends the mean of the channels of each of frames to dst.
func means(dst []float32, frames []Frame) []float32 {
	for _, f := range frames {
		dst = append(dst, float32(int(f[0])+int(f[1]))/2)
	}
	return dst
}

// SeekSample moves to frame pos of the source.
func (s *Stretched) SeekSample(pos int64) error {
	if err := s.src.SeekSample(pos); err != nil {
		return err
	}
	s.drain = nil
	if s.stretching {
		s.reset()
	}
	return nil
}

// Position is the frame of the source Read puts out next.
func (s *Stretched) Position() int64 {
	if !s.stretching {
		return max(0, s.src.Position()-int64(len(s.drain)))
	}
	// What waits in out was taken from before the next piece.
	behind := int64(float64(len(s.out)) * s.Tempo())
	return max(0, s.base+int64(s.at)-behind)
}
