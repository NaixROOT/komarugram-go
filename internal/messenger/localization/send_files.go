// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of the box for sending files, as Telegram Desktop's: the keys are
// its own, the words are ours for when the language pack has none.
func init() {
	for key, texts := range map[string][3]string{
		"files.image":       {"lng_send_image", "Отправить изображение", "Send an image"},
		"files.video":       {"lng_send_video", "Отправить видео", "Send a video file"},
		"files.file":        {"lng_send_file", "Отправить как файл", "Send as a file"},
		"files.group":       {"lng_send_grouped", "Сгруппировать", "Group items"},
		"files.documents":   {"lng_send_as_documents", "Отправить как файлы", "Send as documents"},
		"files.document":    {"lng_send_as_documents_one", "Отправить как файл", "Send as a document"},
		"files.quality":     {"lng_send_high_quality", "Высокое качество", "High Quality"},
		"files.caption":     {"lng_photo_caption", "Подпись", "Caption"},
		"files.empty":       {"lng_send_image_empty", "Файл {name} пуст, его нельзя отправить.", "File: {name} is empty and can't be sent."},
		"files.invalid":     {"lng_send_media_invalid_files", "Подходящих файлов не найдено.", "Sorry, no valid files found."},
		"files.add":         {"", "Добавить", "Add"},
		"files.remove":      {"", "Убрать из отправки", "Remove from sending"},
		"files.loading":     {"", "Читаю файл…", "Reading the file…"},
		"files.path_prompt": {"", "Выбрать файлы нельзя: введите путь к файлу", "No file chooser: enter the file path"},
	} {
		if texts[0] != "" {
			TelegramKeys[key] = texts[0]
		}
		russian[key], english[key] = texts[1], texts[2]
	}
	// Counted texts have a form for each number, as Telegram's language pack
	// has.
	for key, texts := range map[string]struct {
		server                      string
		one, few, many, oneE, other string
	}{
		"files.images_selected": {"lng_send_images_selected", "Выбрано {count} изображение", "Выбрано {count} изображения", "Выбрано {count} изображений", "{count} image selected", "{count} images selected"},
		"files.files_selected":  {"lng_send_files_selected", "Выбран {count} файл", "Выбрано {count} файла", "Выбрано {count} файлов", "{count} file selected", "{count} files selected"},
		"files.caption_limit":   {"lng_caption_limit_reached", "Подпись слишком длинная: сократите её на {count} символ.", "Подпись слишком длинная: сократите её на {count} символа.", "Подпись слишком длинная: сократите её на {count} символов.", "The caption is too long: make it shorter by {count} character.", "The caption is too long: make it shorter by {count} characters."},
	} {
		TelegramKeys[key] = texts.server
		russian[key+"#one"], russian[key+"#few"], russian[key+"#many"] = texts.one, texts.few, texts.many
		russian[key] = texts.many
		english[key+"#one"], english[key] = texts.oneE, texts.other
	}
}

// Texts of the areas files dragged over a chat are dropped on, as Telegram
// Desktop's.
func init() {
	for key, texts := range map[string][3]string{
		"drop.files_here":     {"lng_drag_files_here", "Перетащите файлы сюда", "Drop files here"},
		"drop.images_here":    {"lng_drag_images_here", "Перетащите изображения сюда", "Drop images here"},
		"drop.photos_here":    {"lng_drag_photos_here", "Перетащите фото сюда", "Drop photos here"},
		"drop.media_here":     {"lng_drag_media_here", "Перетащите фото и видео", "Drop photos and videos"},
		"drop.quick":          {"lng_drag_to_send_quick", "для быстрой отправки", "to send them in a quick way"},
		"drop.no_compression": {"lng_drag_to_send_no_compression", "чтобы отправить без сжатия", "to send them without compression"},
		"drop.as_files":       {"lng_drag_to_send_files", "чтобы отправить как файлы", "to send them as documents"},
		"drop.as_media":       {"lng_drag_to_send_media", "чтобы отправить как медиафайлы", "to send them as media files"},
	} {
		TelegramKeys[key] = texts[0]
		russian[key], english[key] = texts[1], texts[2]
	}
}
