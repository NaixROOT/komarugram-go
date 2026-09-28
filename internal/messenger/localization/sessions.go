// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of the Devices section, as Telegram Desktop's Active Sessions has
// them where it does.
func init() {
	for key, telegram := range map[string]string{
		"sessions.this": "lng_sessions_header", "sessions.other": "lng_sessions_other_header",
		"sessions.other_none": "lng_sessions_other_desc", "sessions.incomplete": "lng_sessions_incomplete",
		"sessions.incomplete_about": "lng_sessions_incomplete_about", "sessions.info": "lng_sessions_info",
		"sessions.application": "lng_sessions_application", "sessions.system": "lng_sessions_system",
		"sessions.ip": "lng_sessions_ip", "sessions.location": "lng_sessions_location",
		"sessions.location_about": "lng_sessions_location_about", "sessions.online": "lng_status_online",
		"sessions.done": "lng_about_done", "sessions.yes": "lng_box_yes", "sessions.no": "lng_box_no",
	} {
		TelegramKeys[key] = telegram
	}
	for key, value := range map[string]string{
		"settings.devices":          "Устройства",
		"settings.devices_hint":     "Сеансы на других устройствах",
		"sessions.this":             "Это устройство",
		"sessions.other":            "Активные устройства",
		"sessions.other_none":       "Вы можете войти в Telegram с другого телефона, планшета или компьютера, используя тот же номер. Все данные мгновенно синхронизируются.",
		"sessions.incomplete":       "Незавершённые попытки входа",
		"sessions.incomplete_about": "У этих устройств нет доступа к сообщениям: код введён верно, но пароль — нет.",
		"sessions.info":             "Сведения",
		"sessions.application":      "Приложение",
		"sessions.system":           "Версия системы",
		"sessions.official":         "Официальное приложение",
		"sessions.ip":               "IP-адрес",
		"sessions.location":         "Местоположение",
		"sessions.location_about":   "Местоположение определено по IP-адресу и может быть неточным.",
		"sessions.online":           "в сети",
		"sessions.done":             "Готово",
		"sessions.yes":              "Да",
		"sessions.no":               "Нет",
		"sessions.failed":           "Не удалось загрузить сеансы",
		"sessions.count":            "{count} устройств",
		"sessions.count#one":        "{count} устройство",
		"sessions.count#few":        "{count} устройства",
		"sessions.count#many":       "{count} устройств",
	} {
		russian[key] = value
	}
	for key, value := range map[string]string{
		"settings.devices":          "Devices",
		"settings.devices_hint":     "Sessions on other devices",
		"sessions.this":             "This device",
		"sessions.other":            "Active Devices",
		"sessions.other_none":       "You can log in to Telegram from other mobile, tablet and desktop devices, using the same phone number. All your data will be instantly synchronized.",
		"sessions.incomplete":       "Incomplete login attempts",
		"sessions.incomplete_about": "The devices above have no access to your messages. The code was entered correctly, but no correct password was given.",
		"sessions.info":             "Info",
		"sessions.application":      "Application",
		"sessions.system":           "System version",
		"sessions.official":         "Official app",
		"sessions.ip":               "IP address",
		"sessions.location":         "Location",
		"sessions.location_about":   "This location estimate is based on the IP address and may not always be accurate.",
		"sessions.online":           "online",
		"sessions.done":             "Done",
		"sessions.yes":              "Yes",
		"sessions.no":               "No",
		"sessions.failed":           "Could not load the sessions",
		"sessions.count":            "{count} devices",
		"sessions.count#one":        "{count} device",
	} {
		english[key] = value
	}
}
