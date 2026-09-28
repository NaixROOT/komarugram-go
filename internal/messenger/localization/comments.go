// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of the comments to channel posts, as Telegram Desktop has them.
func init() {
	for key, telegram := range map[string]string{
		"comments.open": "lng_comments_open_count", "comments.none": "lng_comments_open_none",
		"comments.header": "lng_comments_header", "comments.header_none": "lng_comments_header_none",
		"comments.empty": "lng_replies_no_comments",
	} {
		TelegramKeys[key] = telegram
	}
	for key, value := range map[string]string{
		"comments.open":        "{count} комментариев",
		"comments.open#one":    "{count} комментарий",
		"comments.open#few":    "{count} комментария",
		"comments.open#many":   "{count} комментариев",
		"comments.none":        "Оставить комментарий",
		"comments.header":      "{count} комментариев",
		"comments.header#one":  "{count} комментарий",
		"comments.header#few":  "{count} комментария",
		"comments.header#many": "{count} комментариев",
		"comments.header_none": "Комментарии",
		"comments.empty":       "Здесь пока нет комментариев…",
	} {
		russian[key] = value
	}
	for key, value := range map[string]string{
		"comments.open":        "{count} comments",
		"comments.open#one":    "{count} comment",
		"comments.none":        "Leave a comment",
		"comments.header":      "{count} comments",
		"comments.header#one":  "{count} comment",
		"comments.header_none": "Comments",
		"comments.empty":       "No comments here yet...",
	} {
		english[key] = value
	}
}
