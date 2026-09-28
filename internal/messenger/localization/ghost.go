// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of what the accounts tell others, AyuGram's Ghost Mode. AyuGram's
// strings are not in Telegram's language pack.
func init() {
	for key, texts := range map[string][2]string{
		"ghost.title":             {"Режим призрака", "Ghost Mode"},
		"ghost.body":              {"Что аккаунты сообщают о себе. Без отметок никто не видит, что вы читаете, в сети или печатаете.", "What the accounts tell about themselves. With nothing checked, nobody sees that you read, are online or type."},
		"ghost.read":              {"Отправлять прочтение", "Send read receipts"},
		"ghost.online":            {"Показывать «в сети»", "Show online"},
		"ghost.typing":            {"Отправлять «печатает»", "Send typing"},
		"ghost.interact":          {"Прочитывать при действии", "Read on interact"},
		"ghost.interact_body":     {"Отмечает чат прочитанным, когда вы отправляете в него сообщение или ставите реакцию.", "Marks a chat read when you send a message to it or react in it."},
		"menu.read":               {"Прочитать", "Read Message"},
		"menu.edits":              {"История правок", "Edits History"},
		"menu.filter":             {"Добавить фильтр", "Add Filter"},
		"look.title":              {"Сообщения и аватары", "Messages and Avatars"},
		"look.preview":            {"Так будут выглядеть сообщения", "Messages will look like this"},
		"look.bubble":             {"Радиус пузырей", "Message Bubble Radius"},
		"look.avatar":             {"Скругление аватаров", "Avatar Corners"},
		"look.seconds":            {"Секунды во времени сообщений", "Show Message Seconds"},
		"look.edited_mark":        {"Отметка «изменено»", "Edited Mark"},
		"look.deleted_mark":       {"Отметка «удалено»", "Deleted Mark"},
		"privacy.streamer":        {"Режим стримера", "Streamer Mode"},
		"privacy.streamer_body":   {"Окна мессенджера не попадают в запись и трансляцию экрана: там, где они, будет пусто. Как режим стримера AyuGram.", "The messenger's windows stay out of screen recordings and streams, which show nothing where they are, as AyuGram's Streamer Mode."},
		"filters.added":           {"Фильтр добавлен в общие.", "Filter added to the shared filters."},
		"filters.title":           {"Фильтры сообщений", "Message Filters"},
		"filters.body":            {"Скрывают чужие сообщения, в которых находится выражение (синтаксис Go regexp), или, «наоборот», не находится.", "Hide others' messages an expression (Go regexp syntax) finds, or, reversed, does not."},
		"filters.enable":          {"Включить фильтры", "Enable Filters"},
		"filters.in_chats":        {"Общие фильтры и в чатах", "Shared Filters in Chats Too"},
		"filters.blocked":         {"Скрывать от заблокированных", "Hide from Blocked Users"},
		"filters.channels_only":   {"Без «в чатах» фильтры действуют только в каналах. Меню чата показывает скрытое.", "Without chats, filters apply in channels only. The chat's menu shows what they hid."},
		"filters.expression":      {"Выражение", "Expression"},
		"filters.reversed":        {"Наоборот", "Reversed"},
		"filters.case":            {"Без учёта регистра", "Case insensitive"},
		"filters.add":             {"Добавить", "Add"},
		"filters.error":           {"Ошибка в выражении", "Regex syntax error"},
		"filters.empty":           {"Фильтров пока нет.", "No filters here yet."},
		"filters.one_chat":        {"в одном чате", "in one chat"},
		"filters.delete":          {"Удалить фильтр", "Delete filter"},
		"chat_menu.show_filtered": {"Показать отфильтрованные", "Show Filtered Messages"},
		"chat_menu.hide_filtered": {"Скрыть отфильтрованные", "Hide Filtered Messages"},
		"edits.title":             {"История правок", "Edits History"},
		"edits.current":           {"сейчас", "now"},
		"edits.none":              {"Прежние версии не сохранены: сообщение изменили до того, как оно попало в кеш, или без изменения текста.", "No earlier versions were kept: the message was edited before it was cached, or without changing its text."},
		"edits.failed":            {"Не удалось прочитать историю правок", "Could not read the edits history"},
		"history.deleted_mark":    {"🧹", "🧹"},
		"keep.title":              {"Сохранение сообщений", "Keeping messages"},
		"keep.body":               {"Кеш на этом компьютере помнит то, что в Telegram удалили или изменили.", "The cache on this computer remembers what was deleted or changed in Telegram."},
		"keep.deleted":            {"Сохранять удалённые сообщения", "Save deleted messages"},
		"keep.edits":              {"Сохранять историю правок", "Save edits history"},
		"keep.deleted_body":       {"Удалённые сообщения остаются в истории с отметкой 🧹; в чатах с ботами не сохраняются. Ваши собственные удаления стирают сообщения.", "Deleted messages stay in the history marked 🧹, but in chats with bots. Your own deletions erase messages."},
	} {
		russian[key], english[key] = texts[0], texts[1]
	}
}
