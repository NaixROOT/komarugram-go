// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"slices"
	"testing"
	"time"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
)

// keepStore has the edits of messages.
type keepStore struct {
	*menuStore
}

func (s *keepStore) SetKeep(model.Keep) {}
func (s *keepStore) MessageEdits(_ context.Context, m model.Message) ([]model.Message, error) {
	first := m
	first.Text, first.EditedAt = "first", time.Time{}
	return []model.Message{first}, nil
}

// A message kept deleted takes no reply, forward or reaction, and says it
// was deleted; an edited one shows its versions.
func TestKeptMessages(t *testing.T) {
	h := newMenuHarnessOn(t, func(_ *menuStore, messages []model.Message) {
		messages[4].Deleted = true
		messages[5].EditedAt = time.Unix(2000, 0)
	}, func(s *menuStore) model.ConversationStore {
		return &keepStore{menuStore: s}
	})
	p := h.page
	deleted, _ := p.messageByID(5)
	if p.canReact(deleted) || p.canReply(deleted) {
		t.Fatal("a deleted message takes a reaction or a reply")
	}
	h.openMenu(5)
	for _, a := range []menuAction{actionReply, actionForward} {
		if slices.Contains(p.messageMenu.shown, a) {
			t.Fatalf("a deleted message offers %d", a)
		}
	}
	p.closeMenu()
	h.frames(30)
	h.openMenu(6)
	if !slices.Contains(p.messageMenu.shown, actionEdits) {
		t.Fatal("an edited message does not offer its edits")
	}
	h.choose(actionEdits)
	deadline := time.Now().Add(5 * time.Second)
	for !p.edits.loaded {
		if time.Now().After(deadline) {
			t.Fatal("the edits did not load")
		}
		time.Sleep(time.Millisecond)
		h.frame()
	}
	if len(p.edits.versions) != 1 || p.edits.versions[0].Text != "first" {
		t.Fatalf("versions %+v", p.edits.versions)
	}
	if got := localization.For("ru").T("history.deleted_mark"); got != "🧹" {
		t.Fatalf("the mark is %q", got)
	}
}

// translateStore translates.
type translateStore struct {
	*menuStore
	asked []model.MessageID
}

func (s *translateStore) Translate(_ context.Context, _ int64, id model.MessageID, text string, to string) (string, error) {
	s.asked = append(s.asked, id)
	return to + ": " + text, nil
}

// The menu translates a message into the interface's language.
func TestTranslateMenu(t *testing.T) {
	h := newMenuHarnessOn(t, nil, func(s *menuStore) model.ConversationStore {
		return &translateStore{menuStore: s}
	})
	p := h.page
	h.openMenu(4)
	if !slices.Contains(p.messageMenu.shown, actionTranslate) {
		t.Fatal("the menu does not translate")
	}
	h.choose(actionTranslate)
	deadline := time.Now().Add(5 * time.Second)
	for !p.translation.loaded {
		if time.Now().After(deadline) {
			t.Fatal("the translation did not come")
		}
		time.Sleep(time.Millisecond)
		h.frame()
	}
	if p.translation.result != "en: Message 4" {
		t.Fatalf("translated %q", p.translation.result)
	}
}
