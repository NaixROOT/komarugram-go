// SPDX-License-Identifier: Unlicense OR MIT

package localization

import (
	"strings"
	"time"
	"unicode/utf16"

	"komarugram/internal/messenger/model"
)

// ServiceContext is what wording a service message needs besides it.
type ServiceContext struct {
	// Chat names the chat: who did what a message without a sender's
	// name tells, as the other side of a private chat.
	Chat string
	// Channel is set in a channel, where actions are the channel's own.
	Channel bool
	// Pinned is the message a pin action pinned, when it is loaded;
	// PinnedGone, when it is known to be deleted.
	Pinned     *model.Message
	PinnedGone bool
	// Now is the time a scheduled video chat's date is worded against.
	Now time.Time
}

// pinnedTextLimit is how many characters of a pinned message's text a pin
// action quotes, as in Telegram Desktop.
const pinnedTextLimit = 16

// Service words a service message, as Telegram Desktop does
// (HistoryItem::setServiceMessageByAction). An action it does not know is
// a service message only; a hidden one, as a group's migration, is empty.
func (c Catalog) Service(m model.Message, ctx ServiceContext) string {
	s := m.Service
	if s == nil || s.Kind == "" {
		return c.T("history.service")
	}
	from := m.SenderName
	if from == "" {
		if m.Outgoing {
			from = c.T("history.you")
		} else {
			from = ctx.Chat
		}
	}
	peer := func(i int) string {
		if i < len(s.Peers) {
			if s.Peers[i].Name != "" {
				return s.Peers[i].Name
			}
		}
		return c.T("service.somebody")
	}
	list := func(prefix string) string {
		var out string
		for i := range s.Peers {
			switch {
			case i == 0:
				out = peer(i)
			case i == len(s.Peers)-1:
				out = c.Format(prefix+"_and_last", map[string]string{"accumulated": out, "user": peer(i)})
			default:
				out = c.Format(prefix+"_and_one", map[string]string{"accumulated": out, "user": peer(i)})
			}
		}
		return out
	}
	args := map[string]string{"from": from}
	f := func(key string) string { return c.Format("service."+key, args) }
	channel := ctx.Channel || m.Post
	switch s.Kind {
	case model.ServiceHidden:
		return ""
	case model.ServiceAddUser:
		switch {
		case len(s.Peers) == 1 && s.Peers[0].ID == m.SenderID:
			return f("user_joined")
		case len(s.Peers) <= 1:
			args["user"] = peer(0)
			return f("add_user")
		}
		args["users"] = list("service.add_users")
		return f("add_users_many")
	case model.ServiceJoinedByLink:
		return f("user_joined_by_link")
	case model.ServiceJoinedByRequest:
		return f("user_joined_by_request")
	case model.ServiceDeleteUser:
		if len(s.Peers) == 0 || s.Peers[0].ID == m.SenderID {
			return f("user_left")
		}
		args["user"] = peer(0)
		return f("kick_user")
	case model.ServiceCreateChat:
		args["title"] = s.Title
		return f("created_chat")
	case model.ServiceCreateChannel:
		if channel {
			return f("created_channel")
		}
		args["title"] = s.Title
		return f("created_chat")
	case model.ServiceEditTitle:
		args["title"] = s.Title
		if channel {
			return f("changed_title_channel")
		}
		return f("changed_title")
	case model.ServiceEditPhoto:
		if channel {
			return f("changed_photo_channel")
		}
		return f("changed_photo")
	case model.ServiceDeletePhoto:
		if channel {
			return f("removed_photo_channel")
		}
		return f("removed_photo")
	case model.ServicePin:
		return c.pinned(args, ctx)
	case model.ServicePhoneCall:
		return c.call(m)
	case model.ServiceGroupCall:
		if s.Count == 0 {
			if channel {
				return f("group_call_started_channel")
			}
			return f("group_call_started_group")
		}
		args["duration"] = c.callDuration(s.Count)
		if channel {
			return f("group_call_finished")
		}
		return f("group_call_finished_group")
	case model.ServiceGroupCallScheduled:
		args["date"] = c.scheduled(s.Date, ctx.Now)
		if channel {
			return f("group_call_scheduled_channel")
		}
		return f("group_call_scheduled_group")
	case model.ServiceInviteToGroupCall:
		args["chat"] = c.T("service.invite_user_chat")
		if len(s.Peers) <= 1 {
			args["user"] = peer(0)
			return f("invite_user")
		}
		args["users"] = list("service.invite_users")
		return f("invite_users_many")
	case model.ServiceScreenshot:
		if m.Outgoing {
			return f("you_took_screenshot")
		}
		return f("took_screenshot")
	case model.ServiceCustom:
		return s.Title
	case model.ServiceContactSignUp:
		return f("user_registered")
	case model.ServiceTTL:
		args["duration"] = c.ttl(s.Count)
		switch {
		case s.Count != 0 && len(s.Peers) > 0 && m.Outgoing && s.Peers[0].ID == m.SenderID:
			return f("ttl_global_me")
		case s.Count != 0 && len(s.Peers) > 0:
			args["from"] = peer(0)
			return f("ttl_global")
		case channel && s.Count == 0:
			return f("ttl_removed_channel")
		case channel:
			return f("ttl_changed_channel")
		case m.Outgoing && s.Count == 0:
			return f("ttl_removed_you")
		case m.Outgoing:
			return f("ttl_changed_you")
		case s.Count == 0:
			return f("ttl_removed")
		}
		return f("ttl_changed")
	case model.ServiceChatTheme:
		args["emoji"] = s.Title
		switch {
		case s.Title == "" && m.Outgoing:
			return f("you_theme_disabled")
		case s.Title == "":
			return f("theme_disabled")
		case m.Outgoing:
			return f("you_theme_changed")
		}
		return f("theme_changed")
	case model.ServiceTopicCreate:
		args["topic"] = s.Title
		return f("topic_created")
	case model.ServiceTopicEdit:
		args["link"] = c.T("service.topic_placeholder")
		args["title"] = s.Title
		switch {
		case s.Reason == "closed":
			return f("topic_closed_inside_by")
		case s.Reason == "reopened":
			return f("topic_reopened_inside_by")
		case s.Title != "":
			return f("topic_renamed")
		}
		return c.T("history.service")
	case model.ServiceBoost:
		if m.Outgoing {
			return f("boost_apply_me")
		}
		return c.Count("service.boost_apply", int(s.Count), args)
	case model.ServiceWallpaper:
		args["user"] = from
		switch {
		case m.Outgoing && s.Reason == "both":
			args["user"] = ctx.Chat
			return f("set_wallpaper_both_me")
		case m.Outgoing && s.Reason == "same":
			return f("set_same_wallpaper_me")
		case m.Outgoing:
			return f("set_wallpaper_me")
		case s.Reason == "same":
			return f("set_same_wallpaper")
		}
		return f("set_wallpaper")
	case model.ServiceGameScore:
		if m.Outgoing {
			return c.Count("service.game_you_scored_no_game", int(s.Count), args)
		}
		return c.Count("service.game_score_no_game", int(s.Count), args)
	case model.ServiceStarGift:
		args["user"] = from
		if s.Count == 0 {
			if m.Outgoing {
				return f("gift_unique_sent")
			}
			return f("gift_unique_received")
		}
		args["cost"] = c.Count("service.gift_for_stars", int(s.Count), nil)
		if m.Outgoing {
			return f("gift_sent")
		}
		return f("gift_received")
	case model.ServiceNoForwards:
		switch {
		case s.Reason == "still":
			return f("no_forwards_still_disabled")
		case m.Outgoing && s.Reason == "disabled":
			return f("no_forwards_you_disabled")
		case m.Outgoing:
			return f("no_forwards_you_enabled")
		case s.Reason == "disabled":
			return f("no_forwards_disabled")
		}
		return f("no_forwards_enabled")
	case model.ServiceSuggestPhoto:
		if m.Outgoing {
			args["user"] = ctx.Chat
			return f("suggested_photo_me")
		}
		args["user"] = from
		return f("suggested_photo")
	case model.ServiceWebViewData:
		args["text"] = s.Title
		return f("webview_data_done")
	}
	return c.T("history.service")
}

// PinnedSummary words a pin whose message is not at hand, as in the chat
// list: who pinned a message, in chat when the message does not say.
func (c Catalog) PinnedSummary(m model.Message, chat string) string {
	from := m.SenderName
	switch {
	case from == "" && m.Outgoing:
		from = c.T("history.you")
	case from == "":
		from = chat
	}
	return c.Format("service.pinned_summary", map[string]string{"from": from})
}

// pinned words a pin: a few words of the text pinned, or what its media
// is, as Telegram Desktop's preparePinnedText does.
func (c Catalog) pinned(args map[string]string, ctx ServiceContext) string {
	p := ctx.Pinned
	switch {
	case p == nil && ctx.PinnedGone:
		args["media"] = c.T("service.deleted_message")
		return c.Format("service.pinned_media", args)
	case p == nil:
		args["media"] = c.T("service.loading")
		return c.Format("service.pinned_media", args)
	}
	if media := c.pinnedMedia(*p); media != "" {
		args["media"] = media
		return c.Format("service.pinned_media", args)
	}
	args["text"] = PinnedQuote(p.Text)
	return c.Format("service.pinned_message", args)
}

// PinnedQuote is the start of a pinned message's text that a pin action
// quotes: its first characters, cut with an ellipsis when it is longer.
func PinnedQuote(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	units := 0
	for i, r := range text {
		if units >= pinnedTextLimit {
			// Five more characters are kept rather than cut.
			if len(utf16.Encode([]rune(text[i:]))) > 5 {
				return text[:i] + "…"
			}
			break
		}
		units += len(utf16.Encode([]rune{r}))
	}
	return text
}

// pinnedMedia names what a pinned message's media is, or nothing for a
// message of text.
func (c Catalog) pinnedMedia(m model.Message) string {
	switch m.Kind {
	case model.MessagePhoto:
		return c.T("service.pinned_media_photo")
	case model.MessageVideo:
		return c.T("service.pinned_media_video")
	case model.MessageGIF:
		return c.T("service.pinned_media_gif")
	case model.MessageVoice:
		return c.T("service.pinned_media_voice")
	case model.MessageMusic:
		return c.T("service.pinned_media_audio")
	case model.MessageFile:
		return c.T("service.pinned_media_file")
	case model.MessageSticker:
		return c.T("service.pinned_media_sticker")
	case model.MessagePoll:
		if m.Poll != nil {
			return "«" + m.Poll.Question + "»"
		}
	}
	return ""
}

// call words a call as its message says it: outgoing, missed, declined,
// and how long it was.
func (c Catalog) call(m model.Message) string {
	s := m.Service
	video := ""
	if s.Video {
		video = "video_"
	}
	var key string
	switch {
	case m.Outgoing && s.Reason == "missed":
		key = "cancelled"
	case m.Outgoing:
		key = "outgoing"
	case s.Reason == "missed":
		key = "missed"
	case s.Reason == "busy":
		key = "declined"
	default:
		key = "incoming"
	}
	text := c.T("call." + video + key)
	if s.Count > 0 {
		text += " (" + c.callDuration(s.Count) + ")"
	}
	return text
}

// callDuration words how long a call went, in its largest unit.
func (c Catalog) callDuration(seconds int64) string {
	switch {
	case seconds/86400 > 1:
		return c.Count("duration.days", int(seconds/86400), nil)
	case seconds/3600 > 1:
		return c.Count("duration.hours", int(seconds/3600), nil)
	case seconds/60 > 1:
		return c.Count("duration.minutes", int(seconds/60), nil)
	}
	return c.Count("duration.seconds", int(seconds), nil)
}

// ttl words an auto-delete period as Telegram Desktop's FormatTTL does.
func (c Catalog) ttl(seconds int64) string {
	const day = 86400
	switch {
	case seconds == 5:
		return c.Count("duration.seconds", 5, nil)
	case seconds < day:
		return c.Count("duration.hours", int(seconds/3600), nil)
	case seconds < 7*day:
		return c.Count("duration.days", int(seconds/day), nil)
	case seconds < 31*day:
		days := int(seconds / day)
		text := c.Count("duration.weeks", days/7, nil)
		if days%7 != 0 {
			text += " " + c.Count("duration.days", days%7, nil)
		}
		return text
	case seconds <= 31*day*11:
		return c.Count("duration.months", int(seconds/(31*day)), nil)
	}
	return c.Count("duration.years", int((seconds+365*day/2)/(365*day)), nil)
}

// scheduled words when a video chat is scheduled: today, tomorrow or a
// date, and the time.
func (c Catalog) scheduled(at, now time.Time) string {
	at = at.Local()
	if now.IsZero() {
		now = time.Now()
	}
	now = now.Local()
	args := map[string]string{"time": at.Format("15:04")}
	y, m, d := at.Date()
	day := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	ny, nm, nd := now.Date()
	today := time.Date(ny, nm, nd, 0, 0, 0, 0, time.Local)
	switch {
	case day.Equal(today):
		return c.Format("service.call_starts_today", args)
	case day.Equal(today.AddDate(0, 0, 1)):
		return c.Format("service.call_starts_tomorrow", args)
	}
	args["date"] = at.Format("02.01.2006")
	return c.Format("service.call_starts_date", args)
}

func init() {
	// The texts of actions keep Telegram's names after "service.".
	actions := map[string][2]string{
		"add_user":                     {"{from} добавил(а) {user}", "{from} added {user}"},
		"add_users_many":               {"{from} добавил(а) {users}", "{from} added {users}"},
		"add_users_and_one":            {"{accumulated}, {user}", "{accumulated}, {user}"},
		"add_users_and_last":           {"{accumulated} и {user}", "{accumulated} and {user}"},
		"kick_user":                    {"{from} удалил(а) {user}", "{from} removed {user}"},
		"user_left":                    {"{from} покинул(а) группу", "{from} left the group"},
		"user_joined":                  {"{from} вступил(а) в группу", "{from} joined the group"},
		"user_joined_by_link":          {"{from} вступил(а) в группу по ссылке-приглашению", "{from} joined the group via invite link"},
		"user_joined_by_request":       {"{from} принят(а) в группу", "{from} was accepted to the group"},
		"user_registered":              {"{from} теперь в Telegram", "{from} joined Telegram"},
		"removed_photo":                {"{from} удалил(а) фото группы", "{from} removed group photo"},
		"removed_photo_channel":        {"Фото канала удалено", "Channel photo removed"},
		"changed_photo":                {"{from} обновил(а) фото группы", "{from} updated group photo"},
		"changed_photo_channel":        {"Фото канала обновлено", "Channel photo updated"},
		"changed_title":                {"{from} изменил(а) название группы на «{title}»", "{from} changed group name to «{title}»"},
		"changed_title_channel":        {"Название канала изменено на «{title}»", "Channel name was changed to «{title}»"},
		"created_chat":                 {"{from} создал(а) группу «{title}»", "{from} created the group «{title}»"},
		"created_channel":              {"Канал создан", "Channel created"},
		"pinned_message":               {"{from} закрепил(а) «{text}»", "{from} pinned \"{text}\""},
		"pinned_media":                 {"{from} закрепил(а) {media}", "{from} pinned {media}"},
		"pinned_media_photo":           {"фото", "a photo"},
		"pinned_media_video":           {"видео", "a video"},
		"pinned_media_audio":           {"аудиофайл", "an audio file"},
		"pinned_media_voice":           {"голосовое сообщение", "a voice message"},
		"pinned_media_file":            {"файл", "a file"},
		"pinned_media_gif":             {"GIF", "a GIF"},
		"pinned_media_sticker":         {"стикер", "a sticker"},
		"took_screenshot":              {"{from} сделал(а) снимок экрана!", "{from} took a screenshot!"},
		"you_took_screenshot":          {"Вы сделали снимок экрана!", "You took a screenshot!"},
		"group_call_started_group":     {"{from} начал(а) видеочат", "{from} started a video chat"},
		"group_call_started_channel":   {"Трансляция началась", "Live stream started"},
		"group_call_scheduled_group":   {"{from} запланировал(а) видеочат на {date}", "{from} scheduled a video chat for {date}"},
		"group_call_scheduled_channel": {"Трансляция запланирована на {date}", "Live stream scheduled for {date}"},
		"group_call_finished":          {"Трансляция завершена ({duration})", "Live stream finished ({duration})"},
		"group_call_finished_group":    {"{from} завершил(а) видеочат ({duration})", "{from} ended the video chat ({duration})"},
		"invite_user":                  {"{from} пригласил(а) {user} в {chat}", "{from} invited {user} to {chat}"},
		"invite_users_many":            {"{from} пригласил(а) {users} в {chat}", "{from} invited {users} to {chat}"},
		"invite_users_and_one":         {"{accumulated}, {user}", "{accumulated}, {user}"},
		"invite_users_and_last":        {"{accumulated} и {user}", "{accumulated} and {user}"},
		"invite_user_chat":             {"видеочат", "the video chat"},
		"ttl_changed":                  {"{from} установил(а) автоудаление сообщений через {duration}", "{from} set messages to auto-delete in {duration}"},
		"ttl_changed_you":              {"Вы установили автоудаление сообщений через {duration}", "You set messages to auto-delete in {duration}"},
		"ttl_changed_channel":          {"Сообщения в этом канале будут автоматически удаляться через {duration}", "Messages in this channel will be automatically deleted after {duration}"},
		"ttl_global":                   {"{from} использует таймер автоудаления для всех чатов. Все новые сообщения в этом чате будут автоматически удалены через {duration} после отправки.", "{from} uses a self-destruct timer for all chats. All new messages in this chat will be automatically deleted after {duration} they are sent."},
		"ttl_global_me":                {"Вы установили таймер автоудаления для всех чатов. Все новые сообщения в этом чате будут автоматически удалены через {duration} после отправки.", "You set a self-destruct timer for all chats. All new messages in this chat will be automatically deleted after {duration} they’ve been sent."},
		"ttl_removed":                  {"{from} отключил(а) автоудаление сообщений", "{from} disabled the auto-delete timer"},
		"ttl_removed_you":              {"Вы отключили автоудаление сообщений", "You disabled the auto-delete timer"},
		"ttl_removed_channel":          {"Сообщения в этом канале больше не будут удаляться автоматически", "Messages in this channel will no longer be automatically deleted"},
		"theme_changed":                {"{from} изменил(а) тему чата на {emoji}", "{from} changed the chat theme to {emoji}"},
		"you_theme_changed":            {"Вы изменили тему чата на {emoji}", "You changed the chat theme to {emoji}"},
		"theme_disabled":               {"{from} отключил(а) тему чата", "{from} disabled the chat theme"},
		"you_theme_disabled":           {"Вы отключили тему чата", "You disabled the chat theme"},
		"topic_created":                {"Создана тема «{topic}»", "The topic \"{topic}\" was created"},
		"topic_renamed":                {"{from} переименовал(а) {link} в «{title}»", "{from} renamed the {link} to \"{title}\""},
		"topic_placeholder":            {"тему", "topic"},
		"topic_closed_inside_by":       {"{from} закрыл(а) тему", "{from} closed the topic"},
		"topic_reopened_inside_by":     {"{from} снова открыл(а) тему", "{from} reopened the topic"},
		"boost_apply_me":               {"Вы усилили группу", "You boosted the group"},
		"set_wallpaper":                {"{user} установил(а) новые обои для этого чата", "{user} set a new wallpaper for this chat"},
		"set_wallpaper_me":             {"Вы установили новые обои для этого чата", "You set a new wallpaper for this chat"},
		"set_wallpaper_both_me":        {"Вы установили новые обои для себя и {user}", "You set a new wallpaper for {user} and you."},
		"set_same_wallpaper":           {"{user} установил(а) те же обои для этого чата", "{user} set the same wallpaper for this chat"},
		"set_same_wallpaper_me":        {"Вы установили те же обои, что и у собеседника", "You set the same wallpaper as your chat partner"},
		"gift_received":                {"{user} отправил(а) вам подарок за {cost}", "{user} sent you a gift for {cost}"},
		"gift_sent":                    {"Вы отправили подарок за {cost}", "You sent a gift for {cost}"},
		"gift_unique_received":         {"{user} отправил(а) вам уникальный коллекционный подарок", "{user} sent you a unique collectible item"},
		"gift_unique_sent":             {"Вы отправили уникальный коллекционный подарок", "You sent a unique collectible item"},
		"no_forwards_you_disabled":     {"Вы запретили пересылку в этом чате", "You disabled sharing in this chat"},
		"no_forwards_you_enabled":      {"Вы разрешили пересылку в этом чате", "You enabled sharing in this chat"},
		"no_forwards_disabled":         {"{from} запретил(а) пересылку в этом чате", "{from} disabled sharing in this chat"},
		"no_forwards_enabled":          {"{from} разрешил(а) пересылку в этом чате", "{from} enabled sharing in this chat"},
		"no_forwards_still_disabled":   {"Пересылка в этом чате по-прежнему запрещена", "Sharing in this chat is still disabled"},
		"suggested_photo":              {"{user} предлагает это фото для вашего профиля в Telegram.", "{user} suggests this photo for your Telegram profile."},
		"suggested_photo_me":           {"Вы предложили это фото для профиля {user} в Telegram.", "You suggested this photo for {user}'s Telegram profile."},
		"webview_data_done":            {"Данные кнопки «{text}» переданы боту.", "Data from the \"{text}\" button was transferred to the bot."},
	}
	for key, texts := range actions {
		TelegramKeys["service."+key] = "lng_action_" + key
		russian["service."+key] = texts[0]
		english["service."+key] = texts[1]
	}
	others := map[string]string{
		"service.boost_apply":             "lng_action_boost_apply",
		"service.game_score_no_game":      "lng_action_game_score_no_game",
		"service.game_you_scored_no_game": "lng_action_game_you_scored_no_game",
		"service.gift_for_stars":          "lng_action_gift_for_stars",
		"duration.seconds":                "lng_seconds",
		"duration.minutes":                "lng_minutes",
		"duration.hours":                  "lng_hours",
		"duration.days":                   "lng_days",
		"duration.weeks":                  "lng_weeks",
		"duration.months":                 "lng_months",
		"duration.years":                  "lng_years",
		"service.deleted_message":         "lng_deleted_message",
		"service.loading":                 "lng_contacts_loading",
		"service.call_starts_today":       "lng_group_call_starts_today",
		"service.call_starts_tomorrow":    "lng_group_call_starts_tomorrow",
		"service.call_starts_date":        "lng_group_call_starts_date",
		"call.outgoing":                   "lng_call_outgoing",
		"call.incoming":                   "lng_call_incoming",
		"call.missed":                     "lng_call_missed",
		"call.cancelled":                  "lng_call_cancelled",
		"call.declined":                   "lng_call_declined",
		"call.video_outgoing":             "lng_call_video_outgoing",
		"call.video_incoming":             "lng_call_video_incoming",
		"call.video_missed":               "lng_call_video_missed",
		"call.video_cancelled":            "lng_call_video_cancelled",
		"call.video_declined":             "lng_call_video_declined",
	}
	for key, telegram := range others {
		TelegramKeys[key] = telegram
	}
	counted := map[string][4]string{
		// Russian one, few, many; English one, other.
		"service.boost_apply":             {"{from} усилил(а) группу", "{from} усилил(а) группу {count} раза", "{from} усилил(а) группу {count} раз", "{from} boosted the group|{from} boosted the group {count} times"},
		"service.game_score_no_game":      {"{from} набрал(а) {count} очко", "{from} набрал(а) {count} очка", "{from} набрал(а) {count} очков", "{from} scored {count}|{from} scored {count}"},
		"service.game_you_scored_no_game": {"Вы набрали {count} очко", "Вы набрали {count} очка", "Вы набрали {count} очков", "You scored {count}|You scored {count}"},
		"service.gift_for_stars":          {"{count} звезду", "{count} звезды", "{count} звёзд", "{count} Star|{count} Stars"},
		"duration.seconds":                {"{count} секунду", "{count} секунды", "{count} секунд", "{count} second|{count} seconds"},
		"duration.minutes":                {"{count} минуту", "{count} минуты", "{count} минут", "{count} minute|{count} minutes"},
		"duration.hours":                  {"{count} час", "{count} часа", "{count} часов", "{count} hour|{count} hours"},
		"duration.days":                   {"{count} день", "{count} дня", "{count} дней", "{count} day|{count} days"},
		"duration.weeks":                  {"{count} неделю", "{count} недели", "{count} недель", "{count} week|{count} weeks"},
		"duration.months":                 {"{count} месяц", "{count} месяца", "{count} месяцев", "{count} month|{count} months"},
		"duration.years":                  {"{count} год", "{count} года", "{count} лет", "{count} year|{count} years"},
	}
	for key, forms := range counted {
		russian[key+"#one"], russian[key+"#few"], russian[key+"#many"] = forms[0], forms[1], forms[2]
		en := strings.SplitN(forms[3], "|", 2)
		english[key+"#one"], english[key+"#other"] = en[0], en[1]
		russian[key], english[key] = forms[2], en[1]
	}
	for key, texts := range map[string][2]string{
		"service.deleted_message":      {"Удалённое сообщение", "Deleted message"},
		"service.pinned_summary":       {"{from} закрепил(а) сообщение", "{from} pinned a message"},
		"service.loading":              {"Загрузка…", "Loading..."},
		"service.somebody":             {"кто-то", "somebody"},
		"service.call_starts_today":    {"сегодня в {time}", "today at {time}"},
		"service.call_starts_tomorrow": {"завтра в {time}", "tomorrow at {time}"},
		"service.call_starts_date":     {"{date} в {time}", "{date} at {time}"},
		"call.outgoing":                {"Исходящий звонок", "Outgoing call"},
		"call.incoming":                {"Входящий звонок", "Incoming call"},
		"call.missed":                  {"Пропущенный звонок", "Missed call"},
		"call.cancelled":               {"Отменённый звонок", "Canceled call"},
		"call.declined":                {"Отклонённый звонок", "Declined call"},
		"call.video_outgoing":          {"Исходящий видеозвонок", "Outgoing video call"},
		"call.video_incoming":          {"Входящий видеозвонок", "Incoming video call"},
		"call.video_missed":            {"Пропущенный видеозвонок", "Missed video call"},
		"call.video_cancelled":         {"Отменённый видеозвонок", "Canceled video call"},
		"call.video_declined":          {"Отклонённый видеозвонок", "Declined video call"},
	} {
		russian[key], english[key] = texts[0], texts[1]
	}
}
