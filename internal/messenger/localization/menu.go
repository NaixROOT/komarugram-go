// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of a message's context menu and of replying, as Telegram Desktop
// has them (history_view_context_menu.cpp, the compose controls).
func init() {
	for key, telegram := range map[string]string{
		"menu.reply": "lng_context_reply_msg", "menu.copy_text": "lng_context_copy_text",
		"menu.copy_selected": "lng_context_copy_selected", "menu.copy_post_link": "lng_context_copy_post_link",
		"menu.copy_message_link": "lng_context_copy_message_link", "menu.forward": "lng_context_forward_msg",
		"menu.forward_selected": "lng_context_forward_selected", "menu.delete": "lng_context_delete_msg",
		"menu.delete_selected": "lng_context_delete_selected", "menu.select": "lng_context_select_msg",
		"menu.clear_selection": "lng_context_clear_selection", "menu.emoji_pack": "lng_context_animated_emoji",
		"menu.emoji_packs": "lng_context_animated_emoji_many", "menu.link_copied": "lng_channel_public_link_copied",
		"menu.private_link": "lng_context_about_private_link",
		"composer.reply_to": "lng_preview_reply_to", "composer.reply_remove": "lng_reply_remove",
		"emoji_packs.title": "lng_custom_emoji_used_sets",
		"pinned.title":      "lng_pinned_message", "pinned.previous": "lng_pinned_previous",
		"pinned.hide":  "lng_pinned_hide_all",
		"menu.reacted": "lng_context_seen_reacted", "reacted.title": "lng_manage_peer_reactions",
	} {
		TelegramKeys[key] = telegram
	}
	for key, value := range map[string]string{
		"menu.reply":             "Ответить",
		"menu.copy_text":         "Копировать текст",
		"menu.copy_selected":     "Копировать выделенный текст",
		"menu.copy_post_link":    "Копировать ссылку",
		"menu.copy_message_link": "Копировать ссылку",
		"menu.forward":           "Переслать",
		"menu.forward_selected":  "Переслать выбранные",
		"menu.delete":            "Удалить",
		"menu.delete_selected":   "Удалить выбранные",
		"menu.select":            "Выделить",
		"menu.clear_selection":   "Отменить выделение",
		"menu.emoji_pack":        "Это сообщение содержит эмодзи из набора {name}.",
		"menu.emoji_packs#one":   "Это сообщение содержит эмодзи из {count} набора.",
		"menu.emoji_packs#few":   "Это сообщение содержит эмодзи из {count} наборов.",
		"menu.emoji_packs#many":  "Это сообщение содержит эмодзи из {count} наборов.",
		"menu.link_copied":       "Ссылка скопирована в буфер обмена.",
		"menu.private_link":      "Эта ссылка будет работать только для участников чата.",
		"composer.reply_to":      "Ответ {name}",
		"composer.reply_remove":  "Не отвечать",
		"emoji_packs.title":      "Наборы использованных эмодзи",
		"emoji_packs.failed":     "Не удалось загрузить наборы",
		"pinned.title":           "Закреплённое сообщение",
		"pinned.previous":        "Предыдущее сообщение",
		"pinned.hide":            "Не показывать закреплённые",
		"menu.reacted#one":       "{count} реакция",
		"menu.reacted#few":       "{count} реакции",
		"menu.reacted#many":      "{count} реакций",
		"reacted.title":          "Реакции",
		"reacted.all":            "Все",
		"reacted.failed":         "Не удалось загрузить список",
	} {
		russian[key] = value
	}
	for key, value := range map[string]string{
		"menu.reply":             "Reply",
		"menu.copy_text":         "Copy Text",
		"menu.copy_selected":     "Copy Selected Text",
		"menu.copy_post_link":    "Copy Post Link",
		"menu.copy_message_link": "Copy Message Link",
		"menu.forward":           "Forward",
		"menu.forward_selected":  "Forward Selected",
		"menu.delete":            "Delete",
		"menu.delete_selected":   "Delete Selected",
		"menu.select":            "Select",
		"menu.clear_selection":   "Clear Selection",
		"menu.emoji_pack":        "This message contains emoji from {name} pack.",
		"menu.emoji_packs#one":   "This message contains emoji from {count} pack.",
		"menu.emoji_packs#other": "This message contains emoji from {count} packs.",
		"menu.link_copied":       "Link copied to clipboard.",
		"menu.private_link":      "This link will only work for members of this chat.",
		"composer.reply_to":      "Reply to {name}",
		"composer.reply_remove":  "Do Not Reply",
		"emoji_packs.title":      "Sets of used emoji",
		"emoji_packs.failed":     "Could not load the sets",
		"pinned.title":           "Pinned message",
		"pinned.previous":        "Previous message",
		"pinned.hide":            "Don't show pinned messages",
		"menu.reacted#one":       "{count} Reacted",
		"menu.reacted#other":     "{count} Reacted",
		"reacted.title":          "Reactions",
		"reacted.all":            "All",
		"reacted.failed":         "Could not load the list",
	} {
		english[key] = value
	}
}
