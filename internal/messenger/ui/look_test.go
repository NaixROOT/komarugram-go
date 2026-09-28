// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
	"komarugram/internal/messenger/preferences"
)

// The look of a frame rounds bubbles and avatars, writes seconds and puts
// its own marks.
func TestLook(t *testing.T) {
	gtx := layout.Context{Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Values: map[string]any{}}
	l := localization.For("en")
	at := time.Date(2026, 9, 25, 12, 34, 56, 0, time.Local)
	m := model.Message{Date: at, EditedAt: at, Deleted: true}
	if got := footerText(gtx, m, l); got != "edited · 🧹 · 12:34" {
		t.Fatalf("the default footer reads %q", got)
	}
	if got := bubbleRadiusOf(gtx); got != 16 {
		t.Fatalf("the default radius is %d", got)
	}
	if r := avatarShape(gtx, image.Pt(46, 46)); r.NW != 23 {
		t.Fatalf("the default avatar is not round: %+v", r)
	}
	withLook(gtx, preferences.Look{BubbleRadius: 4, AvatarCorners: 0, Seconds: true, EditedMark: "✎", DeletedMark: "✗"})
	if got := footerText(gtx, m, l); got != "✎ · ✗ · 12:34:56" {
		t.Fatalf("the footer reads %q", got)
	}
	if got := bubbleRadiusOf(gtx); got != 4 {
		t.Fatalf("the radius is %d", got)
	}
	if r := avatarShape(gtx, image.Pt(46, 46)); r.NW != 0 {
		t.Fatalf("the avatar is not square: %+v", r)
	}
	if s := shapeOf(gtx, joinAbove, false); s.nw != 4 || s.sw != 4 {
		t.Fatalf("a joined corner is rounder than the others: %+v", s)
	}
}
