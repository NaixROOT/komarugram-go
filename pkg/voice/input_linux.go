// SPDX-License-Identifier: Unlicense OR MIT

package voice

import "context"

// inputs are PulseAudio's default source, which PipeWire serves too, then
// ALSA's.
func inputs(context.Context, string) ([][]string, error) {
	return [][]string{{"-f", "pulse", "-i", "default"}, {"-f", "alsa", "-i", "default"}}, nil
}
