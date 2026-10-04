// SPDX-License-Identifier: Unlicense OR MIT

package model

import (
	"context"
	"errors"
	"regexp"
)

// ErrNotInlineBot is a username that is not of a bot taking inline queries.
var ErrNotInlineBot = errors.New("not an inline bot")

// InlineBot is a bot that answers "@bot query" typed in the field.
type InlineBot struct {
	ID          int64
	Username    string
	Placeholder string
}

// InlineResult is one answer of an inline bot; Item sends it.
type InlineResult struct {
	Kind, Title, Description string
	// Thumb is the picture a list shows, Item.Media what a gallery shows.
	Thumb *MessageMedia
	Item  PickerItem
}

// InlineResults are a bot's answers to a query; Gallery asks for a grid.
type InlineResults struct {
	Results []InlineResult
	Gallery bool
	Next    string
}

type InlineBotSource interface {
	InlineBot(ctx context.Context, username string) (InlineBot, error)
	InlineResults(ctx context.Context, chat, bot int64, query, offset string) (InlineResults, error)
}

// BotStarter starts a bot with the parameter of a t.me/bot?start= link.
type BotStarter interface {
	StartBot(ctx context.Context, chat int64, param string) error
}

var inlineQuery = regexp.MustCompile(`^@([A-Za-z][A-Za-z0-9_]{2,31}) (.*)$`)

// InlineQuery reads "@bot query" from the text of the field, as Telegram
// Desktop: the name and a space start it.
func InlineQuery(text string) (username, query string, ok bool) {
	m := inlineQuery.FindStringSubmatch(text)
	if m == nil {
		return "", "", false
	}
	return m[1], m[2], true
}
