package appwindow

import (
	"image"
	"strings"
	"testing"
	"time"

	"gioui.org/app"
	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"komarugram/internal/crash"
)

func TestPanicDialogRendersWithLongReport(t *testing.T) {
	window, err := headless.NewWindow(620, 320)
	if err != nil {
		t.Fatal(err)
	}
	defer window.Release()
	dialog := newPanicDialog(&Window{Window: new(app.Window)}, &crash.Panic{
		Where: "window account", Value: strings.Repeat("unexpected state ", 30),
		Stack: []byte("stack"), Path: "/tmp/" + strings.Repeat("long-directory/", 12) + "panic.txt",
	})
	ops := new(op.Ops)
	gtx := layout.Context{Ops: ops, Now: time.Now(), Constraints: layout.Exact(image.Pt(620, 320)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Values: map[string]any{}}
	dialog.Theme(gtx)
	dialog.Update(gtx)
	dialog.Layout(gtx)
	if err := window.Frame(ops); err != nil {
		t.Fatal(err)
	}
}

func TestPanicDialogDoesNotRepeatAfterIgnore(t *testing.T) {
	h := new(Host)
	if !h.beginPanicNotice("window") || h.beginPanicNotice("window") {
		t.Fatal("duplicate panic dialog opened")
	}
	h.ignorePanicNotice("window")
	if h.beginPanicNotice("window") {
		t.Fatal("ignored panic dialog reopened")
	}
	if !h.beginPanicNotice("another") {
		t.Fatal("different panic was ignored")
	}
}
