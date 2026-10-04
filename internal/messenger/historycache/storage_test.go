package historycache

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"komarugram/internal/messenger/model"
	"komarugram/internal/messenger/security"
)

// eachMode runs f on a plaintext cache and on an encrypted one.
func eachMode(t *testing.T, f func(t *testing.T, open func() *Cache)) {
	for _, encrypted := range []bool{false, true} {
		name := "plain"
		if encrypted {
			name = "adiantum"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			var p *security.Manager
			if encrypted {
				var e error
				if p, e = security.OpenPath(filepath.Join(dir, "security"), &fakeTPM{}); e != nil {
					t.Fatal(e)
				}
				if e = p.Enable(context.Background(), "test"); e != nil {
					t.Fatal(e)
				}
			}
			f(t, func() *Cache {
				c, e := Open(filepath.Join(dir, "history"), "a", p)
				if e != nil {
					t.Fatal(e)
				}
				return c
			})
		})
	}
}

func bytesOf(n int) []byte { return make([]byte, n) }

func TestStorageUsageReconciles(t *testing.T) {
	eachMode(t, func(t *testing.T, open func() *Cache) {
		ctx := context.Background()
		c := open()
		defer c.Close()
		save := func(key string, n int, ref MediaRef) {
			t.Helper()
			if e := c.SaveMedia(ctx, key, bytesOf(n), ref); e != nil {
				t.Fatal(e)
			}
		}
		save("photo/1/y", 1000, MediaRef{Chat: 1, Message: 10, Category: model.StoragePhotos})
		save("document/2", 3000, MediaRef{Chat: 1, Message: 11, Category: model.StorageVideos})
		save("document/3/range/0", 500, MediaRef{Chat: 2, Message: 5, Category: model.StorageVideos})
		save("document/3/range/131072", 500, MediaRef{Chat: 2, Message: 5, Category: model.StorageVideos})
		save("document/4", 2000, MediaRef{Chat: 1, Message: 12, Category: model.StorageGIFs})
		avatar := RefOf(model.Message{Kind: model.MessagePhoto}, "avatar/5/9")
		save("avatar/5/9", 100, avatar)
		save("emoji/77", 50, MediaRef{Category: model.StorageStickers})
		save("document/6", 400, MediaRef{Chat: 3, Message: 1, Category: model.StorageMusic})
		save("document/6/range/0", 40, MediaRef{Chat: 3, Message: 1})
		// The same GIF shown in a second chat, read from the cache there.
		if b, e := c.Media(ctx, "document/4", MediaRef{Chat: 2, Message: 7, Category: model.StorageGIFs}); e != nil || len(b) != 2000 {
			t.Fatal(len(b), e)
		}
		if _, e := c.db.Exec(`INSERT INTO media VALUES('legacy/1', ?, 1)`, bytesOf(700)); e != nil {
			t.Fatal(e)
		}

		u, e := c.StorageUsage(ctx)
		if e != nil {
			t.Fatal(e)
		}
		if u.Media != 1000+3000+500+500+2000+100+50+400+40+700 {
			t.Error("media", u.Media)
		}
		want := map[model.StorageCategory]int64{
			model.StoragePhotos: 1000, model.StorageVideos: 4000, model.StorageGIFs: 2000,
			model.StorageProfilePhotos: 100, model.StorageStickers: 50, model.StorageMusic: 440,
		}
		var sum int64
		for i, n := range u.Categories {
			sum += n
			if n != want[model.StorageCategory(i)] {
				t.Errorf("category %d: %d, want %d", i, n, want[model.StorageCategory(i)])
			}
		}
		if u.Unattributed != 700 || sum+u.Unattributed != u.Media {
			t.Error("unattributed", u.Unattributed, sum)
		}
		chats := map[int64]int64{}
		for _, cu := range u.Chats {
			chats[cu.Chat] = cu.Bytes
		}
		if len(chats) != 4 || chats[1] != 6000 || chats[2] != 3000 || chats[3] != 440 || chats[5] != 100 {
			t.Error("chats", u.Chats)
		}
		if u.Chats[0].Chat != 1 || u.Chats[0].Categories[model.StorageGIFs] != 2000 {
			t.Error("order", u.Chats[0])
		}
		if u.Shared != 2000 || u.NoChat != 50 || u.Database <= 0 {
			t.Error("shared", u.Shared, "no chat", u.NoChat, "database", u.Database)
		}

		if _, e = c.db.Exec(`DELETE FROM media WHERE key IN ('document/4','legacy/1')`); e != nil {
			t.Fatal(e)
		}
		var objects, refs int
		if e = c.db.QueryRow(`SELECT (SELECT count(*) FROM media_objects WHERE key='document/4'), (SELECT count(*) FROM media_refs WHERE key='document/4')`).Scan(&objects, &refs); e != nil || objects+refs != 0 {
			t.Fatal("removed media left", objects, refs, e)
		}
		if u, e = c.StorageUsage(ctx); e != nil || u.Media != 5590 || u.Shared != 0 || u.Unattributed != 0 {
			t.Fatal(u, e)
		}
	})
}

func TestMediaAccessIsBatched(t *testing.T) {
	eachMode(t, func(t *testing.T, open func() *Cache) {
		ctx := context.Background()
		c := open()
		defer c.Close()
		now := time.Unix(1000, 0)
		c.mu.Lock()
		c.clock = func() time.Time { return now }
		c.mu.Unlock()
		if e := c.SaveMedia(ctx, "photo/1/y", bytesOf(10), MediaRef{Chat: 1, Message: 1, Category: model.StoragePhotos}); e != nil {
			t.Fatal(e)
		}
		accessed := func() int64 {
			var at int64
			if e := c.db.QueryRow(`SELECT accessed FROM media_objects WHERE key='photo/1/y'`).Scan(&at); e != nil {
				t.Fatal(e)
			}
			return at
		}
		written := accessed()
		now = now.Add(time.Second)
		if _, e := c.Media(ctx, "photo/1/y", MediaRef{Chat: 1, Message: 1}); e != nil {
			t.Fatal(e)
		}
		if accessed() != written {
			t.Fatal("a read wrote at once")
		}
		now = now.Add(accessFlush)
		if _, e := c.Media(ctx, "photo/1/y", MediaRef{Chat: 1, Message: 1}); e != nil {
			t.Fatal(e)
		}
		if accessed() != now.UnixNano() {
			t.Fatal("reads not written after", accessFlush)
		}
		now = now.Add(time.Hour)
		if _, e := c.StorageUsage(ctx); e != nil {
			t.Fatal(e)
		}
		if accessed() == now.UnixNano() {
			t.Fatal("statistics counted as access")
		}
	})
}

func TestBackfillAttributesLegacyMediaAcrossRestart(t *testing.T) {
	defer func(n int) { backfillBatch = n }(backfillBatch)
	backfillBatch = 1
	eachMode(t, func(t *testing.T, open func() *Cache) {
		ctx := context.Background()
		c := open()
		waitBackfill(t, c)
		msg := func(chat int64, id int, kind model.MessageKind, media *model.MessageMedia) model.Message {
			return model.Message{Key: model.MessageKey{AccountID: "a", ChatID: chat, MessageID: model.MessageID(id)}, Kind: kind, Media: media}
		}
		gif := &model.MessageMedia{ID: "document/2", Thumbnail: &model.MessageMedia{ID: "document/2/thumb"}}
		if e := c.SaveMessages(ctx, []model.Message{
			msg(1, 1, model.MessagePhoto, &model.MessageMedia{ID: "photo/1/y", Variants: []model.MessageMedia{{ID: "photo/1/m"}}}),
			msg(1, 2, model.MessageGIF, gif),
			msg(2, 3, model.MessageVideo, &model.MessageMedia{ID: "document/3"}),
			msg(3, 4, model.MessageGIF, gif),
			msg(4, 5, model.MessageText, nil),
		}); e != nil {
			t.Fatal(e)
		}
		edited, _ := json.Marshal(msg(4, 5, model.MessagePhoto, &model.MessageMedia{ID: "photo/8/x"}))
		if _, e := c.db.Exec(`INSERT INTO edits VALUES(4,5,1,?)`, edited); e != nil {
			t.Fatal(e)
		}
		legacy := map[string]int{"photo/1/m": 10, "document/2/thumb": 20, "document/3/range/0": 30, "photo/8/x": 40, "avatar/9/1": 50, "orphan": 60}
		for k, n := range legacy {
			if _, e := c.db.Exec(`INSERT INTO media VALUES(?,?,1)`, k, bytesOf(n)); e != nil {
				t.Fatal(e)
			}
		}
		c.mu.Lock()
		if _, e := c.db.Exec(`DELETE FROM kv WHERE key=?`, backfillKey); e != nil {
			t.Fatal(e)
		}
		c.candidates = nil
		for range 3 {
			c.mu.Unlock()
			if _, e := c.backfillStep(); e != nil {
				t.Fatal(e)
			}
			c.mu.Lock()
		}
		c.mu.Unlock()
		if u, e := c.StorageUsage(ctx); e != nil || u.Backfilled {
			t.Fatal("backfill not cut short", e)
		}
		c.Close()

		c = open()
		defer c.Close()
		waitBackfill(t, c)
		u, e := c.StorageUsage(ctx)
		if e != nil {
			t.Fatal(e)
		}
		if !u.Backfilled || u.Unattributed != 60 || u.Media != 210 {
			t.Fatal("backfill", u.Backfilled, u.Unattributed, u.Media)
		}
		want := map[model.StorageCategory]int64{model.StoragePhotos: 50, model.StorageGIFs: 20, model.StorageVideos: 30, model.StorageProfilePhotos: 50}
		for cat, n := range want {
			if u.Categories[cat] != n {
				t.Error("category", cat, u.Categories[cat], "want", n)
			}
		}
		chats := map[int64]int64{}
		for _, cu := range u.Chats {
			chats[cu.Chat] = cu.Bytes
		}
		if chats[1] != 30 || chats[2] != 30 || chats[3] != 20 || chats[4] != 40 || chats[9] != 50 || u.Shared != 20 {
			t.Fatal("chats", chats, "shared", u.Shared)
		}
	})
}

func waitBackfill(t *testing.T, c *Cache) {
	t.Helper()
	for {
		done, e := c.backfillStep()
		if e != nil {
			t.Fatal(e)
		}
		if done {
			return
		}
	}
}
