// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"context"
	"sync/atomic"
	"testing"

	"komarugram/internal/appwindow"
	"komarugram/internal/messenger/preferences"
)

type fakeTray struct{ available bool }

func (t *fakeTray) Available() bool { return t.available }

func TestKeepInBackground(t *testing.T) {
	h := newAccountWindows(nil, appwindow.Options{}, nil, nil, preferences.Memory(), nil, nil)
	if h.keepInBackground("a") {
		t.Fatal("kept without a tray")
	}
	icon := &fakeTray{available: true}
	h.tray = icon
	if !h.keepInBackground("a") {
		t.Fatal("account not kept with the icon shown")
	}
	if h.keepInBackground("") {
		t.Fatal("a window without an account kept its session")
	}
	icon.available = false
	if h.keepInBackground("a") {
		t.Fatal("kept although no panel shows the icon")
	}
	icon.available = true
	h.leaving["a"] = true
	if h.keepInBackground("a") {
		t.Fatal("an account being logged out of kept running")
	}
	h.Quit()
	if h.keepInBackground("b") {
		t.Fatal("kept while quitting")
	}
}

func TestSessionStopsOnce(t *testing.T) {
	holds := 0
	s := newSession(func() { holds++ }, func(*accountSession) {})
	var finished atomic.Bool
	s.start(func(ctx context.Context) {
		<-ctx.Done()
		finished.Store(true)
	})
	s.stop()
	if !finished.Load() {
		t.Fatal("stop returned before the worker")
	}
	s.stop()
	if holds != 1 {
		t.Fatalf("hold released %d times", holds)
	}
}
