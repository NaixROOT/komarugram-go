package localization

func init() {
	russian["shared.total"] = "Всего: %s"
	english["shared.total"] = "Total: %s"
	russian["shared.stories"] = "Публикации"
	russian["shared.gifts"] = "Подарки"
	russian["shared.groups"] = "Общие группы"
	english["shared.stories"] = "Stories"
	english["shared.gifts"] = "Gifts"
	english["shared.groups"] = "Groups in common"
	ru := map[string]string{"shared.photos": "Фото", "shared.videos": "Видео", "shared.files": "Файлы", "shared.music": "Аудиофайлы", "shared.voice": "Голосовые и видеосообщения", "shared.links": "Ссылки", "shared.gifs": "GIF", "shared.polls": "Опросы", "shared.saved": "Сохранённые сообщения", "shared.empty": "В этом разделе пока ничего нет", "shared.more": "Загрузить ещё", "chat_theme.title": "Оформление чата", "chat_theme.default": "По умолчанию", "chat_theme.apply": "Применить для участников чата", "chat_theme.preview": "Предпросмотр", "chat_theme.reset": "Сбросить", "chat_theme.local": "Применить только здесь", "chat_theme.import": "Импорт .tdesktop-theme", "chat_theme.path": "Путь к файлу темы", "file.open": "Открыть файл", "poll.votes": "Голосов: %d", "poll.closed": "Опрос завершён", "poll.open": "Опрос", "poll.quiz": "Викторина"}
	en := map[string]string{"shared.photos": "Photos", "shared.videos": "Videos", "shared.files": "Files", "shared.music": "Music", "shared.voice": "Voice and video messages", "shared.links": "Links", "shared.gifs": "GIFs", "shared.polls": "Polls", "shared.saved": "Saved messages", "shared.empty": "Nothing in this section yet", "shared.more": "Load more", "chat_theme.title": "Chat appearance", "chat_theme.default": "Default", "chat_theme.apply": "Apply for chat participants", "chat_theme.preview": "Preview", "chat_theme.reset": "Reset", "chat_theme.local": "Apply only here", "chat_theme.import": "Import .tdesktop-theme", "chat_theme.path": "Theme file path", "file.open": "Open file", "poll.votes": "Votes: %d", "poll.closed": "Poll closed", "poll.open": "Poll", "poll.quiz": "Quiz"}
	for k, v := range ru {
		russian[k] = v
	}
	for k, v := range en {
		english[k] = v
	}
	russian["viewer.loading_more"] = "Загрузка фотографий…"
	english["viewer.loading_more"] = "Loading photos…"
}
