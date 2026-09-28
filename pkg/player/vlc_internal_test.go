// SPDX-License-Identifier: Unlicense OR MIT

package player

import (
	"testing"
	"time"
)

// TestVLCAnswers checks that answers go to the queries in the order they were
// sent, past the events and command results VLC writes between them.
func TestVLCAnswers(t *testing.T) {
	p := &vlc{queries: []string{"get_time", "get_length"}}
	for _, line := range []string{
		"status change: ( new input: http://127.0.0.1:1/token/video )",
		"status change: ( play state: 3 )",
		"seek: returned 0 (no error)",
		"12",
		"status change: ( time: 12s )",
		"36",
		"status change: ( pause state: 4 )",
	} {
		p.apply(line)
	}
	if p.status.Position != 12*time.Second || p.status.Duration != 36*time.Second {
		t.Errorf("position %v, duration %v; want 12s and 36s", p.status.Position, p.status.Duration)
	}
	if !p.status.Paused {
		t.Error("not paused after a pause state")
	}
	if len(p.queries) != 0 {
		t.Errorf("queries left: %v", p.queries)
	}
	p.apply("status change: ( play state: 3 )")
	if p.status.Paused {
		t.Error("still paused after a play state")
	}
	// An answer without a query in flight changes nothing.
	p.apply("7")
	if p.status.Position != 12*time.Second {
		t.Errorf("position %v after a stray number", p.status.Position)
	}
}
