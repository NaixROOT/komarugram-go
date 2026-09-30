// SPDX-License-Identifier: Unlicense OR MIT

package audio

import (
	"errors"
	"io"
	"math"
)

// The resampling filter is a sinc, windowed by a Kaiser window: its half
// width is resampleHalf input samples, widened by how much more sound there
// is in a second of input than of output, and its cut-off is resampleCut of
// the lower of the two Nyquist frequencies less its own transition, so that
// what lies above the lower one, images and aliases, is 90 dB down while the
// music below 19 kHz is not touched.
const (
	resampleHalf = 48
	resampleBeta = 9.0
	resampleCut  = 0.47
	// maxPhases bounds the table of the filter's positions between two input
	// samples: there are as many as the rates need to be exact, 160 for
	// 44.1 kHz, up to this.
	maxPhases = 2048
	maxHalf   = 256
)

// filter is a windowed sinc at each of the positions between two input
// samples, phases of them, each taps long; the point it makes a sample of is
// after input sample i, and the taps start at i-before.
type filter struct {
	taps, phases, before int
	coef                 []float32
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// bessel0 is the modified Bessel function of the first kind, order 0.
func bessel0(x float64) float64 {
	sum, term := 1.0, 1.0
	for k := 1; k < 64; k++ {
		term *= (x / 2 / float64(k)) * (x / 2 / float64(k))
		sum += term
		if term < 1e-14*sum {
			break
		}
	}
	return sum
}

// newFilter makes the filter that takes sound of rate to Rate.
func newFilter(rate int64) *filter {
	phases := int(min(int64(maxPhases), Rate/gcd(rate, Rate)))
	// Sound of a higher rate than Rate is cut at Rate's Nyquist frequency,
	// which takes a filter that many times wider in input samples.
	scale := max(1, float64(rate)/Rate)
	half := min(maxHalf, int(math.Ceil(resampleHalf*scale)))
	cut := resampleCut / scale
	f := &filter{taps: 2 * half, phases: phases, before: half - 1, coef: make([]float32, phases*2*half)}
	norm := bessel0(resampleBeta)
	for p := range phases {
		frac := float64(p) / float64(phases)
		row := f.coef[p*f.taps : (p+1)*f.taps]
		sum := 0.0
		for k := range row {
			// The distance in input samples from tap k to the point.
			x := float64(k-f.before) - frac
			h := 2 * cut
			if x != 0 {
				u := 2 * cut * x * math.Pi
				h = 2 * cut * math.Sin(u) / u
			}
			r := x / float64(half)
			w := 0.0
			if math.Abs(r) < 1 {
				w = bessel0(resampleBeta*math.Sqrt(1-r*r)) / norm
			}
			v := h * w
			row[k] = float32(v)
			sum += v
		}
		// Every position passes a steady level as it is.
		for k := range row {
			row[k] = float32(float64(row[k]) / sum)
		}
	}
	return f
}

// Resampled is sound at a rate of its own taken to Rate: each frame it puts
// out is made of the frames of src around its place by a filter that keeps
// what is heard below the higher Nyquist frequency and leaves nothing above
// it. src's Read, SeekSample and Position count at its own rate.
type Resampled struct {
	src  Source
	rate int64
	// length and pos count at Rate.
	length, pos int64
	f           *filter
	// buf are src's frames from bufStart on; src is at bufStart+len(buf), or
	// has ended, and what lies beyond is silence.
	buf      []Frame
	bufStart int64
	eof      bool
	chunk    []Frame
	window   []Frame
}

// Resample takes src, of length frames at rate, to Rate.
func Resample(src Source, rate, length int64) (*Resampled, error) {
	if rate <= 0 || rate > 8*Rate || length <= 0 {
		return nil, errors.New("audio: bad rate or length")
	}
	r := &Resampled{src: src, rate: rate, length: length * Rate / rate, chunk: make([]Frame, 4096)}
	if rate != Rate {
		r.f = newFilter(rate)
		r.window = make([]Frame, r.f.taps)
	}
	return r, nil
}

// Samples is how many frames at Rate the sound plays.
func (r *Resampled) Samples() int64 { return r.length }

// Position is the next frame Read puts out, at Rate.
func (r *Resampled) Position() int64 { return r.pos }

// SeekSample moves to frame pos, at Rate.
func (r *Resampled) SeekSample(pos int64) error {
	r.pos = max(0, min(pos, r.length))
	r.buf, r.eof = r.buf[:0], false
	if r.f == nil {
		return r.src.SeekSample(r.pos)
	}
	r.bufStart = max(0, r.pos*r.rate/Rate-int64(r.f.before))
	return r.src.SeekSample(r.bufStart)
}

// Read puts out the next frames at Rate.
func (r *Resampled) Read(out []Frame) (int, error) {
	if r.f == nil {
		return r.readSame(out)
	}
	f := r.f
	n := 0
	for n < len(out) && r.pos < r.length {
		num := r.pos * r.rate
		i, frac := num/Rate, num%Rate
		first, last := i-int64(f.before), i+int64(f.taps-f.before-1)
		if !r.eof && r.bufStart+int64(len(r.buf)) <= last {
			if err := r.fill(first); err != nil {
				return n, err
			}
			continue
		}
		phase := int(frac * int64(f.phases) / Rate)
		coef := f.coef[phase*f.taps : (phase+1)*f.taps]
		win := r.frames(first, f.taps)
		var l, rr float32
		for k, c := range coef {
			l += c * float32(win[k][0])
			rr += c * float32(win[k][1])
		}
		out[n] = Frame{clip(int(math.Round(float64(l)))), clip(int(math.Round(float64(rr))))}
		n++
		r.pos++
	}
	if n == 0 && r.pos >= r.length {
		return 0, io.EOF
	}
	return n, nil
}

// frames is the n frames of src from frame first on, in the buffer, with
// silence where there are none.
func (r *Resampled) frames(first int64, n int) []Frame {
	at := first - r.bufStart
	if at >= 0 && at+int64(n) <= int64(len(r.buf)) {
		return r.buf[at : at+int64(n)]
	}
	for k := range n {
		j := at + int64(k)
		if j >= 0 && j < int64(len(r.buf)) {
			r.window[k] = r.buf[j]
		} else {
			r.window[k] = Frame{}
		}
	}
	return r.window[:n]
}

// fill reads more of src, dropping what lies before frame first.
func (r *Resampled) fill(first int64) error {
	if drop := min(first-r.bufStart, int64(len(r.buf))); drop > 0 {
		r.buf = r.buf[:copy(r.buf, r.buf[drop:])]
		r.bufStart += drop
	}
	k, err := r.src.Read(r.chunk)
	r.buf = append(r.buf, r.chunk[:k]...)
	switch {
	case errors.Is(err, io.EOF):
		r.eof = true
		// src ended before the length it told, or at it.
		r.length = min(r.length, (r.bufStart+int64(len(r.buf)))*Rate/r.rate)
		return nil
	case err != nil:
		return err
	case k == 0:
		return io.ErrNoProgress
	}
	return nil
}

// readSame is Read for sound that is at Rate already.
func (r *Resampled) readSame(out []Frame) (int, error) {
	out = out[:int(min(int64(len(out)), r.length-r.pos))]
	if len(out) == 0 {
		return 0, io.EOF
	}
	n, err := r.src.Read(out)
	r.pos += int64(n)
	if errors.Is(err, io.EOF) {
		// src ended before the length it told.
		r.length = r.pos
		if n > 0 {
			err = nil
		}
	}
	return n, err
}
