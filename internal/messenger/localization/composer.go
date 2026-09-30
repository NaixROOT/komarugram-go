package localization

func init() {
	ru := map[string]string{"message": "Сообщение…", "emoji": "Эмодзи", "stickers": "Стикеры", "gif": "GIF", "search": "Поиск…", "recent": "Недавние", "emoji_category1": "Смайлы и люди", "emoji_category2": "Животные и природа", "emoji_category3": "Еда и напитки", "emoji_category4": "Активность", "emoji_category5": "Путешествия и места", "emoji_category6": "Предметы", "emoji_category7": "Символы и флаги", "all": "Все эмодзи", "results": "Результаты поиска", "featured": "Популярные", "view_pack": "Открыть", "empty": "Ничего не найдено", "loading": "Загрузка…", "send": "Отправить", "sending": "Отправка…", "retry": "Повторить", "photo": "Фото или видео", "file": "Файл", "tasks": "Список задач", "path": "Путь к файлу", "browse": "Выбрать…", "title": "Название списка", "items": "Задачи — по одной в строке", "cancel": "Отмена", "unsupported": "Отправка недоступна", "attach": "Вложения", "more": "Показать ещё", "close": "Закрыть", "task_hint": "Создание списков задач требует Telegram Premium"}
	en := map[string]string{"message": "Message…", "emoji": "Emoji", "stickers": "Stickers", "gif": "GIF", "search": "Search…", "recent": "Recent", "emoji_category1": "Emoji & People", "emoji_category2": "Nature", "emoji_category3": "Food & Drink", "emoji_category4": "Activity", "emoji_category5": "Travel & Places", "emoji_category6": "Objects", "emoji_category7": "Symbols & Flags", "all": "All emoji", "results": "Search results", "featured": "Trending", "view_pack": "Open", "empty": "Nothing found", "loading": "Loading…", "send": "Send", "sending": "Sending…", "retry": "Retry", "photo": "Photo or video", "file": "File", "tasks": "Checklist", "path": "File path", "browse": "Browse…", "title": "Checklist title", "items": "Tasks — one per line", "cancel": "Cancel", "unsupported": "Sending unavailable", "attach": "Attachments", "more": "Show more", "close": "Close", "task_hint": "Creating checklists requires Telegram Premium"}
	for k, v := range ru {
		russian["composer."+k] = v
	}
	for k, v := range en {
		english["composer."+k] = v
	}
}
