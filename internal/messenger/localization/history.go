package localization

// TelegramKeys names only keys whose placeholder contract we implement.
// Application-specific controls deliberately keep their local catalog entries.
var TelegramKeys = map[string]string{
	"nav.saved": "lng_saved_messages", "history.edited": "lng_edited",
	"history.photo": "lng_in_dlg_photo", "history.video": "lng_in_dlg_video",
	"history.sticker": "lng_in_dlg_sticker", "history.empty_message": "lng_message_empty",
	"stickers.count": "lng_stickers_count", "stickers.emoji_count": "lng_custom_emoji_count",
	"stickers.add": "lng_stickers_add_pack", "stickers.remove": "lng_stickers_remove_pack_confirm",
	"composer.featured": "lng_stickers_featured_tab",
}

func (c Catalog) WithTelegram(values map[string]string) Catalog { c.server = values; return c }
func init() {
	ru := map[string]string{"history.selected": "Выбрано: %d", "history.select_forward": "Переслать", "history.select_delete": "Удалить", "history.select_snapshot": "Снимок", "history.offline": "Нет соединения · локальный кеш", "history.loading": "Загрузка…", "history.empty": "Сообщений пока нет", "history.empty_message": "Пустое сообщение", "history.readonly": "Только чтение", "history.actions_disabled": "Действия бота недоступны в режиме чтения", "history.older": "Загрузить предыдущие", "history.newer": "Загрузить следующие", "history.retry": "Повторить", "history.failed": "Не удалось загрузить. Повторите попытку.", "history.edited": "изменено", "history.photo": "Фото", "history.video": "Видео", "history.sticker": "Стикер", "history.gif": "GIF", "history.file": "Файл", "history.service": "Служебное сообщение", "history.reply": "Ответ на сообщение", "history.forward": "Пересланное сообщение", "history.media_failed": "Медиа недоступно · нажмите для повтора", "history.link": "Открыть ссылку", "history.cancel": "Отмена", "history.external": "Открыть во внешнем браузере?", "history.limit": "Файл больше 64 МиБ", "history.paused": "Анимации приостановлены", "viewer.counter": "%d из %d", "viewer.close": "Закрыть", "viewer.failed": "Не удалось загрузить фото · нажмите, чтобы повторить", "viewer.loading_more": "Ищем фото в кеше…", "viewer.window": "Фото"}
	en := map[string]string{"history.selected": "Selected: %d", "history.select_forward": "Forward", "history.select_delete": "Delete", "history.select_snapshot": "Snapshot", "history.offline": "Offline · local cache", "history.loading": "Loading…", "history.empty": "No messages yet", "history.empty_message": "Empty message", "history.readonly": "Read only", "history.actions_disabled": "Bot actions are unavailable in read-only mode", "history.older": "Load older messages", "history.newer": "Load newer messages", "history.retry": "Retry", "history.failed": "Could not load. Please try again.", "history.edited": "edited", "history.photo": "Photo", "history.video": "Video", "history.sticker": "Sticker", "history.gif": "GIF", "history.file": "File", "history.service": "Service message", "history.reply": "Reply to message", "history.forward": "Forwarded message", "history.media_failed": "Media unavailable · click to retry", "history.link": "Open link", "history.cancel": "Cancel", "history.external": "Open in external browser?", "history.limit": "File exceeds 64 MiB", "history.paused": "Animations paused", "viewer.counter": "%d of %d", "viewer.close": "Close", "viewer.failed": "Could not load the photo · click to retry", "viewer.loading_more": "Looking for photos in the cache…", "viewer.window": "Photo"}
	ru["stickers.title"], en["stickers.title"] = "Набор стикеров", "Sticker set"
	ru["stickers.open"], en["stickers.open"] = "Открыть набор стикеров", "Open sticker set"
	ru["stickers.close"], en["stickers.close"] = "Закрыть", "Close"
	ru["stickers.retry"], en["stickers.retry"] = "Повторить", "Retry"
	ru["stickers.failed"], en["stickers.failed"] = "Не удалось загрузить набор", "Could not load sticker set"
	ru["stickers.action_failed"], en["stickers.action_failed"] = "Не удалось изменить набор", "Could not change sticker set"
	ru["stickers.add"], en["stickers.add"] = "Добавить стикеры", "Add stickers"
	ru["stickers.remove"], en["stickers.remove"] = "Удалить", "Remove"
	ru["stickers.author"], en["stickers.author"] = "Автор набора", "Sticker set creator"
	ru["stickers.author_copied"], en["stickers.author_copied"] = "Не удалось открыть автора. ID скопирован: {id}", "Could not open the creator. ID copied: {id}"
	ru["stickers.options"], en["stickers.options"] = "Действия с набором", "Sticker set options"
	ru["stickers.download_zip"], en["stickers.download_zip"] = "Скачать ZIP-архив", "Download ZIP archive"
	ru["stickers.export_failed"], en["stickers.export_failed"] = "Не удалось сохранить набор", "Could not save sticker set"
	ru["stickers.saved"], en["stickers.saved"] = "Архив сохранён: {path}", "Archive saved: {path}"
	ru["stickers.count#one"], en["stickers.count#one"] = "{count} стикер", "{count} sticker"
	ru["stickers.count#few"], ru["stickers.count#many"] = "{count} стикера", "{count} стикеров"
	en["stickers.count#other"] = "{count} stickers"
	ru["stickers.emoji_count"], en["stickers.emoji_count"] = "{count} эмодзи", "{count} emoji"
	for k, v := range ru {
		russian[k] = v
	}
	for k, v := range en {
		english[k] = v
	}
}
