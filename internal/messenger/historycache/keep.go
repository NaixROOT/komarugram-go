// SPDX-License-Identifier: Unlicense OR MIT

package historycache

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"komarugram/internal/messenger/model"
)

// deletedKept marks a message Telegram deleted that the cache keeps, as
// AyuGram's saved deleted messages: it stays in the history, and its
// payload says Deleted. 1 is a tombstone, which keeps nothing.
const deletedKept = 2

// SetKeep chooses whether messages Telegram deletes are kept, marked
// deleted, and whether the text others' messages had before an edit is.
func (c *Cache) SetKeep(deleted, edits bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.keepDeleted, c.keepEdits = deleted, edits
}

// KeepsDeleted reports whether messages Telegram deletes are kept.
func (c *Cache) KeepsDeleted() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.keepDeleted
}

// MarkDeleted keeps messages Telegram deleted in the history, marked
// Deleted, and returns them; chat 0 is the ids of every chat but channels,
// which share their ids. A message of a chat skip returns true for, or any
// when deleted messages are not kept, is deleted as by Delete.
func (c *Cache) MarkDeleted(ctx context.Context, chat int64, ids []int, skip func(chat int64) bool) ([]model.Message, error) {
	c.mu.Lock()
	keep := c.keepDeleted
	c.mu.Unlock()
	if !keep {
		return nil, c.Delete(ctx, chat, ids)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var kept []model.Message
	for _, id := range ids {
		chats := []int64{chat}
		if chat == 0 {
			chats, err = chatsOf(ctx, tx, id)
			if err != nil {
				return nil, err
			}
		}
		for _, one := range chats {
			if skip != nil && skip(one) {
				if _, err = tx.ExecContext(ctx, `UPDATE messages SET payload=NULL,deleted=1 WHERE chat=? AND id=?`, one, id); err != nil {
					return nil, err
				}
				continue
			}
			m, ok, err := markDeleted(ctx, tx, one, id)
			if err != nil {
				return nil, err
			}
			if ok {
				kept = append(kept, m)
			}
		}
	}
	return kept, tx.Commit()
}

// chatsOf are the chats, channels aside, that have a message of id.
func chatsOf(ctx context.Context, tx *sql.Tx, id int) ([]int64, error) {
	rows, err := tx.QueryContext(ctx, `SELECT chat FROM messages WHERE id=? AND chat > -1000000000000 AND deleted=0`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var chat int64
		if err = rows.Scan(&chat); err != nil {
			return nil, err
		}
		out = append(out, chat)
	}
	return out, rows.Err()
}

// markDeleted marks message id of chat deleted, if the cache has it.
func markDeleted(ctx context.Context, tx *sql.Tx, chat int64, id int) (model.Message, bool, error) {
	var b []byte
	err := tx.QueryRowContext(ctx, `SELECT payload FROM messages WHERE chat=? AND id=? AND deleted=0 AND payload IS NOT NULL`, chat, id).Scan(&b)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Message{}, false, nil
	}
	if err != nil {
		return model.Message{}, false, err
	}
	var m model.Message
	if err = json.Unmarshal(b, &m); err != nil {
		return model.Message{}, false, err
	}
	m.Deleted = true
	m.ContentRevision = model.Revision(m)
	if b, err = json.Marshal(m); err != nil {
		return model.Message{}, false, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE messages SET payload=?,deleted=? WHERE chat=? AND id=?`, b, deletedKept, chat, id)
	return m, err == nil, err
}

// keepEdit keeps the message m replaces when m changes its text, as its
// version before the edit.
func keepEdit(ctx context.Context, tx *sql.Tx, m model.Message) error {
	var b []byte
	err := tx.QueryRowContext(ctx, `SELECT payload FROM messages WHERE chat=? AND id=? AND deleted=0 AND payload IS NOT NULL`, m.Key.ChatID, m.Key.MessageID).Scan(&b)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	var old model.Message
	if err = json.Unmarshal(b, &old); err != nil {
		return err
	}
	// As AyuGram does, only a text that changed, from something.
	if old.Text == "" || old.Text == m.Text || m.EditedAt.IsZero() || old.EditedAt.Equal(m.EditedAt) {
		return nil
	}
	at := old.EditedAt
	if at.IsZero() {
		at = old.Date
	}
	_, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO edits(chat,id,at,payload) VALUES(?,?,?,?)`, m.Key.ChatID, m.Key.MessageID, at.Unix(), b)
	return err
}

// Edits are the versions message id of chat had before its edits, oldest
// first.
func (c *Cache) Edits(ctx context.Context, chat int64, id int) ([]model.Message, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	rows, err := c.db.QueryContext(ctx, `SELECT payload FROM edits WHERE chat=? AND id=? ORDER BY at`, chat, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Message
	for rows.Next() {
		var b []byte
		if err = rows.Scan(&b); err != nil {
			return nil, err
		}
		var m model.Message
		if err = json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
