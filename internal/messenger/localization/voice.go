// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of recording voice messages, Telegram Desktop's.
func init() {
	for key, texts := range map[string][3]string{
		"record.cancel":   {"lng_record_cancel_recording", "Отменить запись", "Cancel recording"},
		"history.voice":   {"lng_in_dlg_audio", "Голосовое сообщение", "Voice message"},
		"audio.today":     {"lng_player_message_today", "сегодня в {time}", "today at {time}"},
		"audio.yesterday": {"lng_player_message_yesterday", "вчера в {time}", "yesterday at {time}"},
		"audio.date":      {"lng_player_message_date", "{date} в {time}", "{date} at {time}"},
		"audio.close":     {"lng_close", "Закрыть", "Close"},
		"record.problem":  {"lng_record_audio_problem", "Не удалось начать запись звука. Пожалуйста, проверьте микрофон.", "Could not start audio recording. Please check your microphone."},
	} {
		TelegramKeys[key] = texts[0]
		russian[key], english[key] = texts[1], texts[2]
	}
	russian["audio.play"], english["audio.play"] = "Слушать", "Play"
	russian["audio.pause"], english["audio.pause"] = "Пауза", "Pause"
}
