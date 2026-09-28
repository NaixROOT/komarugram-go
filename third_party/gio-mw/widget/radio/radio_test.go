// SPDX-License-Identifier: Unlicense OR MIT

package radio

import (
	"image"
	"testing"

	"gio-mw/defaults"
	"gio-mw/defaults/schemes"
	"gio-mw/wdk"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// TestLabelClickSelects verifies that clicking a radio's label, not only its
// icon, selects the option.
func TestLabelClickSelects(t *testing.T) {
	for _, kind := range []Kind{LeadingKind, TrailingKind} {
		var r input.Router
		options := []string{"a", "b"}
		labels := map[string]string{"a": "Option A", "b": "Option B"}
		radios := NewRadios(options, "a", func(string) {})

		var rowHeight, rowWidth int
		frame := func() {
			gtx := layout.Context{
				Ops:         new(op.Ops),
				Source:      r.Source(),
				Constraints: layout.Exact(image.Pt(400, 400)),
				Metric:      unit1(),
				Values:      map[string]any{},
			}
			wdk.InitMaterialThemeInContext(gtx, defaults.NewTheme(gtx, schemes.SchemeBaselineLight()))
			radios.Update(gtx)
			dims := radios.Layout(gtx, kind, labels)
			rowHeight = dims.Size.Y / len(options)
			rowWidth = dims.Size.X
			r.Frame(gtx.Ops)
		}
		frame()

		// Click in the label area of the second row, away from the icon.
		labelX := float32(rowWidth / 2)
		pos := f32.Pt(labelX, float32(rowHeight)+float32(rowHeight)/2)
		r.Queue(
			pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: pos},
			pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: pos},
		)
		frame()

		if got := radios.GetValue(); got != "b" {
			t.Errorf("kind %d: clicking label: value = %q, want %q", kind, got, "b")
		}
	}
}

func unit1() unit.Metric { return unit.Metric{PxPerDp: 1, PxPerSp: 1} }
