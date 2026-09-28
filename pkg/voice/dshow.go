// SPDX-License-Identifier: Unlicense OR MIT

package voice

import (
	"regexp"
	"strings"
)

// dshowDevice is a line of ffmpeg's list of DirectShow devices:
// `[dshow @ 0x…] "Microphone (USB Audio)" (audio)`.
var dshowDevice = regexp.MustCompile(`"([^"]+)" \(audio\)`)

// dshowAudio are the audio devices in ffmpeg's list of DirectShow devices,
// in its order.
func dshowAudio(list string) []string {
	var names []string
	for _, line := range strings.Split(list, "\n") {
		if m := dshowDevice.FindStringSubmatch(line); m != nil {
			names = append(names, m[1])
		}
	}
	return names
}
