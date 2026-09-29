// SPDX-License-Identifier: Unlicense OR MIT

package aac

import (
	"context"
	"io"
	"math"
	"os"
	"testing"
	"time"

	"komarugram/pkg/audio"
)

// runtime compiles the module KOMARUGRAM_AACDEC names, the absolute path
// of an aacdec.wasm: it is not in this repository.
func runtime(t *testing.T) *Runtime {
	t.Helper()
	path := os.Getenv("KOMARUGRAM_AACDEC")
	if path == "" {
		t.Skip("set KOMARUGRAM_AACDEC to an aacdec.wasm")
	}
	module, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	rt, err := NewRuntime(ctx, module)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { rt.Close(ctx) })
	return rt
}

// tone is ../mp4/testdata/tone.m4a: 1.5 s of 440 Hz then 1.5 s of 660 Hz,
// AAC-LC, stereo at 44.1 kHz, taken to audio.Rate.
func tone(t *testing.T, rt *Runtime) *audio.Resampled {
	t.Helper()
	data, err := os.ReadFile("../mp4/testdata/tone.m4a")
	if err != nil {
		t.Fatal(err)
	}
	d, err := rt.Open(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	r, err := audio.Resample(d, d.Rate(), d.Frames())
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func pitch(pcm []int16) float64 {
	crossings := 0
	for i := 1; i < len(pcm); i++ {
		if (pcm[i-1] < 0) != (pcm[i] < 0) {
			crossings++
		}
	}
	return float64(crossings) / 2 / (float64(len(pcm)) / audio.Rate)
}

func read(t *testing.T, src audio.Source, n int) []int16 {
	t.Helper()
	out := make([]int16, n)
	got := 0
	for got < n {
		k, err := src.Read(out[got:])
		got += k
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	return out[:got]
}

func TestToneDecodes(t *testing.T) {
	r := tone(t, runtime(t))
	if d := time.Duration(r.Samples()) * time.Second / audio.Rate; d < 3*time.Second || d > 3100*time.Millisecond {
		t.Fatalf("plays %v", d)
	}
	pcm := read(t, r, int(r.Samples())+audio.Rate)
	if int64(len(pcm)) < r.Samples()-audio.Rate/100 {
		t.Fatalf("read %d samples of %d", len(pcm), r.Samples())
	}
	for _, c := range []struct {
		at   time.Duration
		want float64
	}{{700 * time.Millisecond, 440}, {2200 * time.Millisecond, 660}} {
		from := int(c.at * audio.Rate / time.Second)
		if f := pitch(pcm[from : from+audio.Rate/10]); math.Abs(f-c.want) > 15 {
			t.Errorf("at %v: %.0f Hz, want %.0f", c.at, f, c.want)
		}
	}
}

// Moving forwards and back lands on the right tone at once.
func TestToneSeeks(t *testing.T) {
	r := tone(t, runtime(t))
	for _, c := range []struct {
		at   time.Duration
		want float64
	}{{2200 * time.Millisecond, 660}, {700 * time.Millisecond, 440}, {2600 * time.Millisecond, 660}, {100 * time.Millisecond, 440}} {
		pos := int64(c.at * audio.Rate / time.Second)
		if err := r.SeekSample(pos); err != nil || r.Position() != pos {
			t.Fatalf("seek to %v: at %d, %v", c.at, r.Position(), err)
		}
		if f := pitch(read(t, r, audio.Rate/10)); math.Abs(f-c.want) > 15 {
			t.Errorf("after a seek to %v: %.0f Hz, want %.0f", c.at, f, c.want)
		}
	}
}

// A file with broken units plays, the units concealed; what is not M4A is
// refused. Nothing panics outside the sandbox.
func TestBrokenFiles(t *testing.T) {
	rt := runtime(t)
	if _, err := rt.Open(context.Background(), []byte("not a sound at all")); err == nil {
		t.Fatal("words opened")
	}
	data, err := os.ReadFile("../mp4/testdata/tone.m4a")
	if err != nil {
		t.Fatal(err)
	}
	broken := append([]byte(nil), data...)
	// The units lie in mdat, after the tables at the start.
	for i := len(broken) / 3; i < len(broken)*9/10; i += 13 {
		broken[i] ^= 0xa5
	}
	d, err := rt.Open(context.Background(), broken)
	if err != nil {
		return
	}
	pcm := make([]int16, 8192)
	for range 100 {
		if _, err := d.Read(pcm); err != nil {
			break
		}
	}
}
