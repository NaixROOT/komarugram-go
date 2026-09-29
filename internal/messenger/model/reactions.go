// SPDX-License-Identifier: Unlicense OR MIT

package model

import (
	"context"
	"time"
)

// Reactor lets the account react to messages.
type Reactor interface {
	// ChatReactions returns the reactions the account may put on messages
	// of chat and how many of them at once; ok is false while they load,
	// which the call starts.
	ChatReactions(chat int64) (available []Reaction, limit int, ok bool)
	// ToggleReaction chooses reaction on msg, or takes it back when it is
	// chosen. The message changes at once; report gets the error when
	// Telegram refuses, and the change is undone.
	ToggleReaction(msg Message, reaction Reaction, report func(error))
}

// Same reports whether r and o are the same reaction, whatever their counts.
func (r Reaction) Same(o Reaction) bool {
	return r.Emoji == o.Emoji && r.DocumentID == o.DocumentID && r.Paid == o.Paid
}

// ToggleReaction returns reactions with the account's choice of r flipped,
// as Telegram Desktop does (MessageReactions::add and remove): a chosen
// reaction is taken back; otherwise the account keeps the first limit-1 of
// its own in their order and lets the rest go, and r is chosen and moves to
// the end. reactions itself is not changed.
func ToggleReaction(reactions []Reaction, r Reaction, limit int) []Reaction {
	for _, one := range reactions {
		if one.Same(r) && one.Chosen {
			return takeBack(reactions, r)
		}
	}
	limit = max(limit, 1)
	out := make([]Reaction, 0, len(reactions)+1)
	mine := 0
	for _, one := range reactions {
		if one.Chosen && !one.Paid {
			if mine++; mine >= limit {
				one.Chosen = false
				if one.Count--; one.Count <= 0 {
					continue
				}
			}
		}
		out = append(out, one)
	}
	for i, one := range out {
		if one.Same(r) {
			one.Chosen = true
			one.Count++
			return append(append(out[:i:i], out[i+1:]...), one)
		}
	}
	return append(out, Reaction{Emoji: r.Emoji, DocumentID: r.DocumentID, Paid: r.Paid, Count: 1, Chosen: true})
}

// takeBack returns reactions without the account's r.
func takeBack(reactions []Reaction, r Reaction) []Reaction {
	out := make([]Reaction, 0, len(reactions))
	for _, one := range reactions {
		if one.Same(r) {
			one.Chosen = false
			one.Count--
			if one.Count <= 0 {
				continue
			}
		}
		out = append(out, one)
	}
	return out
}

// ChosenReactions lists the account's reactions, in their order, without the
// paid one, which is sent apart.
func ChosenReactions(reactions []Reaction) []Reaction {
	var out []Reaction
	for _, r := range reactions {
		if r.Chosen && !r.Paid {
			out = append(out, r)
		}
	}
	return out
}

// QuickReactor knows the reaction a double click puts on a message: the
// account's default one, which Telegram's config names.
type QuickReactor interface {
	// QuickReaction is the reaction for messages of chat; ok is false
	// when the chat does not allow it or it is not known yet.
	QuickReaction(chat int64) (r Reaction, ok bool)
}

// Reacted is someone who put a reaction on a message.
type Reacted struct {
	PeerID   int64
	Name     string
	Reaction Reaction
	Date     time.Time
}

// ReactedPage is a page of who reacted to a message: Count in all, and
// Next to ask for the next page with, empty after the last.
type ReactedPage struct {
	Count int
	List  []Reacted
	Next  string
}

// ReactionLister lists who reacted to messages whose ReactionsListed is set.
type ReactionLister interface {
	// Reacted lists who put reaction on msg, or any reaction when it is
	// the zero Reaction, from offset.
	Reacted(ctx context.Context, msg Message, reaction Reaction, offset string, limit int) (ReactedPage, error)
}
