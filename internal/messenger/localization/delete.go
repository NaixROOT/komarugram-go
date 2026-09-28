package localization

func init() {
	ru := map[string]string{"title": "Удалить выбранные сообщения?", "self": "Удалить у меня", "all": "Удалить у всех", "channel": "В каналах и супергруппах сообщения удаляются у всех участников.", "busy": "Удаление…", "failed": "Не удалось удалить сообщения"}
	en := map[string]string{"title": "Delete the selected messages?", "self": "Delete for me", "all": "Delete for everyone", "channel": "In channels and supergroups, messages are deleted for all participants.", "busy": "Deleting…", "failed": "Could not delete messages"}
	for k, v := range ru {
		russian["delete."+k] = v
	}
	for k, v := range en {
		english["delete."+k] = v
	}
}
