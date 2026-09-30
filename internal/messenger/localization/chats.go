// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of the chats' settings and of a chat's theme. The ones Telegram
// Desktop shows as well map to its keys, whose texts the server's
// language pack gives.
func init() {
	for key, texts := range map[string][2]string{
		"settings.chats":        {"Настройки чатов", "Chat Settings"},
		"chats.themes":          {"Темы", "Themes"},
		"chats.theme_app":       {"Material", "Material"},
		"chats.theme_classic":   {"Классика", "Classic"},
		"chats.theme_day":       {"Дневная", "Day"},
		"chats.theme_tinted":    {"Тёмно-синяя", "Tinted"},
		"chats.theme_night":     {"Ночная", "Night"},
		"chats.themes_hint":     {"Светлое и тёмное оформление приложения помнят каждое свою тему чатов; тёмная тема включает тёмное оформление. Тема, заданная в самом чате, важнее.", "The light and the dark app theme each keep their own chat theme; a dark one turns the app dark. A theme set in a chat comes first."},
		"chats.accent":          {"Цвет акцента", "Choose accent color"},
		"chats.wallpaper":       {"Обои чатов", "Chat wallpaper"},
		"chats.from_gallery":    {"Выбрать из галереи", "Choose from gallery"},
		"chats.from_file":       {"Выбрать из файла", "Choose from file"},
		"chats.wallpaper_reset": {"Обои темы", "Theme's wallpaper"},
		"chats.gallery":         {"Выберите обои", "Choose a Wallpaper"},
		"chats.preview":         {"Предпросмотр обоев", "Wallpaper preview"},
		"chats.blur":            {"Размытие", "Blurred"},
		"chats.apply":           {"Применить", "Apply"},
		"chat_theme.choose":     {"Выберите тему", "Select theme"},
		"chat_theme.none":       {"Без\nтемы", "No\nTheme"},
		"chat_theme.today":      {"Сегодня", "Today"},
		"chat_theme.text_in":    {"Эх, молодёжь со своим техно! Слушали бы классику, например Хассельхоффа!", "Ah, you kids today with techno music! You should enjoy the classics, like Hasselhoff!"},
		"chat_theme.text_out":   {"Не могу воспринимать тебя всерьёз.", "I can't even take you seriously right now."},
	} {
		russian[key], english[key] = texts[0], texts[1]
	}
	for key, lng := range map[string]string{
		"settings.chats":      "lng_settings_section_chat_settings",
		"chats.themes":        "lng_settings_themes",
		"chats.theme_classic": "lng_settings_theme_classic",
		"chats.theme_day":     "lng_settings_theme_day",
		"chats.theme_tinted":  "lng_settings_theme_tinted",
		"chats.theme_night":   "lng_settings_theme_night",
		"chats.accent":        "lng_settings_theme_accent_title",
		"chats.wallpaper":     "lng_settings_section_background",
		"chats.from_gallery":  "lng_settings_bg_from_gallery",
		"chats.from_file":     "lng_settings_bg_from_file",
		"chats.gallery":       "lng_backgrounds_header",
		"chats.preview":       "lng_background_header",
		"chats.blur":          "lng_background_blur",
		"chats.apply":         "lng_settings_apply",
		"chat_theme.choose":   "lng_chat_theme_title",
		"chat_theme.none":     "lng_chat_theme_none",
		"chat_theme.text_in":  "lng_background_text1",
		"chat_theme.text_out": "lng_background_text2",
		"chat_theme.local":    "lng_background_apply_me",
	} {
		TelegramKeys[key] = lng
	}
}
