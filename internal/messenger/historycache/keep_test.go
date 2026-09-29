// SPDX-License-Identifier: Unlicense OR MIT

package historycache

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"komarugram/internal/messenger/model"
)

func keepCache(t *testing.T) *Cache {
	t.Helper()
	c, err := Open(filepath.Join(t.TempDir(), "h"), "a", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	c.SetKeep(true, true)
	return c
}

func msg(chat int64, id int, text string) model.Message {
	return model.Message{Key: model.MessageKey{AccountID: "a", ChatID: chat, MessageID: model.MessageID(id)}, Text: text, Date: time.Unix(100, 0)}
}

// A message Telegram deletes stays, marked deleted, in the history but not
// in a chat skip names; a late page does not bring its old version back.
func TestKeepDeleted(t *testing.T) {
	c := keepCache(t)
	ctx := context.Background()
	if err := c.SaveMessages(ctx, []model.Message{msg(1, 5, "stays"), msg(1, 6, "next"), msg(2, 7, "bot's")}); err != nil {
		t.Fatal(err)
	}
	kept, err := c.MarkDeleted(ctx, 0, []int{5, 7}, func(chat int64) bool { return chat == 2 })
	if err != nil || len(kept) != 1 || kept[0].Key.MessageID != 5 || !kept[0].Deleted {
		t.Fatalf("kept %+v, %v", kept, err)
	}
	if err := c.SaveMessages(ctx, []model.Message{msg(1, 5, "stays")}); err != nil {
		t.Fatal(err)
	}
	got, err := c.Around(ctx, 1, 0, 10)
	if err != nil || len(got) != 2 || !got[0].Deleted || got[0].Text != "stays" {
		t.Fatalf("the history is %+v, %v", got, err)
	}
	if got, _ := c.Around(ctx, 2, 0, 10); len(got) != 0 {
		t.Fatalf("a bot's deleted message stayed: %+v", got)
	}
	// What a page of Telegram lacks is kept too.
	removed, kept, err := c.Reconcile(ctx, 1, 1, 10, []int{5})
	if err != nil || len(removed) != 0 || len(kept) != 1 || kept[0].Key.MessageID != 6 {
		t.Fatalf("reconciled %v, kept %+v, %v", removed, kept, err)
	}
	// Not keeping them, a deletion deletes.
	c.SetKeep(false, false)
	if err := c.SaveMessages(ctx, []model.Message{msg(1, 8, "goes")}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.MarkDeleted(ctx, 1, []int{8}, nil); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := c.Message(ctx, 1, 8); ok {
		t.Fatal("a message deleted without keeping stayed")
	}
}

// An edit of someone's message keeps the text it had; the account's own
// messages and edits that keep the text keep nothing.
func TestKeepEdits(t *testing.T) {
	c := keepCache(t)
	ctx := context.Background()
	first := msg(1, 5, "first")
	own := msg(1, 6, "mine")
	own.Outgoing = true
	if err := c.SaveMessages(ctx, []model.Message{first, own}); err != nil {
		t.Fatal(err)
	}
	second := first
	second.Text, second.EditedAt = "second", time.Unix(200, 0)
	third := first
	third.Text, third.EditedAt = "third", time.Unix(300, 0)
	sameText := third
	sameText.EditedAt = time.Unix(400, 0)
	ownEdit := own
	ownEdit.Text, ownEdit.EditedAt = "mine, edited", time.Unix(200, 0)
	for _, m := range []model.Message{second, second, third, sameText, ownEdit} {
		if err := c.SaveMessages(ctx, []model.Message{m}); err != nil {
			t.Fatal(err)
		}
	}
	edits, err := c.Edits(ctx, 1, 5)
	if err != nil || len(edits) != 2 || edits[0].Text != "first" || edits[1].Text != "second" {
		t.Fatalf("edits %+v, %v", edits, err)
	}
	if edits, _ := c.Edits(ctx, 1, 6); len(edits) != 0 {
		t.Fatalf("own edits kept: %+v", edits)
	}
}
