// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of what the accounts tell others, AyuGram's Ghost Mode. AyuGram's
// strings are not in Telegram's language pack.
func init() {
	for key, texts := range map[string][2]string{
		"ghost.title":          {"Режим призрака", "Ghost Mode"},
		"ghost.body":           {"Что аккаунты сообщают о себе. Без отметок никто не видит, что вы читаете, в сети или печатаете.", "What the accounts tell about themselves. With nothing checked, nobody sees that you read, are online or type."},
		"ghost.read":           {"Отправлять прочтение", "Send read receipts"},
		"ghost.online":         {"Показывать «в сети»", "Show online"},
		"ghost.typing":         {"Отправлять «печатает»", "Send typing"},
		"ghost.interact":       {"Прочитывать при действии", "Read on interact"},
		"ghost.interact_body":  {"Отмечает чат прочитанным, когда вы отправляете в него сообщение или ставите реакцию.", "Marks a chat read when you send a message to it or react in it."},
		"menu.read":            {"Прочитать", "Read Message"},
		"menu.edits":           {"История правок", "Edits History"},
		"edits.title":          {"История правок", "Edits History"},
		"edits.current":        {"сейчас", "now"},
		"edits.none":           {"Прежние версии не сохранены: сообщение изменили до того, как оно попало в кеш, или без изменения текста.", "No earlier versions were kept: the message was edited before it was cached, or without changing its text."},
		"edits.failed":         {"Не удалось прочитать историю правок", "Could not read the edits history"},
		"history.deleted_mark": {"🧹", "🧹"},
		"keep.title":           {"Сохранение сообщений", "Keeping messages"},
		"keep.body":            {"Кеш на этом компьютере помнит то, что в Telegram удалили или изменили.", "The cache on this computer remembers what was deleted or changed in Telegram."},
		"keep.deleted":         {"Сохранять удалённые сообщения", "Save deleted messages"},
		"keep.edits":           {"Сохранять историю правок", "Save edits history"},
		"keep.deleted_body":    {"Удалённые сообщения остаются в истории с отметкой 🧹; в чатах с ботами не сохраняются. Ваши собственные удаления стирают сообщения.", "Deleted messages stay in the history marked 🧹, but in chats with bots. Your own deletions erase messages."},
	} {
		russian[key], english[key] = texts[0], texts[1]
	}
}
