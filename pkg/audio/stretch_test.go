// SPDX-License-Identifier: Unlicense OR MIT

package audio

import (
	"io"
	"math"
	"testing"
)

// tone is a sine of hz in each channel as a Source; a zero is silence.
type tone struct {
	hz      [2]float64
	samples int64
	pos     int64
}

func (t *tone) Read(pcm []Frame) (int, error) {
	if t.pos >= t.samples {
		return 0, io.EOF
	}
	n := int(min(int64(len(pcm)), t.samples-t.pos))
	for i := range n {
		for c, hz := range t.hz {
			pcm[i][c] = int16(12000 * math.Sin(2*math.Pi*hz*float64(t.pos+int64(i))/Rate))
		}
	}
	t.pos += int64(n)
	return n, nil
}
func (t *tone) SeekSample(pos int64) error { t.pos = pos; return nil }
func (t *tone) Position() int64            { return t.pos }

// pitch is the frequency of channel c of pcm, by its zero crossings.
func pitch(pcm []Frame, c int) float64 {
	crossings := 0
	for i := 1; i < len(pcm); i++ {
		if (pcm[i-1][c] < 0) != (pcm[i][c] < 0) {
			crossings++
		}
	}
	return float64(crossings) / 2 * Rate / float64(len(pcm))
}

func TestStretchKeepsPitch(t *testing.T) {
	const samples = 3 * Rate
	for _, tempo := range []float64{1.5, 2, 0.5} {
		s := Stretch(&tone{hz: [2]float64{220, 330}, samples: samples})
		s.SetTempo(tempo)
		out := readAll(t, s)
		if want := float64(samples) / tempo; math.Abs(float64(len(out))-want) > 0.03*want {
			t.Errorf("tempo %v: %d samples out of %d, want about %.0f", tempo, len(out), samples, want)
		}
		for c, want := range []float64{220, 330} {
			if hz := pitch(out, c); math.Abs(hz-want) > 4 {
				t.Errorf("tempo %v: channel %d pitch %.1f Hz, want %.0f", tempo, c, hz, want)
			}
		}
		if s.Position() != samples {
			t.Errorf("tempo %v: position %d at the end, want %d", tempo, s.Position(), samples)
		}
	}
}

// A sound in one channel stays in it, and the other stays silent: the
// channels are cut where one place is found, and never mixed.
func TestStretchKeepsTheStereoImage(t *testing.T) {
	s := Stretch(&tone{hz: [2]float64{0, 440}, samples: 2 * Rate})
	s.SetTempo(1.5)
	out := readAll(t, s)
	for i, f := range out {
		if f[0] != 0 {
			t.Fatalf("frame %d has %d in the silent channel", i, f[0])
		}
	}
	if hz := pitch(out, 1); math.Abs(hz-440) > 4 {
		t.Fatalf("pitch %.1f Hz, want 440", hz)
	}
}

func TestStretchAtOwnSpeedIsTheSource(t *testing.T) {
	const samples = Rate
	out := readAll(t, Stretch(&tone{hz: [2]float64{220, 330}, samples: samples}))
	want := readAll(t, &tone{hz: [2]float64{220, 330}, samples: samples})
	if len(out) != len(want) {
		t.Fatalf("%d samples, want %d", len(out), len(want))
	}
	for i := range out {
		if out[i] != want[i] {
			t.Fatalf("sample %d differs", i)
		}
	}
}

// The position follows the source through a change of speed and a move.
func TestStretchPositionAndSeek(t *testing.T) {
	const samples = 4 * Rate
	s := Stretch(&tone{hz: [2]float64{220, 330}, samples: samples})
	readN(t, s, Rate/2)
	if s.Position() != Rate/2 {
		t.Fatalf("position %d after half a second", s.Position())
	}
	s.SetTempo(2)
	readN(t, s, Rate/2)
	// Half a second heard at twice the speed is a second of the source.
	if pos := s.Position(); math.Abs(float64(pos)-1.5*Rate) > 0.06*Rate {
		t.Fatalf("position %d after half a second more at 2x, want about %d", pos, 3*Rate/2)
	}
	if err := s.SeekSample(3 * Rate); err != nil {
		t.Fatal(err)
	}
	if s.Position() != 3*Rate {
		t.Fatalf("position %d after a move", s.Position())
	}
	s.SetTempo(1)
	before := s.Position()
	readN(t, s, 100)
	if s.Position() != before+100 {
		t.Fatalf("position %d, want %d, back at the source's speed", s.Position(), before+100)
	}
}

// Back at the source's own speed, the sound goes on from where it was with
// no move of the source, which a long MP3 pays for by decoding to the place:
// the samples after the change are the source's, one after the other, to
// its end.
func TestStretchBackToOwnSpeedDoesNotSeek(t *testing.T) {
	const samples = 30000
	src := &ramp{n: samples}
	s := Stretch(src)
	s.SetTempo(2)
	readN(t, s, 4000)
	s.SetTempo(1)
	pos := s.Position()
	rest := readAll(t, s)
	if src.seeks != 0 {
		t.Fatalf("the source was moved %d times", src.seeks)
	}
	if len(rest) == 0 || int64(rest[len(rest)-1][0]) != samples-1 {
		t.Fatalf("%d samples after the change, the last %d", len(rest), rest[len(rest)-1][0])
	}
	// What follows the piece in hand is the source's, one sample after the
	// other, to its end: the last run of consecutive samples starts not far
	// after the position the change was made at.
	at := len(rest) - 1
	for at > 0 && rest[at][0] == rest[at-1][0]+1 {
		at--
	}
	if d := int64(rest[at][0]) - pos; d < 0 || d > 2*stretchSeq {
		t.Fatalf("the source goes on from %d, the change was made at %d", rest[at][0], pos)
	}
}
