package historycache

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"komarugram/internal/messenger/model"
)

func searchIDs(t *testing.T, c *Cache, q Query) []model.MessageID {
	t.Helper()
	found, err := c.Search(context.Background(), q, 0, 50)
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]model.MessageID, len(found))
	for i, m := range found {
		ids[i] = m.Key.MessageID
	}
	return ids
}

func equalIDs(a []model.MessageID, b ...model.MessageID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSearchKeepsIndexWithMessages(t *testing.T) {
	ctx := context.Background()
	c, err := Open(filepath.Join(t.TempDir(), "history"), "a", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	day := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	msg := func(chat int64, id int, text string, kind model.MessageKind, age int) model.Message {
		return model.Message{Key: model.MessageKey{AccountID: "a", ChatID: chat, MessageID: model.MessageID(id)}, Text: text, Kind: kind, Date: day.Add(-time.Duration(age) * time.Hour)}
	}
	link := msg(2, 3, "смотри тут", model.MessageText, 1)
	link.Entities = []model.Entity{{Kind: "url", Offset: 0, Length: 6}}
	file := msg(2, 4, "", model.MessageFile, 0)
	file.Media = &model.MessageMedia{FileName: "Отчёт-2026.pdf"}
	err = c.SaveMessages(ctx, []model.Message{
		msg(1, 1, "Ёлка стоит в зале", model.MessageText, 3),
		msg(1, 2, "Привет, мир!", model.MessagePhoto, 2),
		link, file,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ids := searchIDs(t, c, Query{Text: "елк"}); !equalIDs(ids, 1) {
		t.Errorf("Ё is not found as Е: %v", ids)
	}
	if ids := searchIDs(t, c, Query{Text: "ПРИВ мир"}); !equalIDs(ids, 2) {
		t.Errorf("prefixes of every word: %v", ids)
	}
	if ids := searchIDs(t, c, Query{Text: "зала"}); !equalIDs(ids, 1) {
		t.Errorf("other forms of a word: %v", ids)
	}
	if ids := searchIDs(t, c, Query{Text: "отчет"}); !equalIDs(ids, 4) {
		t.Errorf("file names: %v", ids)
	}
	if ids := searchIDs(t, c, Query{}); !equalIDs(ids, 4, 3, 2, 1) {
		t.Errorf("newest first: %v", ids)
	}
	if ids := searchIDs(t, c, Query{Kinds: []model.MessageKind{model.MessagePhoto}}); !equalIDs(ids, 2) {
		t.Errorf("kinds: %v", ids)
	}
	if ids := searchIDs(t, c, Query{Links: true}); !equalIDs(ids, 3) {
		t.Errorf("links: %v", ids)
	}
	if ids := searchIDs(t, c, Query{Chats: []int64{2}}); !equalIDs(ids, 4, 3) {
		t.Errorf("chats: %v", ids)
	}
	if ids := searchIDs(t, c, Query{Text: `"*) OR`}); len(ids) != 0 {
		t.Errorf("FTS syntax leaks through: %v", ids)
	}

	// An edit replaces what is indexed, a deletion removes it.
	edited := msg(1, 1, "Сосна стоит в зале", model.MessageText, 3)
	if err := c.SaveMessages(ctx, []model.Message{edited}); err != nil {
		t.Fatal(err)
	}
	if ids := searchIDs(t, c, Query{Text: "елка"}); len(ids) != 0 {
		t.Errorf("an edit leaves the old text: %v", ids)
	}
	if ids := searchIDs(t, c, Query{Text: "сосна"}); !equalIDs(ids, 1) {
		t.Errorf("an edit is not indexed: %v", ids)
	}
	if err := c.Delete(ctx, 1, []int{1}); err != nil {
		t.Fatal(err)
	}
	if ids := searchIDs(t, c, Query{Text: "сосна"}); len(ids) != 0 {
		t.Errorf("a deleted message is found: %v", ids)
	}
}

func TestSearchIndexesExistingMessages(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "history")
	c, err := Open(path, "a", nil)
	if err != nil {
		t.Fatal(err)
	}
	var msgs []model.Message
	for i := 1; i <= searchBatch+10; i++ {
		msgs = append(msgs, model.Message{Key: model.MessageKey{AccountID: "a", ChatID: 1, MessageID: model.MessageID(i)}, Text: "старое сообщение"})
	}
	if err := c.SaveMessages(ctx, msgs); err != nil {
		t.Fatal(err)
	}
	c.Close()
	// A cache from before the index: no index, no triggers, no version.
	db, err := sql.Open("sqlite3", path+".plain")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DROP TABLE message_text; DROP TRIGGER message_text_insert; DROP TRIGGER message_text_update; DROP TRIGGER message_text_delete; DELETE FROM kv WHERE key LIKE 'search/%'`); err != nil {
		t.Fatal(err)
	}
	db.Close()

	c, err = Open(path, "a", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	deadline := time.Now().Add(5 * time.Second)
	for {
		found, err := c.Search(ctx, Query{Text: "старое"}, 0, searchBatch+100)
		if err != nil {
			t.Fatal(err)
		}
		if len(found) == len(msgs) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("indexed %d of %d existing messages", len(found), len(msgs))
		}
		time.Sleep(20 * time.Millisecond)
	}
}
