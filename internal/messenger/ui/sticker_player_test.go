// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"testing"

	"gioui.org/layout"

	"komarugram/internal/messenger/localization"
)

func TestDecoderChoicesLayout(t *testing.T) {
	for _, language := range []string{"ru", "en"} {
		t.Run(language, func(t *testing.T) {
			s := newDecoderSettings()
			for _, c := range []*decoderChoice{s.stickers, s.animations, s.audio} {
				t.Run(c.title, func(t *testing.T) {
					h := &focusHarness{draw: func(gtx layout.Context) {
						c.radios.Update(gtx)
						c.Layout(gtx, localization.For(language))
					}}
					h.frame()
					h.frame()
				})
			}
		})
	}
}
