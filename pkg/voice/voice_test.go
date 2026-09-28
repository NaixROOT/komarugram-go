// SPDX-License-Identifier: Unlicense OR MIT

package voice

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// unpack reads 5-bit bars as Telegram does: bar i is the 16-bit little
// endian word at byte 5i/8, shifted by 5i%8.
func unpack(data []byte) []int {
	var bars []int
	for i := 0; (i*5+5+7)/8 <= len(data) && i < WaveformBars; i++ {
		bit := i * 5
		word := int(data[bit/8])
		if bit/8+1 < len(data) {
			word |= int(data[bit/8+1]) << 8
		}
		bars = append(bars, (word>>(bit%8))&31)
	}
	return bars
}

func TestWaveform(t *testing.T) {
	// A ramp: bar i of 100 is i/99 of the loudest.
	pcm := make([]int16, 100*480)
	for i := range pcm {
		v := int16(i / 480 * 300)
		if i%2 == 1 {
			v = -v
		}
		pcm[i] = v
	}
	data := Waveform(pcm)
	if len(data) != 63 {
		t.Fatalf("%d bytes", len(data))
	}
	bars := unpack(data)
	if len(bars) != WaveformBars || bars[0] != 0 || bars[99] != 31 || bars[50] != 50*31/99 {
		t.Fatalf("bars %v", bars)
	}
	for i := 1; i < len(bars); i++ {
		if bars[i] < bars[i-1] {
			t.Fatalf("not a ramp at %d: %v", i, bars)
		}
	}
	if !bytes.Equal(Waveform(make([]int16, 1000)), make([]byte, 63)) {
		t.Fatal("silence has bars")
	}
}

func TestDshowAudio(t *testing.T) {
	list := `[dshow @ 000001] "Integrated Camera" (video)
[dshow @ 000001]   Alternative name "@device_pnp_\\?\usb"
[dshow @ 000001] "Microphone Array (Realtek(R) Audio)" (audio)
[dshow @ 000001]   Alternative name "@device_cm_{33D9A762}"
[in#0 @ 000002] "Headset (USB Audio)" (audio)
dummy: Immediate exit requested`
	want := []string{"Microphone Array (Realtek(R) Audio)", "Headset (USB Audio)"}
	if got := dshowAudio(list); !reflect.DeepEqual(got, want) {
		t.Fatalf("devices %q", got)
	}
}

func fixed(inputs [][]string) func(context.Context) ([][]string, error) {
	return func(context.Context) ([][]string, error) { return inputs, nil }
}

func ffmpeg(t *testing.T) string {
	t.Helper()
	path, err := FFmpeg("")
	if err != nil {
		t.Skip(err)
	}
	return path
}

// A tone stands in for the microphone; an input that fails is passed over
// for the next.
func TestRecordAndEncode(t *testing.T) {
	path := ffmpeg(t)
	tone := []string{"-re", "-f", "lavfi", "-i", "sine=frequency=440:sample_rate=48000"}
	r := startWith(context.Background(), path, fixed([][]string{{"-f", "no-such-format", "-i", "x"}, tone}))
	deadline := time.Now().Add(10 * time.Second)
	for r.Duration() < 500*time.Millisecond {
		if err := r.Failed(); err != nil {
			t.Fatal(err)
		}
		if time.Now().After(deadline) {
			t.Fatal("nothing recorded")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if r.Level() < 0.05 {
		t.Fatalf("level %v of a tone", r.Level())
	}
	pcm, err := r.Stop()
	if err != nil || Duration(pcm) < 500*time.Millisecond {
		t.Fatalf("stopped with %v of sound, %v", Duration(pcm), err)
	}
	out := filepath.Join(t.TempDir(), "voice.ogg")
	if err := Encode(context.Background(), path, pcm, out); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil || !bytes.HasPrefix(data, []byte("OggS")) || !bytes.Contains(data, []byte("OpusHead")) {
		t.Fatalf("not Opus in OGG: %v", err)
	}
	if probe, err := exec.LookPath("ffprobe"); err == nil {
		got, err := exec.Command(probe, "-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", out).Output()
		if err != nil {
			t.Fatal(err)
		}
		if d, err := time.ParseDuration(strings.TrimSpace(string(got)) + "s"); err != nil || (d-Duration(pcm)).Abs() > 100*time.Millisecond {
			t.Fatalf("encoded %s, recorded %v", got, Duration(pcm))
		}
	}
}

// With no input that works, the recorder says why.
func TestRecordFails(t *testing.T) {
	path := ffmpeg(t)
	r := startWith(context.Background(), path, fixed([][]string{{"-f", "no-such-format", "-i", "x"}}))
	deadline := time.Now().Add(10 * time.Second)
	for r.Failed() == nil {
		if time.Now().After(deadline) {
			t.Fatal("did not fail")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := r.Stop(); err == nil {
		t.Fatal("no error from a failed recording")
	}
}
