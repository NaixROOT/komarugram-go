// SPDX-License-Identifier: Unlicense OR MIT

package chatmedia

import (
	"testing"

	"komarugram/pkg/video"
)

func TestSandboxGIFsWhenChosen(t *testing.T) {
	if video.ResolveFFmpeg("") == "" {
		t.Skip("no FFmpeg")
	}
	m := New(nil, func() {})
	defer m.Close()
	m.ConfigureDecoders(false, false, "")
	if m.SandboxGIFs() {
		t.Fatal("GIFs play in the sandbox with an FFmpeg")
	}
	m.ConfigureDecoders(false, true, "")
	if !m.SandboxGIFs() {
		t.Fatal("GIFs do not play in the sandbox chosen")
	}
}
