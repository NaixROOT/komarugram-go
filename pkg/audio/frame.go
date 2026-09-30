// SPDX-License-Identifier: Unlicense OR MIT

package audio

// Frame is one sample of each channel of stereo sound, left then right.
// What a Source puts out is frames, and its positions and lengths count
// them.
type Frame [2]int16

// Mono is the frame as one sample: the mean of the two channels.
func (f Frame) Mono() int16 { return int16((int(f[0]) + int(f[1])) / 2) }

// Dual is sample v in both channels, as a mono sound is played.
func Dual(v int16) Frame { return Frame{v, v} }

// clip is v as a sample, the nearest one there is for a value out of range.
func clip(v int) int16 { return int16(min(max(v, -32768), 32767)) }
