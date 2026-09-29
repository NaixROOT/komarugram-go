// SPDX-License-Identifier: Unlicense OR MIT

package appwindow

import "testing"

type recoveringContent struct {
	Content
	recover func() bool
}

func (c recoveringContent) RecoverFrame() bool { return c.recover() }

func TestRecoverContent(t *testing.T) {
	if recoverContent(nil) {
		t.Fatal("content without recovery must keep the fallback")
	}
	called := false
	if !recoverContent(recoveringContent{recover: func() bool {
		called = true
		return true
	}}) || !called {
		t.Fatal("content recovery was not called")
	}
	if recoverContent(recoveringContent{recover: func() bool { return false }}) {
		t.Fatal("declined recovery must keep the fallback")
	}
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	if recoverContent(recoveringContent{recover: func() bool { panic("broken recovery") }}) {
		t.Fatal("failed recovery must keep the fallback")
	}
}
