package localization

func init() {
	names := map[string][2]string{
		"message": {"сообщений", "messages"}, "photos": {"фото", "photos"}, "videos": {"видео", "videos"}, "music": {"музыки", "music"}, "files": {"файлов", "files"}, "voice_messages": {"голосовых сообщений", "voice messages"}, "video_messages": {"видеосообщений", "video messages"}, "stickers": {"стикеров", "stickers"}, "gifs": {"GIF", "GIFs"}, "inline": {"inline-контента", "inline content"}, "polls": {"опросов", "polls"},
	}
	for key, n := range names {
		for _, suffix := range []string{"", ".all", ".until"} {
			k := "restriction." + key + suffix
			ru := "Вам запрещена отправка " + n[0]
			en := "You are not allowed to send " + n[1]
			tlSuffix := ""
			if suffix == ".all" {
				ru = "В этой группе запрещена отправка " + n[0]
				en = "Sending " + n[1] + " is not allowed in this group"
				tlSuffix = "_all"
			}
			if suffix == ".until" {
				ru += " до {date}, {time}"
				en += " until {date}, {time}"
				tlSuffix = "_until"
			}
			russian[k], english[k] = ru, en
			tl := "lng_restricted_send_" + key + tlSuffix
			if suffix == "" && (key == "voice_messages" || key == "video_messages") {
				tl += "_group"
			}
			TelegramKeys[k] = tl
		}
	}
	for k, v := range map[string][2]string{
		"mute": {"Выключить уведомления", "Mute notifications"}, "unmute": {"Включить уведомления", "Unmute notifications"}, "discussion": {"Обсуждение", "Discussion"}, "notifications_stub": {"Уведомления пока не реализованы", "Notifications are not implemented yet"},
	} {
		russian["composer."+k], english["composer."+k] = v[0], v[1]
	}
	TelegramKeys["composer.discussion"] = "lng_profile_view_discussion"
	TelegramKeys["composer.mute"] = "lng_mute_menu_mute"
	TelegramKeys["composer.unmute"] = "lng_mute_menu_unmute"
}
