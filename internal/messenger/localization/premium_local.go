// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of Local Premium, AyuGram's own feature: its strings are not in
// Telegram's language pack.
func init() {
	for key, texts := range map[string][2]string{
		"premium.local":        {"Локальный Telegram Premium", "Local Telegram Premium"},
		"premium.local_body":   {"Показывает звезду Premium у имён ваших аккаунтов в этом приложении. Telegram о ней не знает: лимиты, эмодзи, реакции и размер загрузки не меняются, и никто больше звезду не видит.", "Shows the Premium star beside the names of your accounts in this app. Telegram does not know of it: limits, emoji, reactions and upload size stay as they are, and nobody else sees the star."},
		"premium.local_active": {"Вы используете локальный Telegram Premium. Преимуществ он не даёт: звезда видна только здесь.", "You are using local Telegram Premium. It gives no benefits: the star is shown here only."},
		"premium.on_local":     {"Локальный", "Local"},
	} {
		russian[key], english[key] = texts[0], texts[1]
	}
}
