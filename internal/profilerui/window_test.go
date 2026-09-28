package profilerui

import (
	"image"
	"testing"
	"time"

	"komarugram/internal/diagnostics"

	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestProfilerTabsLayout(t *testing.T) {
	r := diagnostics.Enable()
	defer r.Close()
	r.RecordFrame(diagnostics.Frame{At: time.Now(), Window: "demo", Total: 20 * time.Millisecond, Layout: 18 * time.Millisecond, Phases: []diagnostics.Phase{{Name: "history.rich-text", Duration: 10 * time.Millisecond, Calls: 8}}})
	r.History(diagnostics.History{Window: "demo", Chat: 1, Messages: 300, Visible: 8, Restore: "linear-id-anchor", RestoreScanned: 290})
	r.Event("demo", "history.restore", "anchor:290", time.Millisecond, 290)
	v := Spec(r, t.TempDir()).Build(nil).(*View)
	for tab := 0; tab < 4; tab++ {
		v.tab = tab
		gtx := layout.Context{Ops: new(op.Ops), Now: time.Now(), Constraints: layout.Exact(image.Pt(1150, 820)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Values: map[string]any{}}
		wdk.InitMaterialThemeInContext(gtx, v.Theme(gtx))
		v.Update(gtx)
		v.Layout(gtx)
		if len(v.rows()) == 0 {
			t.Fatal("empty profiler tab", tab)
		}
	}
}
