// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of what the accounts tell others, AyuGram's Ghost Mode. AyuGram's
// strings are not in Telegram's language pack.
func init() {
	for key, texts := range map[string][2]string{
		"ghost.title":         {"Режим призрака", "Ghost Mode"},
		"ghost.body":          {"Что аккаунты сообщают о себе. Без отметок никто не видит, что вы читаете, в сети или печатаете.", "What the accounts tell about themselves. With nothing checked, nobody sees that you read, are online or type."},
		"ghost.read":          {"Отправлять прочтение", "Send read receipts"},
		"ghost.online":        {"Показывать «в сети»", "Show online"},
		"ghost.typing":        {"Отправлять «печатает»", "Send typing"},
		"ghost.interact":      {"Прочитывать при действии", "Read on interact"},
		"ghost.interact_body": {"Отмечает чат прочитанным, когда вы отправляете в него сообщение или ставите реакцию.", "Marks a chat read when you send a message to it or react in it."},
		"menu.read":           {"Прочитать", "Read Message"},
	} {
		russian[key], english[key] = texts[0], texts[1]
	}
}
