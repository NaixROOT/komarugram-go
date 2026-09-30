// SPDX-License-Identifier: Unlicense OR MIT

package audio

import (
	"encoding/binary"
	"io"
	"testing"
	"time"
)

// silence is n frames of silence: a test that plays it on the system's
// output is not heard.
type silence struct{ n, pos int64 }

func (s *silence) Read(pcm []Frame) (int, error) {
	m := min(int64(len(pcm)), s.n-s.pos)
	if m <= 0 {
		return 0, io.EOF
	}
	clear(pcm[:m])
	s.pos += m
	return int(m), nil
}
func (s *silence) SeekSample(pos int64) error { s.pos = max(0, min(pos, s.n)); return nil }
func (s *silence) Position() int64            { return s.pos }

// counter puts out its own positions, left as they are and right negated,
// to see where a read starts.
type counter struct{ silence }

func (c *counter) Read(pcm []Frame) (int, error) {
	from := c.pos
	n, err := c.silence.Read(pcm)
	for i := range n {
		pcm[i] = Frame{int16(from + int64(i)), -int16(from + int64(i))}
	}
	return n, err
}

func TestStereoKeepsTheChannelsAndSeeks(t *testing.T) {
	s := newStereo(&counter{silence{n: 1000}})
	defer s.pump.close()
	ready(t, s.pump)
	b := make([]byte, 4*3)
	if n, err := s.Read(b); n != 12 || err != nil {
		t.Fatalf("%d, %v", n, err)
	}
	for i := range 3 {
		l, r := binary.LittleEndian.Uint16(b[4*i:]), binary.LittleEndian.Uint16(b[4*i+2:])
		if l != uint16(i) || r != uint16(-i) {
			t.Fatalf("frame %d: %d %d", i, l, r)
		}
	}
	if at, err := s.Seek(4*500, io.SeekStart); at != 2000 || err != nil {
		t.Fatalf("seek: %d, %v", at, err)
	}
	ready(t, s.pump)
	s.Read(b)
	if l, r := binary.LittleEndian.Uint16(b), binary.LittleEndian.Uint16(b[2:]); l != 500 || r != uint16(-500&0xffff) {
		t.Fatalf("after the seek, frame %d %d", l, r)
	}
}

// A playback ends, holds the output no longer, and plays again from the
// start; a seek moves it.
func TestPlaybackEndsAndReplays(t *testing.T) {
	if _, err := open(); err != nil {
		t.Skip("no sound output:", err)
	}
	src := &silence{n: Rate / 5}
	p, err := Play(src)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	wait := func(what string, done func() bool) {
		t.Helper()
		for deadline := time.Now().Add(5 * time.Second); !done(); time.Sleep(10 * time.Millisecond) {
			if time.Now().After(deadline) {
				t.Fatal(what)
			}
		}
	}
	wait("it did not end", func() bool { return !p.Playing() })
	if !p.Ended() || playing != 0 {
		t.Fatalf("ended %t, holding the output %d times", p.Ended(), playing)
	}
	p.Resume()
	if !p.Playing() || p.Ended() || playing != 1 {
		t.Fatalf("replay: playing %t, ended %t, holding %d", p.Playing(), p.Ended(), playing)
	}
	p.Pause()
	if err := p.SeekSample(Rate / 10); err != nil {
		t.Fatal(err)
	}
	if p.Playing() || p.Position() != Rate/10 || playing != 0 {
		t.Fatalf("seek while paused: playing %t, at %d, holding %d", p.Playing(), p.Position(), playing)
	}
	p.Resume()
	wait("it did not end again", func() bool { return !p.Playing() })
	p.Close()
	if playing != 0 {
		t.Fatalf("holding the output %d times after Close", playing)
	}
}

// ready waits until the pump has read something ahead.
func ready(t *testing.T, p *pump) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); ; time.Sleep(time.Millisecond) {
		p.mu.Lock()
		queued, seeking := p.queued, p.seeking
		p.mu.Unlock()
		if queued > 0 && !seeking {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("the pump read nothing")
		}
	}
}
