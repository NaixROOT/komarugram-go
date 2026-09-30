// SPDX-License-Identifier: Unlicense OR MIT

package audio

import (
	"io"
	"math"
	"testing"
)

// hum is a sine of hz at rate in the left channel, and another in the right:
// a source at a rate of its own.
type hum struct {
	hz     [2]float64
	rate   float64
	n, pos int64
}

func (h *hum) Read(pcm []Frame) (int, error) {
	n := int(min(int64(len(pcm)), h.n-h.pos))
	if n <= 0 {
		return 0, io.EOF
	}
	for i := range n {
		for c, hz := range h.hz {
			pcm[i][c] = int16(math.Round(16000 * math.Sin(2*math.Pi*hz*float64(h.pos+int64(i))/h.rate)))
		}
	}
	h.pos += int64(n)
	return n, nil
}
func (h *hum) SeekSample(pos int64) error { h.pos = pos; return nil }
func (h *hum) Position() int64            { return h.pos }

// snr is how far the sound in channel c of out, from frame skip to the
// frame before len(out)-skip, is over its difference from the sine of hz at
// Rate, in dB.
func snr(out []Frame, c int, hz float64, skip int) float64 {
	var signal, noise float64
	for i := skip; i < len(out)-skip; i++ {
		want := 16000 * math.Sin(2*math.Pi*hz*float64(i)/Rate)
		d := float64(out[i][c]) - want
		signal += want * want
		noise += d * d
	}
	return 10 * math.Log10(signal/noise)
}

func resampled(t *testing.T, src Source, rate, length int64) []Frame {
	t.Helper()
	r, err := Resample(src, rate, length)
	if err != nil {
		t.Fatal(err)
	}
	return readAll(t, r)
}

// What is heard below the lower of the two Nyquist frequencies comes out as
// it went in, from the rates music has: the linear interpolation this
// replaces was 27 dB from it at 5 kHz, and 8 dB at 15 kHz.
func TestResampleKeepsTheSound(t *testing.T) {
	for _, rate := range []int64{44100, 32000, 22050, 96000, 88200} {
		for _, hz := range [][2]float64{{440, 1000}, {5000, 8000}, {12000, 15000}, {17000, 18000}} {
			n := rate
			out := resampled(t, &hum{hz: hz, rate: float64(rate), n: n}, rate, n)
			if want := n * Rate / rate; int64(len(out)) != want {
				t.Fatalf("%d Hz: %d frames, want %d", rate, len(out), want)
			}
			for c := range 2 {
				if hz[c] > 0.45*float64(min(rate, Rate)) {
					continue
				}
				// A second of it, less the ends, where the sound begins and stops.
				got := snr(out, c, hz[c], 2000)
				t.Logf("%d Hz, %.0f Hz: %.1f dB", rate, hz[c], got)
				if got < 80 {
					t.Errorf("%d Hz, %.0f Hz: %.1f dB, want 80", rate, hz[c], got)
				}
			}
		}
	}
}

// What the lower rate cannot hold is not folded into what it can: 30 kHz
// taken to 48 kHz would be heard at 18 kHz.
func TestResampleDropsWhatItCannotHold(t *testing.T) {
	const rate = 96000
	out := resampled(t, &hum{hz: [2]float64{30000, 30000}, rate: rate, n: rate}, rate, rate)
	var energy float64
	for _, f := range out[2000 : len(out)-2000] {
		energy += float64(f[0]) * float64(f[0])
	}
	rms := math.Sqrt(energy / float64(len(out)-4000))
	if db := 20 * math.Log10(rms/(16000/math.Sqrt2)); db > -80 {
		t.Fatalf("a 30 kHz tone left %.1f dB of itself", db)
	}
}

// Sound at Rate is not touched, and a steady level stays as it is.
func TestResampleAtOwnRateAndSteadyLevel(t *testing.T) {
	out := resampled(t, &hum{hz: [2]float64{440, 880}, rate: Rate, n: 5000}, Rate, 5000)
	want := readAll(t, &hum{hz: [2]float64{440, 880}, rate: Rate, n: 5000})
	for i := range want {
		if out[i] != want[i] {
			t.Fatalf("frame %d: %v, want %v", i, out[i], want[i])
		}
	}
	steady := &steadyLevel{n: 20000}
	for _, f := range resampled(t, steady, 44100, steady.n)[1000:19000] {
		if f != (Frame{10000, -5000}) {
			t.Fatalf("a steady level came out as %v", f)
		}
	}
}

type steadyLevel struct{ n, pos int64 }

func (s *steadyLevel) Read(pcm []Frame) (int, error) {
	n := int(min(int64(len(pcm)), s.n-s.pos))
	if n <= 0 {
		return 0, io.EOF
	}
	for i := range n {
		pcm[i] = Frame{10000, -5000}
	}
	s.pos += int64(n)
	return n, nil
}
func (s *steadyLevel) SeekSample(pos int64) error { s.pos = pos; return nil }
func (s *steadyLevel) Position() int64            { return s.pos }

// A move lands on the frames the sound has there: what is read after it is
// what was read at that place in one go.
func TestResampleSeeks(t *testing.T) {
	const rate, n = 44100, 44100
	src := &hum{hz: [2]float64{1000, 3000}, rate: rate, n: n}
	r, err := Resample(src, rate, n)
	if err != nil {
		t.Fatal(err)
	}
	all := readAll(t, r)
	for _, at := range []int64{0, 7, 12345, 40000} {
		if err := r.SeekSample(at); err != nil || r.Position() != at {
			t.Fatalf("seek to %d: at %d, %v", at, r.Position(), err)
		}
		buf := make([]Frame, 100)
		k, _ := r.Read(buf)
		for i := range k {
			if buf[i] != all[at+int64(i)] {
				t.Fatalf("after a move to %d, frame %d is %v, was %v", at, i, buf[i], all[at+int64(i)])
			}
		}
	}
}

// A source that ends before the length it told ends the sound there, and one
// that is longer is cut at the length.
func TestResampleEndsWhereTheSourceDoes(t *testing.T) {
	short := resampled(t, &hum{hz: [2]float64{440, 440}, rate: 44100, n: 20000}, 44100, 30000)
	if want := 20000 * int64(Rate) / 44100; math.Abs(float64(int64(len(short))-want)) > 2 {
		t.Fatalf("%d frames of a source of %d at 48 kHz", len(short), want)
	}
	long := resampled(t, &hum{hz: [2]float64{440, 440}, rate: 44100, n: 30000}, 44100, 20000)
	if want := 20000 * int64(Rate) / 44100; int64(len(long)) != want {
		t.Fatalf("%d frames, cut at %d", len(long), want)
	}
}

// Mono goes to both channels, stereo stays, 5.1 is folded to no louder than
// its loudest channel, and any other layout keeps its first two channels.
func TestDownmix(t *testing.T) {
	pack := func(channels int, frames ...[]int16) []byte {
		var raw []byte
		for _, f := range frames {
			for _, v := range f {
				raw = append(raw, byte(uint16(v)), byte(uint16(v)>>8))
			}
		}
		return raw
	}
	dst := make([]Frame, 4)
	if got := Downmix(dst, pack(1, []int16{5}, []int16{-7}), 1); len(got) != 2 || got[0] != (Frame{5, 5}) || got[1] != (Frame{-7, -7}) {
		t.Errorf("mono: %v", got)
	}
	if got := Downmix(dst, pack(2, []int16{5, -7}), 2); len(got) != 1 || got[0] != (Frame{5, -7}) {
		t.Errorf("stereo: %v", got)
	}
	if got := Downmix(dst, pack(4, []int16{1, 2, 3, 4}), 4); got[0] != (Frame{1, 2}) {
		t.Errorf("four channels: %v", got)
	}
	// Every channel of 5.1 full, but the low frequencies: no louder than full.
	got := Downmix(dst, pack(6, []int16{30000, 30000, 30000, 0, 30000, 30000}), 6)
	if d := int(got[0][0]) - 30000; d < -2 || d > 0 || got[0][0] != got[0][1] {
		t.Errorf("5.1 at full: %v", got[0])
	}
	// Front left alone comes out in the left.
	if got := Downmix(dst, pack(6, []int16{20000, 0, 0, 0, 0, 0}), 6); got[0][0] <= 0 || got[0][1] != 0 {
		t.Errorf("front left: %v", got[0])
	}
}
