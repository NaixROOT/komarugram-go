// SPDX-License-Identifier: Unlicense OR MIT

package audio

import (
	"encoding/binary"
	"math"
)

// surround is what a channel of 5.1 sound, in the order of WAV files (front
// left and right, centre, low frequencies, back left and right), weighs in
// the left and the right of the stereo it is folded to, as ITU-R BS.775
// has it, over the sum of the weights, so that nothing is louder than it was.
var surround = func() [2][6]float64 {
	k := math.Sqrt(0.5)
	sum := 1 + 2*k
	return [2][6]float64{
		{1 / sum, 0, k / sum, 0, k / sum, 0},
		{0, 1 / sum, k / sum, 0, 0, k / sum},
	}
}()

// Downmix takes the frames of raw, 16-bit samples of channels channels
// one frame after the other, little endian, to stereo: mono is in both
// channels, stereo is as it is, 5.1 is folded, and of any other layout the
// first two channels are kept. It returns the frames in dst, which is as
// long as they are.
func Downmix(dst []Frame, raw []byte, channels int) []Frame {
	channels = max(1, channels)
	frames := len(raw) / (2 * channels)
	dst = dst[:min(len(dst), frames)]
	sample := func(f, c int) int { return int(int16(binary.LittleEndian.Uint16(raw[2*(f*channels+c):]))) }
	for f := range dst {
		switch {
		case channels == 1:
			dst[f] = Dual(int16(sample(f, 0)))
		case channels == 6:
			var l, r float64
			for c := range 6 {
				v := float64(sample(f, c))
				l += v * surround[0][c]
				r += v * surround[1][c]
			}
			dst[f] = Frame{clip(int(math.Round(l))), clip(int(math.Round(r)))}
		default:
			dst[f] = Frame{int16(sample(f, 0)), int16(sample(f, 1))}
		}
	}
	return dst
}
