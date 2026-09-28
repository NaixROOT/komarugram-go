// SPDX-License-Identifier: Unlicense OR MIT

package localization

// The search's sections and its public posts are named as in Telegram
// Desktop, from the server's language pack when it has them.
func init() {
	for key, telegram := range map[string]string{
		"search.chats": "lng_recent_chats", "search.channels": "lng_recent_channels",
		"search.posts": "lng_recent_posts", "search.photos": "lng_all_photos",
		"search.videos": "lng_all_videos", "search.links": "lng_all_links",
		"search.files": "lng_all_files", "search.music": "lng_all_music",
		"search.voice": "lng_all_voice", "search.posts_remaining": "lng_posts_remaining",
		"search.posts_spend": "lng_posts_search_button", "search.posts_limit": "lng_posts_limit_reached",
		"search.posts_unlocks": "lng_posts_limit_unlocks",
		"search.recent":        "lng_recent_title", "search.recent_clear": "lng_recent_clear",
		"search.recent_clear_sure": "lng_recent_clear_sure", "search.recent_remove": "lng_recent_remove",
		"search.recent_clear_all": "lng_recent_clear_all",
	} {
		TelegramKeys[key] = telegram
	}
	for key, value := range map[string]string{
		"search.local":                "Локальный",
		"search.global":               "Глобальный",
		"search.chats":                "Чаты",
		"search.channels":             "Каналы",
		"search.posts":                "Посты",
		"search.photos":               "Фото",
		"search.videos":               "Видео",
		"search.links":                "Ссылки",
		"search.files":                "Файлы",
		"search.music":                "Музыка",
		"search.voice":                "Голосовые",
		"search.messages":             "Сообщения",
		"search.recent":               "Недавние",
		"search.recent_clear":         "Очистить",
		"search.recent_clear_sure":    "Вы действительно хотите очистить историю поиска?",
		"search.recent_remove":        "Удалить из недавних",
		"search.recent_clear_all":     "Очистить всё",
		"search.hint":                 "Введите запрос, чтобы найти чаты и сообщения в Telegram.",
		"search.offline":              "Нет подключения к Telegram. Сохранённое на этом компьютере можно найти локальным поиском.",
		"search.go_local":             "Искать локально",
		"search.failed":               "Не удалось выполнить поиск",
		"search.posts_remaining":      "Осталось бесплатных поисков сегодня: {count}",
		"search.posts_remaining#one":  "Сегодня остался {count} бесплатный поиск",
		"search.posts_remaining#few":  "Сегодня осталось {count} бесплатных поиска",
		"search.posts_remaining#many": "Сегодня осталось {count} бесплатных поисков",
		"search.posts_spend":          "Искать «{query}»",
		"search.posts_limit":          "Лимит исчерпан",
		"search.posts_unlocks":        "бесплатный поиск откроется через {duration}",
	} {
		russian[key] = value
	}
	for key, value := range map[string]string{
		"search.local":               "Local",
		"search.global":              "Global",
		"search.chats":               "Chats",
		"search.channels":            "Channels",
		"search.posts":               "Posts",
		"search.photos":              "Photos",
		"search.videos":              "Videos",
		"search.links":               "Links",
		"search.files":               "Files",
		"search.music":               "Music",
		"search.voice":               "Voice",
		"search.messages":            "Messages",
		"search.recent":              "Recent",
		"search.recent_clear":        "Clear",
		"search.recent_clear_sure":   "Do you want to clear your search history?",
		"search.recent_remove":       "Remove from Recent",
		"search.recent_clear_all":    "Clear all",
		"search.hint":                "Type to find chats and messages in Telegram.",
		"search.offline":             "No connection to Telegram. What is saved on this computer can be found with the local search.",
		"search.go_local":            "Search locally",
		"search.failed":              "Could not search",
		"search.posts_remaining":     "{count} free searches remaining today",
		"search.posts_remaining#one": "{count} free search remaining today",
		"search.posts_spend":         "Search {query}",
		"search.posts_limit":         "Limit Reached",
		"search.posts_unlocks":       "free search unlocks in {duration}",
	} {
		english[key] = value
	}
}
