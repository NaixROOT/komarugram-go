// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of the actions over selected messages, as Telegram Desktop has
// them where it does.
func init() {
	for key, telegram := range map[string]string{
		"delete.sure": "lng_selected_delete_sure", "delete.self_hint": "lng_delete_for_me_chat_hint",
		"delete.all": "lng_delete_for_everyone_check", "delete.confirm": "lng_box_delete",
		"history.forward_to": "lng_forward_choose", "history.forwarded_from": "lng_forwarded",
	} {
		TelegramKeys[key] = telegram
	}
	for key, value := range map[string]string{
		"delete.sure":               "Удалить {count} сообщений?",
		"delete.sure#one":           "Удалить {count} сообщение?",
		"delete.sure#few":           "Удалить {count} сообщения?",
		"delete.sure#many":          "Удалить {count} сообщений?",
		"delete.self_hint":          "Сообщения будут удалены только у вас, но не у других участников чата.",
		"delete.self_hint#one":      "Сообщение будет удалено только у вас, но не у других участников чата.",
		"delete.confirm":            "Удалить",
		"history.forward_to":        "Переслать в…",
		"history.forwarded":         "Переслано в «{chat}»",
		"history.forward_failed":    "Не удалось переслать",
		"history.snapshot_saved":    "Снимок сохранён: {path}",
		"history.snapshot_failed":   "Не удалось сохранить снимок",
		"history.snapshot_too_tall": "Слишком много сообщений для одного снимка",
		"history.forwarded_from":    "Переслано от {name}",
		"history.you":               "Вы",
		"history.reply_missing":     "Сообщение",
	} {
		russian[key] = value
	}
	for key, value := range map[string]string{
		"delete.sure":               "Delete {count} messages?",
		"delete.sure#one":           "Delete {count} message?",
		"delete.self_hint":          "This will delete them just for you, not for other participants of the chat.",
		"delete.self_hint#one":      "This will delete it just for you, not for other participants of the chat.",
		"delete.confirm":            "Delete",
		"history.forward_to":        "Forward to…",
		"history.forwarded":         "Forwarded to “{chat}”",
		"history.forward_failed":    "Could not forward",
		"history.snapshot_saved":    "Snapshot saved: {path}",
		"history.snapshot_failed":   "Could not save the snapshot",
		"history.snapshot_too_tall": "Too many messages for one snapshot",
		"history.forwarded_from":    "Forwarded from {name}",
		"history.you":               "You",
		"history.reply_missing":     "Message",
	} {
		english[key] = value
	}
}
