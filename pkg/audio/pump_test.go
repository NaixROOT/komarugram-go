// SPDX-License-Identifier: Unlicense OR MIT

package audio

import (
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

// slow is a ramp that takes its time to move, as an MP3 without a table of
// its frames does, and to read once it is told to.
type slow struct {
	ramp
	seekTime time.Duration
	moving   atomic.Bool
	stall    atomic.Bool
}

func (s *slow) SeekSample(pos int64) error {
	s.moving.Store(true)
	time.Sleep(s.seekTime)
	err := s.ramp.SeekSample(pos)
	s.moving.Store(false)
	return err
}

func (s *slow) Read(pcm []int16) (int, error) {
	for s.stall.Load() {
		time.Sleep(time.Millisecond)
	}
	return s.ramp.Read(pcm)
}

// within fails the test if f takes longer than limit.
func within(t *testing.T, what string, limit time.Duration, f func()) {
	t.Helper()
	start := time.Now()
	f()
	if d := time.Since(start); d > limit {
		t.Fatalf("%s took %v, more than %v", what, d, limit)
	}
}

// A move that takes long is not waited for: the position is where it was
// asked to go at once, the output gets silence meanwhile, and what it gets
// after is from there. Moving again drops the move under way.
func TestPumpDoesNotWaitForASlowMove(t *testing.T) {
	src := &slow{ramp: ramp{n: 30000}, seekTime: 300 * time.Millisecond}
	p := newPump(src)
	defer p.close()
	ready(t, p)
	buf := make([]int16, 100)
	within(t, "the move", 50*time.Millisecond, func() { p.seek(20000) })
	within(t, "the position", 50*time.Millisecond, func() {
		if at := p.position(); at != 20000 {
			t.Fatalf("position %d at once after a move to 20000", at)
		}
	})
	within(t, "the read", 50*time.Millisecond, func() {
		n, err := p.read(buf)
		if n != 0 || err != nil {
			t.Fatalf("while moving: %d samples, %v", n, err)
		}
	})
	// Moved again while it is moving: only the last counts.
	time.Sleep(50 * time.Millisecond)
	p.seek(10000)
	if at := p.position(); at != 10000 {
		t.Fatalf("position %d after another move", at)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		p.mu.Lock()
		done := !p.seeking && p.queued > 0
		p.mu.Unlock()
		if done {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the move did not end")
		}
		time.Sleep(time.Millisecond)
	}
	n, err := p.read(buf)
	if err != nil || n == 0 || buf[0] != 10000 {
		t.Fatalf("after the move: %d samples from %d, %v", n, buf[0], err)
	}
	if at := p.position(); at != 10000+int64(n) {
		t.Fatalf("position %d, want %d", at, 10000+int64(n))
	}
}

// A source that stalls is not waited for either.
func TestPumpDoesNotWaitForAStall(t *testing.T) {
	src := &slow{ramp: ramp{n: 1 << 20}}
	p := newPump(src)
	defer p.close()
	ready(t, p)
	// Take what is ahead, then stall the source.
	src.stall.Store(true)
	buf := make([]int16, 1000)
	for {
		p.mu.Lock()
		queued := p.queued
		p.mu.Unlock()
		if queued == 0 {
			break
		}
		p.read(buf)
	}
	within(t, "the read", 50*time.Millisecond, func() {
		if n, err := p.read(buf); n != 0 || err != nil {
			t.Fatalf("%d samples, %v", n, err)
		}
	})
	within(t, "the position", 50*time.Millisecond, func() { p.position() })
	src.stall.Store(false)
}

// The end of the source comes after what is ahead of it, and a failure too.
func TestPumpTellsTheEndAndFailures(t *testing.T) {
	p := newPump(&ramp{n: 5000})
	defer p.close()
	var got int
	buf := make([]int16, 700)
	for deadline := time.Now().Add(5 * time.Second); ; {
		n, err := p.read(buf)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if n > 0 && buf[0] != 0 || n > 0 {
			got += n
		}
		if time.Now().After(deadline) {
			t.Fatal("no end")
		}
		time.Sleep(time.Millisecond)
	}
	if got < 5000 {
		t.Fatalf("%d samples before the end", got)
	}
	failing := newPump(&failAt{ramp: ramp{n: 1 << 20}, at: 3000})
	defer failing.close()
	for deadline := time.Now().Add(5 * time.Second); ; time.Sleep(time.Millisecond) {
		if _, err := failing.read(buf); err != nil {
			if !errors.Is(err, errBroken) {
				t.Fatalf("failed with %v", err)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("no failure")
		}
	}
}

var errBroken = errors.New("broken")

type failAt struct {
	ramp
	at int64
}

func (f *failAt) Read(pcm []int16) (int, error) {
	if f.pos >= f.at {
		return 0, errBroken
	}
	return f.ramp.Read(pcm)
}

// What the window asks of a playback every frame, and a move it asks for,
// are answered at once, although the source takes long to move: the freeze
// of a window over a long MP3.
func TestPlaybackAnswersWhileTheSourceMoves(t *testing.T) {
	if _, err := open(); err != nil {
		t.Skip("no sound output:", err)
	}
	src := &slow{ramp: ramp{n: 10 * Rate}, seekTime: 500 * time.Millisecond}
	p, err := Play(src)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	within(t, "the move", 100*time.Millisecond, func() {
		if err := p.SeekSample(5 * Rate); err != nil {
			t.Fatal(err)
		}
	})
	for deadline := time.Now().Add(2 * time.Second); !src.moving.Load(); time.Sleep(time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatal("the source was not moved")
		}
	}
	for range 20 {
		within(t, "what the window asks", 100*time.Millisecond, func() {
			p.Position()
			p.Playing()
			p.Ended()
		})
		if !src.moving.Load() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if at := p.Position(); at < 5*Rate-Rate/5 {
		t.Fatalf("position %d after a move to %d", at, 5*Rate)
	}
}
