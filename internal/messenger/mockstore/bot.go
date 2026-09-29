// SPDX-License-Identifier: Unlicense OR MIT

package mockstore

import (
	"context"
	"strings"
	"time"

	"komarugram/internal/messenger/model"
)

// DemoNotesBot is the demo's bot whose chat is empty: it has the Start button,
// and answers /start with a keyboard.
const DemoNotesBot = 11

// isBot reports whether chat is a chat with a bot. The caller holds s.mu.
func (s *Store) isBot(chat int64) bool {
	for _, c := range s.chats {
		if c.ID == chat {
			return c.Kind == model.KindBot
		}
	}
	return false
}

// PressButton implements model.BotStore with made-up bot answers, by the
// button's data: a notice, an alert, a link, a bot that stays silent.
func (s *Store) PressButton(ctx context.Context, key model.MessageKey, data []byte) (model.BotAnswer, error) {
	select {
	case <-ctx.Done():
		return model.BotAnswer{}, ctx.Err()
	case <-time.After(400 * time.Millisecond): // As long as a round trip.
	}
	switch string(data) {
	case "refresh":
		return model.BotAnswer{Text: "Обновлено: сейчас +18°, без осадков"}, nil
	case "alert":
		return model.BotAnswer{Text: "Это предупреждение бота: его закрывают кнопкой.", Alert: true}, nil
	case "link":
		return model.BotAnswer{URL: "https://telegram.org"}, nil
	case "silent":
		return model.BotAnswer{}, model.ErrBotSilent
	}
	return model.BotAnswer{}, nil
}

// demoBotMenu is the message a bot chat ends with: buttons of every kind.
func demoBotMenu(chat int64, id model.MessageID, now time.Time) model.Message {
	return model.Message{
		Key: model.MessageKey{AccountID: "demo", ChatID: chat, MessageID: id}, Date: now, SenderName: "Бот", ContentRevision: 1,
		Text: "Кнопки бота: нажмите, и он ответит.",
		Buttons: [][]model.MessageButton{
			{{Text: "Обновить", Kind: "callback", Data: []byte("refresh")}, {Text: "Предупреждение", Kind: "callback", Data: []byte("alert")}},
			{{Text: "Открыть ссылку", Kind: "callback", Data: []byte("link")}, {Text: "Молчит", Kind: "callback", Data: []byte("silent")}},
			{{Text: "Telegram", Kind: "url", URL: "https://telegram.org"}, {Text: "Копировать код", Kind: "copy", Copy: "123-456"}},
			{{Text: "Оплата", Kind: "action"}},
		},
	}
}

// botReply is what a bot says to text sent in its chat with a message of the
// id given, or none: /start brings a keyboard, "Настройки" takes it away.
func botReply(chat int64, id model.MessageID, text string, now time.Time) (model.Message, bool) {
	m := model.Message{Key: model.MessageKey{AccountID: "demo", ChatID: chat, MessageID: id}, Date: now, SenderName: "Бот", ContentRevision: 1}
	switch text {
	case "/start":
		m.Text = "Привет! Выберите действие на клавиатуре под полем ввода."
		m.Keyboard = &model.ReplyKeyboard{Rows: [][]model.MessageButton{
			{{Text: "Новая заметка", Kind: "text"}, {Text: "Мои заметки", Kind: "text"}},
			{{Text: "Настройки", Kind: "text"}, {Text: "Отправить номер", Kind: "action"}},
		}}
	case "Настройки":
		m.Text = "Клавиатура убрана."
		m.KeyboardHide = true
	case "Новая заметка", "Мои заметки":
		m.Text = "«" + text + "»: заметок пока нет."
	default:
		if !strings.HasPrefix(text, "/") {
			return model.Message{}, false
		}
		m.Text = "Команда " + text + " принята."
	}
	return m, true
}

// BotCommands implements model.BotCommandsSource with the commands of a demo
// bot.
func (s *Store) BotCommands(chat int64) []model.BotCommand {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.isBot(chat) {
		return nil
	}
	return []model.BotCommand{
		{Command: "start", Description: "Начать работу с ботом"},
		{Command: "help", Description: "Что умеет бот"},
		{Command: "settings", Description: "Настройки"},
		{Command: "stop", Description: "Остановить уведомления"},
	}
}
