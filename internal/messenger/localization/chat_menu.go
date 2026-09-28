// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of the search in a chat and of the menu of a chat's header, as
// Telegram Desktop has them; AyuGram's "To Beginning" is its own.
func init() {
	for key, telegram := range map[string]string{
		"chat_search.hint": "lng_dlg_search_for_messages", "chat_search.position": "lng_search_messages_n_of_amount",
		"chat_search.none": "lng_search_messages_none", "chat_menu.search": "lng_dlg_search_for_messages",
		"chat_menu.profile": "lng_context_view_profile", "chat_menu.group": "lng_context_view_group",
		"chat_menu.channel": "lng_context_view_channel",
	} {
		TelegramKeys[key] = telegram
	}
	for key, texts := range map[string][2]string{
		"chat_search.hint":     {"Поиск сообщений", "Search for messages"},
		"chat_search.position": {"{n} из {amount}", "{n} of {amount}"},
		"chat_search.none":     {"Нет результатов", "No results"},
		"chat_search.older":    {"Раньше", "Earlier"},
		"chat_search.newer":    {"Позже", "Later"},
		"chat_search.close":    {"Закрыть поиск", "Close search"},
		"chat_menu.search":     {"Поиск сообщений", "Search for messages"},
		"chat_menu.profile":    {"Открыть профиль", "View profile"},
		"chat_menu.group":      {"Информация о группе", "View group info"},
		"chat_menu.channel":    {"Информация о канале", "View channel info"},
		"chat_menu.beginning":  {"К началу", "To Beginning"},
		"chat_menu.more":       {"Ещё", "More"},
	} {
		russian[key], english[key] = texts[0], texts[1]
	}
}
