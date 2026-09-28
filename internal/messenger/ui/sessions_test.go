// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"image"
	"os"
	"testing"
	"time"

	"gioui.org/layout"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/mockstore"
	"komarugram/internal/messenger/model"
)

// loadedSessions is the view of the demo sessions, loaded.
func loadedSessions(t *testing.T) *sessionsView {
	t.Helper()
	v := newSessionsView(mockstore.New(time.Now(), 0), nil)
	v.refresh()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, _, _, loaded, _ := v.sessions(); loaded {
			return v
		}
		if time.Now().After(deadline) {
			t.Fatal("the sessions never loaded")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestSessionsGrouped(t *testing.T) {
	v := loadedSessions(t)
	current, others, incomplete, _, err := v.sessions()
	if err != nil || current == nil || !current.Current {
		t.Fatalf("this device %+v, %v", current, err)
	}
	if len(incomplete) != 1 || !incomplete[0].Incomplete {
		t.Errorf("incomplete %+v", incomplete)
	}
	for i := 1; i < len(others); i++ {
		if others[i].Active.After(others[i-1].Active) {
			t.Errorf("others not the last active first: %v", others)
		}
	}
	// The count leaves out the logins that never finished.
	if got := v.count(); got != 1+len(others) {
		t.Errorf("count %d, want %d", got, 1+len(others))
	}
}

func TestSessionPlace(t *testing.T) {
	l := localization.For("ru")
	now := time.Date(2026, 9, 25, 18, 0, 0, 0, time.Local)
	for _, c := range []struct {
		s    model.Session
		want string
	}{
		{model.Session{Current: true, Country: "Россия"}, "Россия · в сети"},
		{model.Session{Country: "Россия", Active: now.Add(-time.Hour)}, "Россия · 17:00"},
		{model.Session{IP: "192.0.2.1", Active: time.Date(2026, 1, 2, 3, 4, 0, 0, time.Local)}, "192.0.2.1 · 02.01.26"},
	} {
		if got := sessionPlace(c.s, now, l); got != c.want {
			t.Errorf("%+v: %q, want %q", c.s, got, c.want)
		}
	}
}

// TestRenderSessions saves the Devices section and a session's dialog, for
// looking at them: SESSIONS_PNG=/tmp/sessions.png.
func TestRenderSessions(t *testing.T) {
	path := os.Getenv("SESSIONS_PNG")
	if path == "" {
		t.Skip("set SESSIONS_PNG to a file")
	}
	v := loadedSessions(t)
	l := localization.For("ru")
	renderFrames(t, image.Pt(560, 820), path, func(gtx layout.Context) {
		layout.UniformInset(16).Layout(gtx, func(gtx layout.Context) layout.Dimensions { return v.Layout(gtx, l) })
	})
	sessions, _ := v.source.Sessions(context.Background())
	v.detail = sessions[1]
	v.dialog.Open()
	renderFrames(t, image.Pt(560, 820), path[:len(path)-4]+"-dialog.png", func(gtx layout.Context) {
		v.layoutDialog(gtx, l)
	})
}
