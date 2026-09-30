// SPDX-License-Identifier: Unlicense OR MIT

package drdec

import (
	"bytes"
	"context"
	"encoding/binary"
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

// read reads n frames of src, as the mean of their channels.
func read(t *testing.T, src audio.Source, n int) []int16 {
	t.Helper()
	out := make([]audio.Frame, n)
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
	mono := make([]int16, got)
	for i := range mono {
		mono[i] = out[i].Mono()
	}
	return mono
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
		pcm := make([]audio.Frame, 8192)
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
	if _, err := d.Read(make([]audio.Frame, 1024)); err != nil {
		t.Fatal(err)
	}
}

// An MP3 without a table of its frames is moved over a long stretch in
// several calls of the module, each short enough for the sandbox's time
// limit, and lands on the right tone, forwards and back.
func TestMP3WithoutTableIsMovedInSteps(t *testing.T) {
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
	defer func(was time.Duration) { seekSpan = was }(seekSpan)
	for _, span := range []time.Duration{time.Hour, 200 * time.Millisecond} {
		seekSpan = span
		// The length is what Telegram told, so there is no table.
		d, err := rt.OpenAt(ctx, MP3, bytes.NewReader(data), int64(len(data)), 60*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range []struct {
			at   time.Duration
			want float64
		}{{2200 * time.Millisecond, 660}, {700 * time.Millisecond, 440}, {2600 * time.Millisecond, 660}, {100 * time.Millisecond, 440}} {
			pos := int64(c.at) * d.Rate() / int64(time.Second)
			if err := d.SeekSample(pos); err != nil || d.Position() != pos {
				t.Fatalf("span %v: seek to %v: at %d, %v", span, c.at, d.Position(), err)
			}
			pcm := read(t, d, int(d.Rate()/10))
			if f := pitch(pcm); math.Abs(f*float64(d.Rate())/float64(audio.Rate)-c.want) > 15 {
				t.Errorf("span %v, after a seek to %v: %.0f Hz, want %.0f", span, c.at, f, c.want)
			}
		}
		// A move over less than the span is one call, and over more is several:
		// the ones forward from the start, from 0 to 2.2 s and to 2.6 s, are.
		if short := span < time.Second; short != (d.moves > 8) {
			t.Errorf("span %v: %d calls of the module", span, d.moves)
		}
	}
}

// A stereo file keeps its channels: they are not mixed into one.
func TestStereoIsKept(t *testing.T) {
	const rate, n = 22050, 22050
	data := make([]byte, 44, 44+4*n)
	copy(data, "RIFF")
	binary.LittleEndian.PutUint32(data[4:], uint32(36+4*n))
	copy(data[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(data[16:], 16)
	binary.LittleEndian.PutUint16(data[20:], 1)
	binary.LittleEndian.PutUint16(data[22:], 2)
	binary.LittleEndian.PutUint32(data[24:], rate)
	binary.LittleEndian.PutUint32(data[28:], 4*rate)
	binary.LittleEndian.PutUint16(data[32:], 4)
	binary.LittleEndian.PutUint16(data[34:], 16)
	copy(data[36:], "data")
	binary.LittleEndian.PutUint32(data[40:], 4*n)
	for i := range n {
		left := int16(10000 * math.Sin(2*math.Pi*440*float64(i)/rate))
		right := int16(-10000 * math.Sin(2*math.Pi*1100*float64(i)/rate))
		data = binary.LittleEndian.AppendUint16(data, uint16(left))
		data = binary.LittleEndian.AppendUint16(data, uint16(right))
	}
	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close(ctx)
	d, err := rt.Open(ctx, WAV, data)
	if err != nil {
		t.Fatal(err)
	}
	frames := make([]audio.Frame, 0, n)
	buf := make([]audio.Frame, 1000)
	for {
		k, err := d.Read(buf)
		frames = append(frames, buf[:k]...)
		if err != nil {
			break
		}
	}
	if len(frames) != n {
		t.Fatalf("%d frames of %d", len(frames), n)
	}
	for c, want := range []float64{440, 1100} {
		crossings := 0
		for i := 1; i < len(frames); i++ {
			if (frames[i-1][c] < 0) != (frames[i][c] < 0) {
				crossings++
			}
		}
		if hz := float64(crossings) / 2 * rate / float64(len(frames)); math.Abs(hz-want) > 10 {
			t.Errorf("channel %d: %.0f Hz, want %.0f", c, hz, want)
		}
	}
}
