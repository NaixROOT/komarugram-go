// SPDX-License-Identifier: Unlicense OR MIT

package localization

// The texts about a frozen account are Telegram Desktop's, taken from the
// server's language pack when it has them.
func init() {
	for key, telegram := range map[string]string{
		"frozen.bar_title": "lng_frozen_bar_title", "frozen.details": "lng_frozen_restrict_text",
		"frozen.restrict_title": "lng_frozen_restrict_title", "frozen.title": "lng_frozen_title",
		"frozen.subtitle1": "lng_frozen_subtitle1", "frozen.text1": "lng_frozen_text1",
		"frozen.subtitle2": "lng_frozen_subtitle2", "frozen.text2": "lng_frozen_text2",
		"frozen.subtitle3": "lng_frozen_subtitle3", "frozen.text3": "lng_frozen_text3",
		"frozen.appeal": "lng_frozen_appeal_button", "frozen.close": "lng_close",
	} {
		TelegramKeys[key] = telegram
	}
	for key, value := range map[string]string{
		"frozen.bar_title":      "Ваш аккаунт заморожен!",
		"frozen.details":        "Нажмите, чтобы узнать подробности",
		"frozen.restrict_title": "Ваш аккаунт заморожен",
		"frozen.title":          "Ваш аккаунт заморожен",
		"frozen.subtitle1":      "Нарушение условий",
		"frozen.text1":          "Ваш аккаунт заморожен за нарушение Условий использования Telegram.",
		"frozen.subtitle2":      "Режим только для чтения",
		"frozen.text2":          "Вы можете пользоваться аккаунтом, но не можете отправлять сообщения и совершать действия.",
		"frozen.subtitle3":      "Обжалование до удаления",
		"frozen.text3":          "Подайте апелляцию через {link} до {date}, иначе аккаунт будет удалён.",
		"frozen.appeal":         "Подать апелляцию",
		"frozen.close":          "Закрыть",
	} {
		russian[key] = value
	}
	for key, value := range map[string]string{
		"frozen.bar_title":      "Your account is frozen!",
		"frozen.details":        "Click to view details",
		"frozen.restrict_title": "Your account is frozen",
		"frozen.title":          "Your Account is Frozen",
		"frozen.subtitle1":      "Violation of Terms",
		"frozen.text1":          "Your account was frozen for breaking Telegram's Terms and Conditions.",
		"frozen.subtitle2":      "Read-Only Mode",
		"frozen.text2":          "You can access your account but can't send messages or take actions.",
		"frozen.subtitle3":      "Appeal Before Deactivation",
		"frozen.text3":          "Appeal via {link} before {date}, or your account will be deleted.",
		"frozen.appeal":         "Submit an Appeal",
		"frozen.close":          "Close",
	} {
		english[key] = value
	}
}
