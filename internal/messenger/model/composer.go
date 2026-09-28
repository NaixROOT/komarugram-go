package model

import (
	"context"
	"errors"
	"strings"
	"unicode/utf16"
)

type PickerTab int

const (
	PickerEmoji PickerTab = iota
	PickerStickers
	PickerGIF
)

// PickerItem contains display data only. Telegram access hashes stay in the store.
type PickerItem struct {
	Keywords   string
	ID, Emoji  string
	DocumentID int64
	Media      Message
	Custom     bool
	QueryID    int64
	ResultID   string
}
type PickerPack struct {
	ID    int64
	Title string
	Items []PickerItem
	// Complete is false while Items only contains featured cover documents.
	Complete bool
	// Ref opens the full set from a featured preview.
	Ref StickerSetRef
}
type PickerPage struct {
	Recent []PickerItem
	Packs  []PickerPack
	// Featured contains server-recommended sets not installed by this account.
	Featured []PickerPack
	Items    []PickerItem
	Next     string
}
type PickerRequest struct {
	Tab                     PickerTab
	Query, Offset, Language string
	ChatID                  int64
}
type OutgoingMessage struct {
	RandomID int64
	Text     string
	Entities []Entity
	Item     *PickerItem
	Path     string
	AsMedia  bool
	Tasks    []string
	// ReplyTo is the message of the same chat this one replies to, 0 for
	// none.
	ReplyTo MessageID
}

func (m OutgoingMessage) Validate() error {
	if m.RandomID == 0 {
		return errors.New("missing message identity")
	}
	if strings.TrimSpace(m.Text) == "" && m.Item == nil && m.Path == "" {
		return errors.New("empty message")
	}
	if len(utf16.Encode([]rune(m.Text))) > 4096 {
		return errors.New("message exceeds 4096 characters")
	}
	if len(m.Tasks) > 0 {
		if strings.TrimSpace(m.Text) == "" || len(m.Tasks) > 30 {
			return errors.New("a checklist needs a title and at most 30 tasks")
		}
		for _, t := range m.Tasks {
			if strings.TrimSpace(t) == "" {
				return errors.New("empty task")
			}
		}
	}
	return nil
}

type ComposerStore interface {
	Picker(context.Context, PickerRequest) (PickerPage, error)
	Send(context.Context, int64, OutgoingMessage) error
}

// PreferSaved deduplicates global search results while retaining saved-pack order.
func PreferSaved(saved, global []PickerItem) []PickerItem {
	out := make([]PickerItem, 0, len(saved)+len(global))
	seen := map[string]bool{}
	for _, items := range [][]PickerItem{saved, global} {
		for _, item := range items {
			if !seen[item.ID] {
				seen[item.ID] = true
				out = append(out, item)
			}
		}
	}
	return out
}

// PickerRecentStore optionally persists selections in the account's local cache.
type PickerRecentStore interface {
	RememberPicker(context.Context, PickerTab, PickerItem) error
}
