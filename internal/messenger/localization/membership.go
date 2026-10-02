// SPDX-License-Identifier: Unlicense OR MIT

package localization

func init() {
	for key, text := range map[string][2]string{
		"verification":    {"Для вступления нужна проверка через бота. Завершите её в Telegram Desktop.", "Joining requires bot verification. Complete it in Telegram Desktop."},
		"join":            {"Подписаться", "Join Channel"},
		"leave_channel":   {"Отписаться от канала", "Leave channel"},
		"leave_group":     {"Выйти из группы", "Leave group"},
		"leave":           {"Выйти", "Leave"},
		"sure_channel":    {"Вы уверены, что хотите отписаться от канала?", "Are you sure you want to leave this channel?"},
		"sure_group":      {"Вы уверены, что хотите выйти из группы?", "Are you sure you want to leave this group?"},
		"successor":       {"После вашего выхода новым владельцем через 7 дней станет {user}. Чтобы назначить другого владельца, отмените выход и сначала передайте права в Telegram Desktop.", "After you leave, {user} will become the new owner in 7 days. To choose someone else, cancel and transfer ownership in Telegram Desktop first."},
		"successor_basic": {"После вашего выхода новым владельцем сразу станет {user}. Чтобы назначить другого владельца, отмените выход и сначала передайте права в Telegram Desktop.", "After you leave, {user} will immediately become the new owner. To choose someone else, cancel and transfer ownership in Telegram Desktop first."},
		"request_sent":    {"Заявка на вступление отправлена", "Join request sent"},
		"changed":         {"Владелец или преемник изменился. Нажмите «Повторить», чтобы проверить условия выхода.", "The owner or successor changed. Retry to review the consequences of leaving."},
		"owner_blocked":   {"Telegram не разрешил владельцу выйти. Сначала передайте права другому участнику в Telegram Desktop.", "Telegram did not allow the owner to leave. Transfer ownership to another member in Telegram Desktop first."},
		"left_channel":    {"Вы отписались от канала", "You left the channel"},
		"left_group":      {"Вы вышли из группы", "You left the group"},
	} {
		russian["membership."+key], english["membership."+key] = text[0], text[1]
	}
	for key, tl := range map[string]string{"join": "lng_profile_join_channel", "leave_channel": "lng_profile_leave_channel", "leave_group": "lng_profile_leave_group", "leave": "lng_box_leave", "sure_channel": "lng_sure_leave_channel", "sure_group": "lng_sure_leave_group"} {
		TelegramKeys["membership."+key] = tl
	}
}
