// SPDX-License-Identifier: Unlicense OR MIT

package historycache

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/ncruces/go-sqlite3"
	"github.com/ncruces/go-sqlite3/ext/fts5"

	"komarugram/internal/messenger/model"
)

// Every connection gets FTS5: the triggers that keep the index write to it
// from any statement that changes messages, including those that copy the
// database when its protection changes.
func init() {
	sqlite3.AutoExtension(fts5.Register)
}

// searchVersion changes whenever what is indexed does, so that the index is
// built again.
const searchVersion = 1

// searchBatch is how many messages one step of building the index reads,
// so that the cache is not held for long.
const searchBatch = 2000

// searchText is what is indexed of a payload: the text, the names of files
// and music, and the link preview's title, site and address. Ё is indexed
// as Е, which Russian text uses for it interchangeably.
func searchText(payload string) string {
	p := "CAST(" + payload + " AS TEXT)"
	fields := []string{"$.Text", "$.Media.FileName", "$.Media.Title", "$.Media.Performer", "$.WebPage.Title", "$.WebPage.Site", "$.WebPage.URL"}
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = fmt.Sprintf("json_extract(%s,'%s')", p, f)
	}
	return "replace(replace(concat_ws(' '," + strings.Join(parts, ",") + "),'ё','е'),'Ё','Е')"
}

// initSearch makes the full-text index of messages and the triggers that
// keep it. The index keeps no text of its own: a match is read back from
// messages by rowid.
func (c *Cache) initSearch() error {
	var version int
	if b, err := c.getLocked("search/version"); err != nil {
		return err
	} else if b != nil {
		version, _ = strconv.Atoi(string(b))
	}
	if version != searchVersion {
		if _, err := c.db.Exec(`DROP TABLE IF EXISTS message_text;
 DROP TRIGGER IF EXISTS message_text_insert;
 DROP TRIGGER IF EXISTS message_text_update;
 DROP TRIGGER IF EXISTS message_text_delete;
 DELETE FROM kv WHERE key='search/rowid'`); err != nil {
			return err
		}
	}
	_, err := c.db.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS message_text USING fts5(body, content='', contentless_delete=1, tokenize='unicode61 remove_diacritics 2');
 CREATE TRIGGER IF NOT EXISTS message_text_insert AFTER INSERT ON messages WHEN new.payload IS NOT NULL AND new.deleted=0 BEGIN
  INSERT OR REPLACE INTO message_text(rowid, body) VALUES (new.rowid, ` + searchText("new.payload") + `);
 END;
 CREATE TRIGGER IF NOT EXISTS message_text_update AFTER UPDATE OF payload, deleted ON messages BEGIN
  DELETE FROM message_text WHERE rowid=old.rowid;
  INSERT INTO message_text(rowid, body) SELECT new.rowid, ` + searchText("new.payload") + ` WHERE new.payload IS NOT NULL AND new.deleted=0;
 END;
 CREATE TRIGGER IF NOT EXISTS message_text_delete AFTER DELETE ON messages BEGIN
  DELETE FROM message_text WHERE rowid=old.rowid;
 END;`)
	if err != nil {
		return fmt.Errorf("historycache: search index: %w", err)
	}
	return c.putLocked("search/version", []byte(strconv.Itoa(searchVersion)))
}

func (c *Cache) getLocked(key string) ([]byte, error) {
	var b []byte
	err := c.db.QueryRow(`SELECT value FROM kv WHERE key=?`, key).Scan(&b)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return b, err
}

func (c *Cache) putLocked(key string, value []byte) error {
	_, err := c.db.Exec(`INSERT INTO kv VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

// indexExisting adds the messages cached before the index existed, a batch
// at a time. The triggers index everything written meanwhile, and the
// batches replace what they index, so the two never disagree.
func (c *Cache) indexExisting() {
	for {
		done, err := c.indexBatch()
		if err != nil {
			log.Printf("historycache: search index: %v", err)
			return
		}
		if done {
			return
		}
		// Let the history have the cache between batches.
		time.Sleep(20 * time.Millisecond)
	}
}

func (c *Cache) indexBatch() (done bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return true, nil
	}
	b, err := c.getLocked("search/rowid")
	if err != nil {
		return false, err
	}
	if string(b) == "done" {
		return true, nil
	}
	after, _ := strconv.ParseInt(string(b), 10, 64)
	var last *int64
	if err := c.db.QueryRow(`SELECT max(rowid) FROM (SELECT rowid FROM messages WHERE rowid>? ORDER BY rowid LIMIT ?)`, after, searchBatch).Scan(&last); err != nil {
		return false, err
	}
	if last == nil {
		return true, c.putLocked("search/rowid", []byte("done"))
	}
	if _, err := c.db.Exec(`INSERT OR REPLACE INTO message_text(rowid, body) SELECT rowid, `+searchText("payload")+` FROM messages WHERE rowid>? AND rowid<=? AND deleted=0 AND payload IS NOT NULL`, after, *last); err != nil {
		return false, err
	}
	return false, c.putLocked("search/rowid", []byte(strconv.FormatInt(*last, 10)))
}

// Query is a search of the cache.
type Query struct {
	// Text is the words every message found contains, each as a whole word
	// or the start of one; empty finds every message.
	Text string
	// Kinds are the kinds of message to find; none is any.
	Kinds []model.MessageKind
	// Links finds only messages with a link.
	Links bool
	// Chats are the chats to search; nil is all of them.
	Chats []int64
}

// matchQuery turns text into an FTS5 query: every word must be there, as a
// word or the start of one. Quoting each word keeps FTS5 syntax out.
func matchQuery(text string) string {
	text = strings.NewReplacer("ё", "е", "Ё", "Е").Replace(text)
	words := strings.FieldsFunc(text, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })
	for i, w := range words {
		words[i] = `"` + stem(w) + `"*`
	}
	return strings.Join(words, " ")
}

// stem drops the ending of a Russian word, so that the start of the word
// finds its other forms: «дача» finds «даче» and «дачный». Up to two of the
// last letters go when they are vowels, й or ь, and three letters stay.
// Telegram's own search knows the language; this only gets close.
func stem(word string) string {
	runes := []rune(word)
	if len(runes) < 4 || !unicode.Is(unicode.Cyrillic, runes[0]) {
		return word
	}
	for range 2 {
		if len(runes) <= 3 || !strings.ContainsRune("аеиоуыэюяйьАЕИОУЫЭЮЯЙЬ", runes[len(runes)-1]) {
			break
		}
		runes = runes[:len(runes)-1]
	}
	return string(runes)
}

// Search finds cached messages, newest first, skipping offset of them.
func (c *Cache) Search(ctx context.Context, q Query, offset, limit int) ([]model.Message, error) {
	payload := "CAST(m.payload AS TEXT)"
	from := `messages m`
	where := []string{"m.deleted=0", "m.payload IS NOT NULL"}
	var args []any
	if match := matchQuery(q.Text); match != "" {
		from = `message_text f JOIN messages m ON m.rowid=f.rowid`
		where = append(where, "message_text MATCH ?")
		args = append(args, match)
	} else if strings.TrimSpace(q.Text) != "" {
		return nil, nil // Only punctuation: nothing can match.
	}
	if len(q.Kinds) > 0 {
		kinds := make([]string, len(q.Kinds))
		for i, k := range q.Kinds {
			kinds[i] = strconv.Itoa(int(k))
		}
		where = append(where, "json_extract("+payload+",'$.Kind') IN ("+strings.Join(kinds, ",")+")")
	}
	if q.Links {
		where = append(where, "(json_extract("+payload+",'$.WebPage') IS NOT NULL OR EXISTS (SELECT 1 FROM json_each("+payload+",'$.Entities') WHERE json_extract(value,'$.Kind')='url'))")
	}
	if q.Chats != nil {
		ids, _ := json.Marshal(q.Chats)
		where = append(where, "m.chat IN (SELECT value FROM json_each(?))")
		args = append(args, string(ids))
	}
	args = append(args, limit, offset)
	query := `SELECT m.payload FROM ` + from + ` WHERE ` + strings.Join(where, " AND ") +
		` ORDER BY unixepoch(json_extract(` + payload + `,'$.Date')) DESC, m.rowid DESC LIMIT ? OFFSET ?`
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, nil
	}
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Message
	for rows.Next() {
		var b []byte
		if err := rows.Scan(&b); err != nil {
			return nil, err
		}
		var m model.Message
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
