// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/gotd/td/tgerr"

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

type fakeConnection struct {
	failed     chan error
	reconnects chan struct{}
}

func (c *fakeConnection) FailConnection(err error)    { c.failed <- err }
func (c *fakeConnection) Reconnects() <-chan struct{} { return c.reconnects }

func TestKeepConnected(t *testing.T) {
	c := &fakeConnection{failed: make(chan error, 1), reconnects: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	broken := errors.New("history cache: disk full")
	var calls atomic.Int32
	done := make(chan struct{})
	go func() {
		defer close(done)
		keepConnected(ctx, func(ctx context.Context) error {
			if calls.Add(1) == 1 {
				return broken
			}
			<-ctx.Done()
			return ctx.Err()
		}, c)
	}()
	if err := <-c.failed; err != broken {
		t.Fatalf("told %v", err)
	}
	if calls.Load() != 1 {
		t.Fatal("connected again without being asked")
	}
	c.reconnects <- struct{}{}
	cancel()
	<-done
	if calls.Load() != 2 {
		t.Fatalf("connected %d times", calls.Load())
	}
	select {
	case err := <-c.failed:
		t.Fatalf("the window closing told as a failure: %v", err)
	default:
	}

	// An ended session is not a connection to make again.
	ended := tgerr.New(401, "SESSION_REVOKED")
	keepConnected(context.Background(), func(context.Context) error { return ended }, c)
	select {
	case err := <-c.failed:
		t.Fatalf("the ended session told as a failure: %v", err)
	default:
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
