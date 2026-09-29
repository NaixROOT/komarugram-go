// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of the menu of a chat in the list. The limit's text is this
// client's own; {limit} and {premium} are the numbers.
func init() {
	for key, telegram := range map[string]string{
		"chat_row.pin":   "lng_context_pin_to_top",
		"chat_row.unpin": "lng_context_unpin_from_top",
		"chat_row.read":  "lng_context_mark_read",
	} {
		TelegramKeys[key] = telegram
	}
	for key, texts := range map[string][2]string{
		"chat_row.pin":           {"Закрепить", "Pin to top"},
		"chat_row.unpin":         {"Открепить", "Unpin from top"},
		"chat_row.read":          {"Прочитано", "Mark as read"},
		"chat_row.pin_limit":     {"Закрепить можно не больше {limit} чатов.", "You can pin no more than {limit} chats."},
		"chat_row.pin_limit_pro": {"Закрепить можно не больше {limit} чатов. С Telegram Premium — до {premium}.", "You can pin no more than {limit} chats. With Telegram Premium, up to {premium}."},
		"chat_row.pin_failed":    {"Не удалось изменить закрепы", "Could not change the pinned chats"},
	} {
		russian[key], english[key] = texts[0], texts[1]
	}
}
