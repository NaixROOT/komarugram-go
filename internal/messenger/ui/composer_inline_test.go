// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"testing"
	"time"

	"gioui.org/io/key"
)

// "@wiki go" asks the bot once typing pauses, shows its articles over the
// field, and a click sends the one chosen.
func TestComposerInlineBot(t *testing.T) {
	h := newComposerHarness(t)
	c := h.p.composer
	h.click(150, 680)
	h.router.Queue(key.EditEvent{Text: "@wiki go"}, key.SelectionEvent{Start: 8, End: 8})
	wait := func(what string, cond func() bool) {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for h.frame(); !cond(); h.frame() {
			if time.Now().After(deadline) {
				t.Fatalf("timed out waiting for %s", what)
			}
			h.now = h.now.Add(50 * time.Millisecond)
			time.Sleep(5 * time.Millisecond)
		}
	}
	wait("the results", func() bool { return len(c.inline.results.Results) > 0 })
	if c.inline.asked != "go" || c.inline.results.Gallery {
		t.Fatalf("asked %q, gallery %v", c.inline.asked, c.inline.results.Gallery)
	}
	first := c.inline.results.Results[0]
	if first.Title != "Go" {
		t.Fatalf("first result %+v", first)
	}
	if c.inline.area.Empty() {
		t.Fatal("no answers drawn")
	}
	c.inline.clicks[0].click.Click()
	wait("the message", func() bool {
		hist := h.p.source.History(1)
		return hist.Messages[len(hist.Messages)-1].Text == "Go — компилируемый язык программирования от Google."
	})
	if got := c.draft(1).editor.Text(); got != "" {
		t.Fatalf("the field kept %q", got)
	}
}

// A gallery scrolled to its end asks for the bot's next page, until there
// is none; "@gif " shows the bot's placeholder.
func TestComposerInlinePagesAndPlaceholder(t *testing.T) {
	h := newComposerHarness(t)
	c := h.p.composer
	h.click(150, 680)
	h.router.Queue(key.EditEvent{Text: "@gif "}, key.SelectionEvent{Start: 5, End: 5})
	deadline := time.Now().Add(3 * time.Second)
	for h.frame(); len(c.inline.results.Results) < 24; h.frame() {
		if time.Now().After(deadline) {
			t.Fatalf("%d answers, next %q", len(c.inline.results.Results), c.inline.results.Next)
		}
		h.now = h.now.Add(50 * time.Millisecond)
		c.inline.list.Position.First = 1000
		time.Sleep(5 * time.Millisecond)
	}
	if c.inline.results.Next != "" || !c.inline.botOK || c.inline.bot.Placeholder == "" {
		t.Fatalf("next %q, bot %+v", c.inline.results.Next, c.inline.bot)
	}
}
