// SPDX-License-Identifier: Unlicense OR MIT

package audio

import (
	"errors"
	"io"
	"testing"
)

// readAll reads src to its end, in the chunks an output asks for.
func readAll(t *testing.T, src Source) []Frame {
	t.Helper()
	var all []Frame
	buf := make([]Frame, 1000)
	for {
		n, err := src.Read(buf)
		all = append(all, buf[:n]...)
		if errors.Is(err, io.EOF) {
			return all
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

// readN reads n frames of src.
func readN(t *testing.T, src Source, n int) {
	t.Helper()
	buf := make([]Frame, n)
	for len(buf) > 0 {
		got, err := src.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		buf = buf[got:]
	}
}

// ramp puts out its own positions, in the left channel and negated in the
// right, and counts its moves.
type ramp struct {
	n, pos int64
	seeks  int
}

func (r *ramp) Read(pcm []Frame) (int, error) {
	if r.pos >= r.n {
		return 0, io.EOF
	}
	n := int(min(int64(len(pcm)), r.n-r.pos))
	for i := range n {
		pcm[i] = Frame{int16(r.pos + int64(i)), -int16(r.pos + int64(i))}
	}
	r.pos += int64(n)
	return n, nil
}
func (r *ramp) SeekSample(pos int64) error { r.pos = pos; r.seeks++; return nil }
func (r *ramp) Position() int64            { return r.pos }
