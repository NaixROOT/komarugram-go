// SPDX-License-Identifier: Unlicense OR MIT

package opus

import (
	"context"
	"io"
	"math"
	"os"
	"testing"
	"time"

	"komarugram/pkg/audio"
)

// testdata/tone.ogg is 2.5 s of 440 Hz then 2.5 s of 660 Hz, mono Opus at
// 24 kbit/s, as ffmpeg encodes a voice message:
//
//	ffmpeg -f lavfi -i sine=frequency=440:duration=2.5 -f lavfi -i sine=frequency=660:duration=2.5 \
//	  -filter_complex "[0][1]concat=n=2:v=0:a=1,volume=0.5" -ac 1 -c:a libopus -b:a 24k -application voip tone.ogg
func tone(t *testing.T) (*Stream, *Decoder) {
	t.Helper()
	data, err := os.ReadFile("testdata/tone.ogg")
	if err != nil {
		t.Fatal(err)
	}
	s, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { rt.Close(ctx) })
	d, err := rt.NewDecoder(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return s, d
}

// frequency is what pcm's zero crossings tell of its pitch.
func frequency(pcm []int16) float64 {
	crossings := 0
	for i := 1; i < len(pcm); i++ {
		if (pcm[i-1] < 0) != (pcm[i] < 0) {
			crossings++
		}
	}
	return float64(crossings) / 2 / (float64(len(pcm)) / Rate)
}

// read reads n frames of r, as their left channel, which is the sound.
func read(t *testing.T, r *Reader, n int) []int16 {
	t.Helper()
	out := make([]audio.Frame, n)
	got := 0
	for got < n {
		m, err := r.Read(out[got:])
		got += m
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	pcm := make([]int16, got)
	for i := range pcm {
		if out[i][0] != out[i][1] {
			t.Fatalf("frame %d has %d on the left and %d on the right", i, out[i][0], out[i][1])
		}
		pcm[i] = out[i][0]
	}
	return pcm
}

func TestToneDecodes(t *testing.T) {
	s, d := tone(t)
	if got := s.Duration(); got != 5*time.Second {
		t.Fatalf("duration %v", got)
	}
	r, err := s.NewReader(context.Background(), d)
	if err != nil {
		t.Fatal(err)
	}
	pcm := read(t, r, int(s.Samples())+Rate)
	if int64(len(pcm)) != s.Samples() {
		t.Fatalf("read %d samples of %d", len(pcm), s.Samples())
	}
	// A tenth of a second in the middle of each tone.
	for _, c := range []struct {
		at   time.Duration
		want float64
	}{{time.Second, 440}, {4 * time.Second, 660}} {
		from := int(c.at * Rate / time.Second)
		if f := frequency(pcm[from : from+Rate/10]); math.Abs(f-c.want) > 15 {
			t.Errorf("at %v: %.0f Hz, want %.0f", c.at, f, c.want)
		}
	}
	if _, err := r.Read(make([]audio.Frame, 10)); err != io.EOF {
		t.Fatalf("after the end: %v", err)
	}
}

// Seeking lands where reading from the start does, forwards and back.
func TestToneSeeks(t *testing.T) {
	s, d := tone(t)
	r, err := s.NewReader(context.Background(), d)
	if err != nil {
		t.Fatal(err)
	}
	whole := read(t, r, int(s.Samples()))
	for _, at := range []int64{4 * Rate, Rate / 2, 3*Rate + 1234, 0} {
		if err := r.SeekSample(at); err != nil {
			t.Fatal(err)
		}
		if r.Position() != at {
			t.Fatalf("at %d after seeking to %d", r.Position(), at)
		}
		got := read(t, r, Rate/10)
		want := whole[at : at+Rate/10]
		// After the pre-roll the decoder has converged: the samples are
		// those of reading from the start, give or take a little.
		worst := 0
		for i := range got {
			worst = max(worst, abs(int(got[i])-int(want[i])))
		}
		if worst > 300 {
			t.Errorf("seek to %d: samples differ by up to %d", at, worst)
		}
	}
	if err := r.SeekSample(s.Samples() + Rate); err != nil || r.Position() != s.Samples() {
		t.Fatalf("seek past the end: at %d, %v", r.Position(), err)
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// What is not an OGG file of Opus is refused; a file cut short plays what
// came whole.
func TestParseRejects(t *testing.T) {
	data, err := os.ReadFile("testdata/tone.ogg")
	if err != nil {
		t.Fatal(err)
	}
	for name, bad := range map[string][]byte{
		"empty":   nil,
		"text":    []byte("not a sound at all, only some words here"),
		"head":    data[:40],
		"vorbis":  append([]byte("OggS\x00\x02"), make([]byte, 21)...),
		"garbage": append([]byte("OggS\x00"), make([]byte, 100)...),
	} {
		if _, err := Parse(bad); err == nil {
			t.Errorf("%s: parsed", name)
		}
	}
	cut, err := Parse(data[:len(data)/2])
	if err != nil {
		t.Fatal(err)
	}
	if d := cut.Duration(); d <= time.Second || d >= 4*time.Second {
		t.Fatalf("half a file plays %v", d)
	}
}

func TestPacketSamples(t *testing.T) {
	for _, c := range []struct {
		packet []byte
		want   int
	}{
		{[]byte{0 << 3}, 480},         // SILK 10 ms
		{[]byte{3 << 3}, 2880},        // SILK 60 ms
		{[]byte{13 << 3}, 960},        // Hybrid 20 ms
		{[]byte{16 << 3}, 120},        // CELT 2.5 ms
		{[]byte{31<<3 | 1}, 1920},     // CELT 20 ms, two frames
		{[]byte{31<<3 | 3, 6}, 5760},  // CELT 20 ms, six frames
		{[]byte{31<<3 | 3, 63}, 5760}, // too many frames: capped at 120 ms
		{[]byte{31<<3 | 3}, 0},        // no frame count
	} {
		if got := samples(c.packet); got != c.want {
			t.Errorf("%08b: %d, want %d", c.packet, got, c.want)
		}
	}
}
