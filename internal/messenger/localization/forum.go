// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of forums, whose chats are lists of topics.
func init() {
	for key, value := range map[string]string{
		"forum.no_topics":   "В этой группе пока нет тем",
		"forum.no_messages": "Здесь пока нет сообщений…",
	} {
		russian[key] = value
	}
	for key, value := range map[string]string{
		"forum.no_topics":   "There are no topics in this group yet",
		"forum.no_messages": "No messages here yet...",
	} {
		english[key] = value
	}
}
