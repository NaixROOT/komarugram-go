// SPDX-License-Identifier: Unlicense OR MIT

package audio

import (
	"errors"
	"io"
)

// Resampled is mono sound at a rate of its own taken to Rate: each sample
// it puts out lies between the two of src around it. src's Read, SeekSample
// and Position count at its own rate.
type Resampled struct {
	src  Source
	rate int64
	// length and pos count at Rate.
	length, pos int64
	// buf are src's samples from its sample bufStart on; src is at
	// bufStart+len(buf).
	buf      []int16
	bufStart int64
	chunk    []int16
}

// Resample takes src, of length samples at rate, to Rate.
func Resample(src Source, rate, length int64) (*Resampled, error) {
	if rate <= 0 || rate > 8*Rate || length <= 0 {
		return nil, errors.New("audio: bad rate or length")
	}
	return &Resampled{src: src, rate: rate, length: length * Rate / rate, chunk: make([]int16, 4096)}, nil
}

// Samples is how many samples at Rate the sound plays.
func (r *Resampled) Samples() int64 { return r.length }

// Position is the next sample Read puts out, at Rate.
func (r *Resampled) Position() int64 { return r.pos }

// SeekSample moves to sample pos, at Rate.
func (r *Resampled) SeekSample(pos int64) error {
	r.pos = max(0, min(pos, r.length))
	r.bufStart, r.buf = r.pos*r.rate/Rate, r.buf[:0]
	return r.src.SeekSample(r.bufStart)
}

// Read puts out the next samples at Rate.
func (r *Resampled) Read(out []int16) (int, error) {
	n := 0
	for n < len(out) && r.pos < r.length {
		at := r.pos * r.rate // src's sample, times Rate
		i, frac := at/Rate-r.bufStart, at%Rate
		if i+1 >= int64(len(r.buf)) {
			if err := r.fill(); err != nil {
				if errors.Is(err, io.EOF) {
					// src ended before the length it told.
					r.length = r.pos
					break
				}
				return n, err
			}
			continue
		}
		a, b := int64(r.buf[i]), int64(r.buf[i+1])
		out[n] = int16(a + (b-a)*frac/Rate)
		n++
		r.pos++
	}
	if n == 0 && r.pos >= r.length {
		return 0, io.EOF
	}
	return n, nil
}

// fill reads more of src, dropping what lies before the next sample.
func (r *Resampled) fill() error {
	if drop := r.pos*r.rate/Rate - r.bufStart; drop > 0 && drop <= int64(len(r.buf)) {
		r.buf = append(r.buf[:0], r.buf[drop:]...)
		r.bufStart += drop
	}
	k, err := r.src.Read(r.chunk)
	r.buf = append(r.buf, r.chunk[:k]...)
	if k == 0 && err == nil {
		err = io.ErrNoProgress
	}
	if k > 0 && errors.Is(err, io.EOF) {
		err = nil
	}
	return err
}
