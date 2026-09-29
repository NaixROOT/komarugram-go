// SPDX-License-Identifier: Unlicense OR MIT

package drdec

import (
	"bytes"
	"context"
	"io"
	"math"
	"os"
	"testing"
	"time"

	"komarugram/pkg/audio"
)

// The test files, made with ffmpeg:
//
//	tone.mp3: 1.5 s of 440 Hz then 1.5 s of 660 Hz, stereo at 44.1 kHz
//	  ffmpeg -f lavfi -i sine=frequency=440:duration=1.5 -f lavfi -i sine=frequency=660:duration=1.5 \
//	    -filter_complex "[0][1]concat=n=2:v=0:a=1,volume=0.5" -ar 44100 -ac 2 -c:a libmp3lame -b:a 64k tone.mp3
//	tone.wav: 0.6 s of 440 Hz, mono at 22.05 kHz
//	  ffmpeg -f lavfi -i sine=frequency=440:duration=0.6 -ar 22050 -ac 1 -c:a pcm_s16le tone.wav
//	tone.flac: 0.6 s of 660 Hz, stereo at 48 kHz
//	  ffmpeg -f lavfi -i sine=frequency=660:duration=0.6 -ar 48000 -ac 2 -c:a flac tone.flac
func open(t *testing.T, name string, format Format) *audio.Resampled {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { rt.Close(ctx) })
	d, err := rt.Open(ctx, format, data)
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

func duration(r *audio.Resampled) time.Duration {
	return time.Duration(r.Samples()) * time.Second / audio.Rate
}

// Each format plays at audio.Rate for as long as it lasts, at its pitch.
func TestFormatsDecode(t *testing.T) {
	for _, c := range []struct {
		name     string
		format   Format
		min, max time.Duration
		at       time.Duration
		want     float64
	}{
		// MP3 encoders add a frame or two of silence; dr_mp3 cuts what the
		// LAME tag says.
		{"tone.mp3", MP3, 3 * time.Second, 3100 * time.Millisecond, 700 * time.Millisecond, 440},
		{"tone.mp3", MP3, 3 * time.Second, 3100 * time.Millisecond, 2200 * time.Millisecond, 660},
		{"tone.wav", WAV, 600 * time.Millisecond, 600 * time.Millisecond, 300 * time.Millisecond, 440},
		{"tone.flac", FLAC, 600 * time.Millisecond, 600 * time.Millisecond, 300 * time.Millisecond, 660},
	} {
		r := open(t, c.name, c.format)
		if d := duration(r); d < c.min || d > c.max {
			t.Errorf("%s plays %v", c.name, d)
		}
		pcm := read(t, r, int(r.Samples())+audio.Rate)
		if int64(len(pcm)) < r.Samples()-audio.Rate/100 {
			t.Errorf("%s: read %d samples of %d", c.name, len(pcm), r.Samples())
		}
		from := int(c.at * audio.Rate / time.Second)
		if f := pitch(pcm[from : from+audio.Rate/10]); math.Abs(f-c.want) > 15 {
			t.Errorf("%s at %v: %.0f Hz, want %.0f", c.name, c.at, f, c.want)
		}
	}
}

// Moving forwards and back lands on the right tone at once: the seek table
// takes MP3's bit reservoir into account.
func TestMP3Seeks(t *testing.T) {
	r := open(t, "tone.mp3", MP3)
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

// What is not a file of the format is refused, and a broken one fails in
// its sandbox, not in the process.
func TestRejects(t *testing.T) {
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close(ctx)
	data, err := os.ReadFile("testdata/tone.mp3")
	if err != nil {
		t.Fatal(err)
	}
	for _, format := range []Format{MP3, FLAC, WAV} {
		if _, err := rt.Open(ctx, format, []byte("not a sound at all, only some words here, and more")); err == nil {
			t.Errorf("format %d opened words", format)
		}
	}
	if _, err := rt.Open(ctx, FLAC, data); err == nil {
		t.Error("an MP3 opened as FLAC")
	}
	broken := append([]byte(nil), data...)
	for i := len(broken) / 3; i < 2*len(broken)/3; i += 7 {
		broken[i] ^= 0xa5
	}
	for _, b := range [][]byte{data[:len(data)/3], broken} {
		d, err := rt.Open(ctx, MP3, b)
		if err != nil {
			continue
		}
		pcm := make([]int16, 8192)
		for range 100 {
			if _, err := d.Read(pcm); err != nil {
				break
			}
		}
		d.Close()
	}
}

func TestFormatOf(t *testing.T) {
	for mime, want := range map[string]Format{"audio/mpeg": MP3, "audio/flac": FLAC, "audio/x-wav": WAV, "audio/ogg": 0, "audio/mp4": 0} {
		if got := FormatOf(mime); got != want {
			t.Errorf("%s: %d, want %d", mime, got, want)
		}
	}
}

// counting reads a file, counts what it reads, and waits first, once, as a
// download may.
type counting struct {
	data  []byte
	read  int
	delay time.Duration
}

func (c *counting) ReadAt(p []byte, off int64) (int, error) {
	if c.delay > 0 {
		time.Sleep(c.delay)
		c.delay = 0
	}
	n := copy(p, c.data[min(int64(len(c.data)), off):])
	c.read += n
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// A file that downloads as it plays is read a range at a time: an MP3 whose
// length Telegram tells starts without being read through.
func TestOpenAtReadsAsItPlays(t *testing.T) {
	data, err := os.ReadFile("testdata/tone.mp3")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close(ctx)
	// An MP3 stream is its frames one after another: twenty copies of the
	// file make one a minute long, whose reading shows.
	long := bytes.Repeat(data, 20)
	c := &counting{data: long}
	d, err := rt.OpenAt(ctx, MP3, c, int64(len(long)), 60*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	// dr_mp3 fills its input buffer, 64 KiB, and looks for tags at the
	// end: a part of a download, 128 KiB, whatever the file's length.
	if c.read > 128<<10 {
		t.Fatalf("read %d bytes of %d to open", c.read, len(long))
	}
	r, err := audio.Resample(d, d.Rate(), d.Frames())
	if err != nil {
		t.Fatal(err)
	}
	if got := duration(r); got != 60*time.Second {
		t.Fatalf("plays %v, not the length told", got)
	}
	pcm := read(t, r, audio.Rate)
	from := int(700 * time.Millisecond * audio.Rate / time.Second)
	if f := pitch(pcm[from : from+audio.Rate/10]); math.Abs(f-440) > 15 {
		t.Errorf("%.0f Hz, want 440", f)
	}
	for _, name := range []string{"tone.flac", "tone.wav"} {
		data, err := os.ReadFile("testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}
		format := FLAC
		if name == "tone.wav" {
			format = WAV
		}
		if _, err := rt.OpenAt(ctx, format, &counting{data: data}, int64(len(data)), 0); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// Waiting for the network is not the decoder being slow: the sandbox,
// which closes a decoder that takes over DefaultLimits.Slow, leaves it out.
func TestSlowReadsAreNotSlowDecoding(t *testing.T) {
	data, err := os.ReadFile("testdata/tone.wav")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close(ctx)
	d, err := rt.OpenAt(ctx, WAV, &counting{data: data, delay: DefaultLimits.Slow + 500*time.Millisecond}, int64(len(data)), 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Read(make([]int16, 1024)); err != nil {
		t.Fatal(err)
	}
}
