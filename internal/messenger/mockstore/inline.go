// SPDX-License-Identifier: Unlicense OR MIT

package mockstore

import (
	"context"
	"fmt"
	"strings"

	"komarugram/internal/messenger/model"
)

// The demo's inline bots: @gif answers with a gallery, @wiki with a list of
// articles, whose text the result sends.
var demoArticles = []struct{ title, text string }{
	{"Go", "Go — компилируемый язык программирования от Google."},
	{"Gio", "Gio — библиотека для интерфейсов на Go без зависимостей от платформы."},
	{"Telegram", "Telegram — облачный мессенджер."},
}

func (s *Store) InlineBot(_ context.Context, username string) (model.InlineBot, error) {
	switch strings.ToLower(username) {
	case "gif":
		return model.InlineBot{ID: 9001, Username: "gif", Placeholder: "Поиск GIF"}, nil
	case "wiki":
		return model.InlineBot{ID: 9002, Username: "wiki", Placeholder: "Поиск в Википедии"}, nil
	}
	return model.InlineBot{}, model.ErrNotInlineBot
}

func (s *Store) InlineResults(ctx context.Context, _, bot int64, query, _ string) (model.InlineResults, error) {
	var out model.InlineResults
	if bot == 9001 {
		out.Gallery = true
		for i := range 8 {
			id := fmt.Sprintf("gif-%d", i)
			out.Results = append(out.Results, model.InlineResult{Kind: "gif", Item: model.PickerItem{ID: id, ResultID: id, QueryID: 1, Media: model.Message{Kind: model.MessageGIF, Media: &model.MessageMedia{ID: "demo/gif", MIMEType: "image/gif", Width: 240, Height: 140}}}})
		}
		return out, ctx.Err()
	}
	for i, a := range demoArticles {
		if query == "" || strings.Contains(strings.ToLower(a.title+" "+a.text), strings.ToLower(query)) {
			id := fmt.Sprintf("wiki-%d", i)
			out.Results = append(out.Results, model.InlineResult{Kind: "article", Title: a.title, Description: a.text, Item: model.PickerItem{ID: id, ResultID: id, QueryID: 2}})
		}
	}
	return out, ctx.Err()
}

// inlineText is the text an article result sends.
func inlineText(id string) string {
	var i int
	if _, err := fmt.Sscanf(id, "wiki-%d", &i); err == nil && i >= 0 && i < len(demoArticles) {
		return demoArticles[i].text
	}
	return ""
}
