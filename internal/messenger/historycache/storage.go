// SPDX-License-Identifier: Unlicense OR MIT

package historycache

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"komarugram/internal/messenger/model"
)

// Every media row has an object: its category, size and times, and the
// chats and messages that show it (media_refs). Rows cached before objects
// were recorded get theirs from the backfill, from the messages that name
// them; legacy marks those, whose references may be incomplete.
const storageSchema = `
 CREATE TABLE IF NOT EXISTS media_objects(key TEXT PRIMARY KEY, category INTEGER NOT NULL, size INTEGER NOT NULL, created INTEGER NOT NULL, accessed INTEGER NOT NULL, legacy INTEGER NOT NULL DEFAULT 0);
 CREATE TABLE IF NOT EXISTS media_refs(key TEXT, chat INTEGER, message INTEGER, PRIMARY KEY(key,chat,message)) WITHOUT ROWID;
 CREATE INDEX IF NOT EXISTS media_refs_chat ON media_refs(chat);
 CREATE TRIGGER IF NOT EXISTS media_forget AFTER DELETE ON media BEGIN
  DELETE FROM media_objects WHERE key=old.key;
  DELETE FROM media_refs WHERE key=old.key;
 END;`

// MediaRef is who a cached object is read or written for.
type MediaRef struct {
	Chat     int64
	Message  int
	Category model.StorageCategory
}

// RefOf is the reference of the media of m under key.
func RefOf(m model.Message, key string) MediaRef {
	r := MediaRef{Chat: m.Key.ChatID, Message: int(m.Key.MessageID), Category: model.StorageCategoryOf(m, key)}
	if model.IsProfilePhoto(m.Key.MessageID) {
		r.Message = 0
	}
	if chat, ok := model.AvatarChat(key); ok && r.Chat == 0 {
		r.Chat = chat
	}
	return r
}

type mediaRefRow struct {
	key     string
	chat    int64
	message int
}

// accessFlush is how long reads' access times and references wait in
// memory before they are written together.
const (
	accessFlush      = 30 * time.Second
	accessFlushCount = 256
)

type mediaAccess struct {
	at       int64
	category model.StorageCategory
}

// noteAccess remembers a read of key for ref; c.mu is held.
func (c *Cache) noteAccess(key string, ref MediaRef) {
	if c.accessed == nil {
		c.accessed, c.accessRefs = map[string]mediaAccess{}, map[mediaRefRow]struct{}{}
		c.flushed = c.clock()
	}
	c.accessed[key] = mediaAccess{c.clock().UnixNano(), ref.Category}
	if ref.Chat != 0 {
		c.accessRefs[mediaRefRow{key, ref.Chat, ref.Message}] = struct{}{}
	}
	if len(c.accessed) >= accessFlushCount || c.clock().Sub(c.flushed) >= accessFlush {
		if err := c.flushAccess(context.Background()); err != nil {
			log.Printf("historycache: media access: %v", err)
		}
	}
}

// flushAccess writes the remembered reads; c.mu is held. A read of a row
// without an object makes one, marked legacy: the backfill still looks for
// the other messages that show it.
func (c *Cache) flushAccess(ctx context.Context) error {
	c.flushed = c.clock()
	if len(c.accessed) == 0 && len(c.accessRefs) == 0 {
		return nil
	}
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for key, a := range c.accessed {
		if _, err = tx.ExecContext(ctx, `INSERT INTO media_objects(key,category,size,created,accessed,legacy) SELECT key,?,length(data),used,?,1 FROM media WHERE key=? ON CONFLICT(key) DO UPDATE SET accessed=max(accessed,excluded.accessed)`, a.category, a.at, key); err != nil {
			return err
		}
	}
	for r := range c.accessRefs {
		if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO media_refs SELECT ?,?,? WHERE EXISTS(SELECT 1 FROM media_objects WHERE key=?)`, r.key, r.chat, r.message, r.key); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	clear(c.accessed)
	clear(c.accessRefs)
	return nil
}

// saveObject records the object of a row just written; tx holds the write.
func (c *Cache) saveObject(ctx context.Context, tx *sql.Tx, key string, size int, ref MediaRef) error {
	now := c.clock().UnixNano()
	category := ref.Category
	if parent := parentKey(key); parent != "" {
		var p model.StorageCategory
		if err := tx.QueryRowContext(ctx, `SELECT category FROM media_objects WHERE key=?`, parent).Scan(&p); err == nil {
			category = p
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO media_objects(key,category,size,created,accessed) VALUES(?,?,?,?,?) ON CONFLICT(key) DO UPDATE SET size=excluded.size, accessed=excluded.accessed, category=CASE WHEN legacy THEN excluded.category ELSE category END`, key, category, size, now, now); err != nil {
		return err
	}
	if ref.Chat == 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO media_refs VALUES(?,?,?)`, key, ref.Chat, ref.Message)
	return err
}

// parentKey is the object key is a part of: a video's ranges, an avatar's
// video; "" for a whole object.
func parentKey(key string) string {
	if i := strings.Index(key, "/range/"); i > 0 {
		return key[:i]
	}
	if p, ok := strings.CutSuffix(key, "/video"); ok && strings.HasPrefix(key, "avatar/") {
		return p
	}
	return ""
}

// StorageUsage tells what the cache holds. It reads sizes and references,
// never the media, and does not count as access to anything.
func (c *Cache) StorageUsage(ctx context.Context) (model.CacheUsage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var u model.CacheUsage
	if err := c.flushAccess(ctx); err != nil {
		return u, err
	}
	// One pass over objects and their references, in key order: both
	// tables are kept by key, so this is a merge of the two, and much
	// cheaper than a query per figure.
	rows, err := c.db.QueryContext(ctx, `SELECT o.key, o.category, o.size, r.chat FROM media_objects o LEFT JOIN media_refs r ON r.key=o.key ORDER BY o.key`)
	if err != nil {
		return u, err
	}
	chats := map[int64]*model.ChatUsage{}
	var key string
	var cat int
	var size int64
	var seen []int64
	flush := func() {
		if key == "" {
			return
		}
		u.Categories[cat] += size
		u.Media += size
		switch {
		case len(seen) == 0:
			u.NoChat += size
		case len(seen) > 1:
			u.Shared += size
		}
		for _, chat := range seen {
			cu := chats[chat]
			if cu == nil {
				cu = &model.ChatUsage{Chat: chat}
				chats[chat] = cu
			}
			cu.Categories[cat] += size
			cu.Bytes += size
		}
	}
	for rows.Next() {
		var k string
		var category int
		var n int64
		var chat sql.NullInt64
		if err = rows.Scan(&k, &category, &n, &chat); err != nil {
			rows.Close()
			return u, err
		}
		if k != key {
			flush()
			key, cat, size, seen = k, categoryIndex(category), n, seen[:0]
		}
		if chat.Valid && !slices.Contains(seen, chat.Int64) {
			seen = append(seen, chat.Int64)
		}
	}
	flush()
	rows.Close()
	if err = rows.Err(); err != nil {
		return u, err
	}
	for _, cu := range chats {
		u.Chats = append(u.Chats, *cu)
	}
	sort.Slice(u.Chats, func(i, j int) bool {
		a, b := u.Chats[i], u.Chats[j]
		return a.Bytes > b.Bytes || a.Bytes == b.Bytes && a.Chat < b.Chat
	})
	var unattributed sql.NullInt64
	if err = c.db.QueryRowContext(ctx, `SELECT sum(length(data)) FROM media WHERE key IN (SELECT key FROM media EXCEPT SELECT key FROM media_objects)`).Scan(&unattributed); err != nil {
		return u, err
	}
	u.Unattributed = unattributed.Int64
	u.Media += u.Unattributed
	b, err := c.getLocked(backfillKey)
	if err != nil {
		return u, err
	}
	u.Backfilled = string(b) == backfillDone
	var pageSize, free int64
	if err = c.db.QueryRowContext(ctx, `SELECT page_size, freelist_count FROM pragma_page_size, pragma_freelist_count`).Scan(&pageSize, &free); err != nil {
		return u, err
	}
	u.Free = pageSize * free
	u.Database = c.fileSize()
	return u, nil
}

func categoryIndex(cat int) int {
	if cat < 0 || cat >= model.StorageCategories {
		return int(model.StorageOther)
	}
	return cat
}

// fileSize is the length of the database's files; c.mu is held.
func (c *Cache) fileSize() int64 {
	base := c.path + ".plain"
	if c.encrypted {
		base = c.path + ".secure"
	}
	var n int64
	for _, suffix := range []string{"", "-journal", "-wal"} {
		if st, err := os.Stat(base + suffix); err == nil {
			n += st.Size()
		}
	}
	return n
}

const (
	backfillKey  = "storage/backfill"
	backfillDone = "done"
)

// backfillBatch is how many messages one step reads.
var backfillBatch = 500

// backfillTables are scanned in this order; the state is "table:rowid".
var backfillTables = []string{"messages", "edits"}

// backfillMedia runs the backfill in batches, letting the history have the
// cache between them, as indexExisting does.
func (c *Cache) backfillMedia() {
	for {
		done, err := c.backfillStep()
		if err != nil {
			log.Printf("historycache: media backfill: %v", err)
			return
		}
		if done {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// backfillStep attributes media rows without a complete object from one
// batch of messages. Every write is idempotent and the position is saved
// with it, so that a backfill cut short goes on after a restart.
func (c *Cache) backfillStep() (done bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return true, nil
	}
	b, err := c.getLocked(backfillKey)
	if err != nil {
		return false, err
	}
	state := string(b)
	if state == backfillDone {
		return true, nil
	}
	if c.candidates == nil {
		if err = c.loadCandidates(); err != nil {
			return false, err
		}
	}
	table, pos, _ := strings.Cut(state, ":")
	if table == "" {
		table = backfillTables[0]
	}
	after, _ := strconv.ParseInt(pos, 10, 64)
	query := `SELECT rowid, chat, id, payload FROM messages WHERE rowid>? ORDER BY rowid LIMIT ?`
	if table == "edits" {
		query = `SELECT rowid, chat, id, payload FROM edits WHERE rowid>? ORDER BY rowid LIMIT ?`
	}
	rows, err := c.db.Query(query, after, backfillBatch)
	if err != nil {
		return false, err
	}
	type hit struct {
		key  string
		ref  MediaRef
		size int
	}
	var hits []hit
	last, n := after, 0
	for rows.Next() {
		var chat int64
		var id int
		var payload []byte
		if err = rows.Scan(&last, &chat, &id, &payload); err != nil {
			rows.Close()
			return false, err
		}
		n++
		kind, keys := mediaKeys(payload)
		msg := model.Message{Key: model.MessageKey{ChatID: chat, MessageID: model.MessageID(id)}, Kind: kind}
		for _, key := range keys {
			for _, k := range append([]string{key}, c.children[key]...) {
				if c.candidates[k] {
					hits = append(hits, hit{key: k, ref: RefOf(msg, key)})
				}
			}
		}
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return false, err
	}
	next := table + ":" + strconv.FormatInt(last, 10)
	if n < backfillBatch {
		next = ""
		for i, t := range backfillTables {
			if t == table && i+1 < len(backfillTables) {
				next = backfillTables[i+1] + ":0"
			}
		}
	}
	tx, err := c.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	for _, h := range hits {
		if _, err = tx.Exec(`INSERT OR IGNORE INTO media_objects(key,category,size,created,accessed,legacy) SELECT key,?,length(data),used,used,1 FROM media WHERE key=?`, h.ref.Category, h.key); err != nil {
			return false, err
		}
		if h.ref.Chat != 0 {
			if _, err = tx.Exec(`INSERT OR IGNORE INTO media_refs SELECT ?,?,? WHERE EXISTS(SELECT 1 FROM media_objects WHERE key=?)`, h.key, h.ref.Chat, h.ref.Message, h.key); err != nil {
				return false, err
			}
		}
	}
	if next == "" {
		// What no message names is attributed by its key where it can be.
		for k := range c.candidates {
			if chat, ok := model.AvatarChat(k); ok {
				if _, err = tx.Exec(`INSERT OR IGNORE INTO media_objects(key,category,size,created,accessed,legacy) SELECT key,?,length(data),used,used,1 FROM media WHERE key=?`, model.StorageProfilePhotos, k); err != nil {
					return false, err
				}
				if _, err = tx.Exec(`INSERT OR IGNORE INTO media_refs SELECT ?,?,0 WHERE EXISTS(SELECT 1 FROM media_objects WHERE key=?)`, k, chat, k); err != nil {
					return false, err
				}
			} else if strings.HasPrefix(k, "emoji/") {
				if _, err = tx.Exec(`INSERT OR IGNORE INTO media_objects(key,category,size,created,accessed,legacy) SELECT key,?,length(data),used,used,1 FROM media WHERE key=?`, model.StorageStickers, k); err != nil {
					return false, err
				}
			}
		}
		next = backfillDone
	}
	if _, err = tx.Exec(`INSERT OR REPLACE INTO kv VALUES(?,?)`, backfillKey, []byte(next)); err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	if next == backfillDone {
		c.candidates, c.children = nil, nil
	}
	return next == backfillDone, nil
}

// loadCandidates reads the keys the backfill looks for: rows without an
// object, and legacy objects; c.mu is held.
func (c *Cache) loadCandidates() error {
	rows, err := c.db.Query(`SELECT key FROM media WHERE NOT EXISTS(SELECT 1 FROM media_objects o WHERE o.key=media.key) UNION SELECT key FROM media_objects WHERE legacy`)
	if err != nil {
		return err
	}
	defer rows.Close()
	c.candidates, c.children = map[string]bool{}, map[string][]string{}
	for rows.Next() {
		var k string
		if err = rows.Scan(&k); err != nil {
			return err
		}
		c.candidates[k] = true
		if p := parentKey(k); p != "" {
			c.children[p] = append(c.children[p], k)
		}
	}
	return rows.Err()
}

// mediaKeys are the message's kind and the keys of all media its payload
// names: its media, thumbnails, variants, and those of a link preview or a
// gift.
func mediaKeys(payload []byte) (kind model.MessageKind, keys []string) {
	var v any
	if json.Unmarshal(payload, &v) != nil {
		return 0, nil
	}
	if m, ok := v.(map[string]any); ok {
		if k, ok := m["Kind"].(float64); ok {
			kind = model.MessageKind(k)
		}
	}
	var walk func(v any, inMedia bool)
	walk = func(v any, inMedia bool) {
		switch v := v.(type) {
		case map[string]any:
			for k, x := range v {
				if s, ok := x.(string); ok && k == "ID" && inMedia && s != "" {
					keys = append(keys, s)
					continue
				}
				walk(x, inMedia || k == "Media" || k == "WebPage" || k == "Gift")
			}
		case []any:
			for _, x := range v {
				walk(x, inMedia)
			}
		}
	}
	walk(v, false)
	return kind, keys
}
