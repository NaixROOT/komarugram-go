// SPDX-License-Identifier: Unlicense OR MIT

package mp4

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"
)

// testdata/tone.m4a is 1.5 s of 440 Hz then 1.5 s of 660 Hz, AAC-LC,
// stereo at 44.1 kHz:
//
//	ffmpeg -f lavfi -i sine=frequency=440:duration=1.5 -f lavfi -i sine=frequency=660:duration=1.5 \
//	  -filter_complex "[0][1]concat=n=2:v=0:a=1,volume=0.5" -ar 44100 -ac 2 -c:a aac -b:a 64k tone.m4a
func TestDemuxAudio(t *testing.T) {
	data, err := os.ReadFile("testdata/tone.m4a")
	if err != nil {
		t.Fatal(err)
	}
	a, err := DemuxAudio(data)
	if err != nil {
		t.Fatal(err)
	}
	// AAC-LC (object type 2) at 44.1 kHz (index 4), two channels.
	if len(a.Config) < 2 || a.Config[0]>>3 != 2 || (a.Config[0]&7)<<1|a.Config[1]>>7 != 4 || a.Config[1]>>3&15 != 2 {
		t.Fatalf("config % x", a.Config)
	}
	if a.Channels != 2 || a.Rate != 44100 || a.Timescale != 44100 {
		t.Fatalf("%d channels at %d Hz, timescale %d", a.Channels, a.Rate, a.Timescale)
	}
	// A unit is 1024 samples; the encoder primes with one more.
	if n := len(a.Units); n < 3*44100/1024 || n > 3*44100/1024+3 {
		t.Fatalf("%d units", n)
	}
	if d := a.Duration(); d < 3*time.Second || d > 3100*time.Millisecond {
		t.Fatalf("plays %v", d)
	}
	for i, u := range a.Units {
		if u.Size == 0 || a.Starts[i+1]-a.Starts[i] != 1024 && i < len(a.Units)-1 {
			t.Fatalf("unit %d: %d bytes, %d samples", i, u.Size, a.Starts[i+1]-a.Starts[i])
		}
		b, err := a.Unit(i, nil)
		if err != nil || len(b) != int(u.Size) || !bytes.Equal(b, data[u.Offset:u.Offset+int64(u.Size)]) {
			t.Fatalf("unit %d read as %d bytes, %v", i, len(b), err)
		}
	}
}

// What is not MP4 is an error, and so are broken tables, not panics.
func TestDemuxAudioRejects(t *testing.T) {
	data, err := os.ReadFile("testdata/tone.m4a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DemuxAudio([]byte("not a file at all")); err == nil {
		t.Fatal("words demuxed")
	}
	for cut := 8; cut < len(data); cut += len(data) / 50 {
		DemuxAudio(data[:cut])
	}
	broken := append([]byte(nil), data...)
	for i := 0; i < len(broken); i += 97 {
		broken[i] ^= 0x5a
		DemuxAudio(broken)
	}
}

// counting reads a file, and counts what it reads.
type counting struct {
	data []byte
	read int
}

func (c *counting) ReadAt(p []byte, off int64) (int, error) {
	n := copy(p, c.data[min(int64(len(c.data)), off):])
	c.read += n
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// Of a file that plays as it downloads, the track is found from the box
// headers and the movie box, wherever it is: ffmpeg puts it at the end.
func TestDemuxAudioReadsLittle(t *testing.T) {
	data, err := os.ReadFile("testdata/tone.m4a")
	if err != nil {
		t.Fatal(err)
	}
	c := &counting{data: data}
	a, err := DemuxAudioAt(c, int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if c.read > len(data)/4 {
		t.Fatalf("read %d bytes of %d to find %d units", c.read, len(data), len(a.Units))
	}
}
