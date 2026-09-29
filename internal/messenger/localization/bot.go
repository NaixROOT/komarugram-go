// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of bots' buttons and keyboards. The texts of what fails are this
// client's own.
func init() {
	TelegramKeys["bot.start"] = "lng_bot_start"
	for key, texts := range map[string][2]string{
		"bot.start":         {"Запустить", "Start"},
		"bot.hide_keyboard": {"Скрыть клавиатуру", "Hide keyboard"},
		"bot.silent":        {"Бот не ответил", "The bot did not answer"},
		"bot.password":      {"Кнопка требует пароль аккаунта: здесь это не поддерживается", "The button needs the account's password, which this client does not support"},
		"bot.failed":        {"Не удалось нажать кнопку", "Could not press the button"},
	} {
		russian[key], english[key] = texts[0], texts[1]
	}
}
