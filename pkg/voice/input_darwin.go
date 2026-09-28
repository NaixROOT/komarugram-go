// SPDX-License-Identifier: Unlicense OR MIT

package voice

import "context"

// inputs are AVFoundation's default microphone, then its first.
func inputs(context.Context, string) ([][]string, error) {
	return [][]string{{"-f", "avfoundation", "-i", ":default"}, {"-f", "avfoundation", "-i", ":0"}}, nil
}
