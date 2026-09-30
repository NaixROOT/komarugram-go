// SPDX-License-Identifier: Unlicense OR MIT

package audio

import (
	"errors"
	"io"
	"sync"
)

// The pump reads ahead pumpAhead samples, a pumpSize at a time.
const (
	pumpAhead = Rate * 150 / 1000
	pumpSize  = 2048
)

// pump reads a Source on a goroutine of its own, a little ahead of what the
// output takes, so that nothing waits for the source but the pump: not the
// output's reads, and not the window, which asks where the sound is every
// frame. A decoder that is slow to move (an MP3 without a table of its
// frames goes through all the file before the place it is moved to), or a
// download that stalls, leaves the output playing silence, and the rest
// answering at once.
type pump struct {
	src Source

	mu   sync.Mutex
	cond *sync.Cond
	// ring are the chunks read and not taken yet, queued their samples.
	ring   []pumpChunk
	queued int
	// end is where the source is after the chunks taken; the place the next
	// one starts at, when there is none in the ring.
	end int64
	// gen counts the moves: what was read before one is dropped.
	gen int
	// seeking is set from a move asked for until it is done; target is where.
	seeking bool
	target  int64
	// eof is set when the source ended, and err when it failed: both are told
	// once the ring is empty.
	eof    bool
	err    error
	closed bool
}

// pumpChunk is samples of the sound put out by the source between the
// positions from and to, of which taken have gone to the output.
type pumpChunk struct {
	pcm      []int16
	from, to int64
	taken    int
}

// newPump starts reading src, from where it is.
func newPump(src Source) *pump {
	p := &pump{src: src, end: src.Position()}
	p.cond = sync.NewCond(&p.mu)
	go p.run()
	return p
}

// run is the pump's goroutine: it moves the source where it was asked to,
// and reads it while the ring has room, unlocked, as either may take long.
func (p *pump) run() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for !p.closed {
		switch {
		case p.seeking:
			gen, target := p.gen, p.target
			p.mu.Unlock()
			err := p.src.SeekSample(target)
			pos := p.src.Position()
			p.mu.Lock()
			if gen != p.gen {
				// Moved again meanwhile: only the last one counts.
				continue
			}
			p.seeking, p.end = false, pos
			if err != nil {
				p.err = err
			}
		case p.err == nil && !p.eof && p.queued < pumpAhead:
			gen := p.gen
			p.mu.Unlock()
			pcm := make([]int16, pumpSize)
			from := p.src.Position()
			n, err := p.src.Read(pcm)
			to := p.src.Position()
			p.mu.Lock()
			if gen != p.gen {
				continue
			}
			if n > 0 {
				p.ring = append(p.ring, pumpChunk{pcm: pcm[:n], from: from, to: to})
				p.queued += n
			}
			switch {
			case errors.Is(err, io.EOF):
				p.eof = true
			case err != nil:
				p.err = err
			case n == 0:
				p.err = io.ErrNoProgress
			}
		default:
			p.cond.Wait()
		}
	}
}

// read takes the samples that are ready. While there are none, the source
// being moved or slow, it takes none: the output, which asks again a
// millisecond later, plays silence, which is no time of the sound, so that
// the position stays where it is.
func (p *pump) read(out []int16) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	n := 0
	for n < len(out) && len(p.ring) > 0 {
		c := &p.ring[0]
		k := copy(out[n:], c.pcm[c.taken:])
		n += k
		c.taken += k
		p.queued -= k
		if c.taken == len(c.pcm) {
			p.end = c.to
			p.ring = p.ring[:copy(p.ring, p.ring[1:])]
		}
	}
	if p.queued < pumpAhead {
		p.cond.Signal()
	}
	switch {
	case n > 0:
		return n, nil
	case p.err != nil && !p.seeking:
		return 0, p.err
	case p.eof && !p.seeking:
		return 0, io.EOF
	}
	return 0, nil
}

// seek moves the sound to sample pos, at once for whoever asks: what was read
// is dropped, and the source is moved by the pump, as soon as it is done
// with what it is doing.
func (p *pump) seek(pos int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.gen++
	p.ring, p.queued = p.ring[:0], 0
	p.seeking, p.target, p.end = true, pos, pos
	p.eof, p.err = false, nil
	p.cond.Signal()
}

// position is the sample the output takes next, or the one it will take
// when the source is done moving.
func (p *pump) position() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.ring) == 0 {
		return p.end
	}
	c := p.ring[0]
	return c.from + (c.to-c.from)*int64(c.taken)/int64(len(c.pcm))
}

// close stops the pump, which does not read the source after the read it is
// in.
func (p *pump) close() {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()
	p.cond.Broadcast()
}
