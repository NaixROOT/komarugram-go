// SPDX-License-Identifier: Unlicense OR MIT

package model

import (
	"context"
	"errors"
	"strings"
)

// ReplyKeyboard is the keyboard a bot puts under the composer: rows of
// buttons, which send text or ask the bot something.
type ReplyKeyboard struct {
	Rows [][]MessageButton
	// SingleUse hides it after one press; Persistent keeps it shown even
	// when the account has hidden it.
	SingleUse, Persistent bool
	// Placeholder is what the composer's field says while it is shown.
	Placeholder string
}

// BotAnswer is what a bot answers to a callback button: a short notice, an
// alert to be closed, or a link to open.
type BotAnswer struct {
	Text  string
	Alert bool
	URL   string
}

// Errors of pressing a bot's button.
var (
	// ErrBotSilent is a bot that did not answer in time.
	ErrBotSilent = errors.New("bot: no answer")
	// ErrBotPassword is a button that asks for the account's password,
	// which this client does not.
	ErrBotPassword = errors.New("bot: the button needs a password")
)

// BotStore is a Store that can press the callback buttons of bots.
type BotStore interface {
	// PressButton sends the callback button's data to the bot whose message
	// key is, and returns what it answers.
	PressButton(ctx context.Context, key MessageKey, data []byte) (BotAnswer, error)
}

// ActiveKeyboard returns the reply keyboard the chat shows, and the key of
// the message that set it, from the chat's messages oldest first: the newest
// one that sets or hides it decides, and a single-use keyboard is gone once
// the account wrote after it.
func ActiveKeyboard(messages []Message) (*ReplyKeyboard, MessageKey) {
	for i := len(messages) - 1; i >= 0; i-- {
		m := messages[i]
		switch {
		case m.KeyboardHide:
			return nil, MessageKey{}
		case m.Keyboard != nil:
			if m.Keyboard.SingleUse {
				for _, later := range messages[i+1:] {
					if later.Outgoing {
						return nil, MessageKey{}
					}
				}
			}
			return m.Keyboard, m.Key
		}
	}
	return nil, MessageKey{}
}

// BotCommand is a command of a bot, as its menu lists it.
type BotCommand struct {
	Command, Description string
}

// BotCommandsSource is a Store that knows the commands of the bots the
// account has a chat with.
type BotCommandsSource interface {
	// BotCommands returns the commands of the bot chat is with: none until
	// they are read, which starts on the first call and redraws the window
	// when it is done.
	BotCommands(chat int64) []BotCommand
}

// MatchCommands returns the commands that a text typed in a bot's chat
// starts, in order: "/" all of them, "/st" the ones beginning "st". A text
// that is not a command being typed, one with a space in it, matches none.
func MatchCommands(commands []BotCommand, text string) []BotCommand {
	if !strings.HasPrefix(text, "/") || strings.ContainsAny(text, " \t\n") {
		return nil
	}
	prefix := strings.ToLower(strings.TrimPrefix(text, "/"))
	var out []BotCommand
	for _, c := range commands {
		if strings.HasPrefix(strings.ToLower(c.Command), prefix) {
			out = append(out, c)
		}
	}
	return out
}
