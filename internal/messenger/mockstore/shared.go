package mockstore

import (
	"context"
	"komarugram/internal/messenger/model"
	"sort"
	"strconv"
	"time"
)

func (s *Store) SharedCounts(ctx context.Context, chat int64) (map[model.SharedKind]int, error) {
	out := map[model.SharedKind]int{}
	for _, k := range model.SharedKinds {
		p, _ := s.SharedMedia(ctx, chat, k, 0, 100)
		out[k] = p.Total
	}
	out[model.SharedGifts] = 9
	out[model.SharedStories] = 0
	out[model.SharedGroups] = 0
	return out, nil
}
func (s *Store) SharedMedia(ctx context.Context, chat int64, kind model.SharedKind, before model.MessageID, limit int) (model.SharedPage, error) {
	page := model.SharedPage{}
	for _, m := range s.History(chat).Messages {
		match := false
		switch kind {
		case model.SharedPhotos:
			match = m.Kind == model.MessagePhoto
		case model.SharedVideos:
			match = m.Kind == model.MessageVideo
		case model.SharedFiles:
			match = m.Kind == model.MessageFile
		case model.SharedMusic:
			match = m.Kind == model.MessageMusic
		case model.SharedVoice:
			match = m.Kind == model.MessageVoice
		case model.SharedGIFs:
			match = m.Kind == model.MessageGIF
		case model.SharedPolls:
			match = m.Poll != nil
		case model.SharedSaved:
			match = m.Key.MessageID%70 == 0
		case model.SharedLinks:
			for _, e := range m.Entities {
				if e.Kind == "url" {
					match = true
				}
			}
		}
		if match {
			page.Total++
			if before == 0 || m.Key.MessageID < before {
				page.Messages = append(page.Messages, m)
			}
		}
	}
	sort.Slice(page.Messages, func(i, j int) bool { return page.Messages[i].Key.MessageID > page.Messages[j].Key.MessageID })
	page.More = len(page.Messages) > limit
	if page.More {
		page.Messages = page.Messages[:limit]
	}
	if len(page.Messages) > 0 {
		page.Next = page.Messages[len(page.Messages)-1].Key.MessageID
	}
	return page, nil
}
func (s *Store) ChatThemes(context.Context) ([]model.ChatTheme, error) {
	light := &model.ChatThemeStyle{Accent: 0x397f91, OutAccent: 0x276c81, Incoming: 0xffffffff, Text: 0x162d39ff, OutText: 0x102d39ff, Outgoing: []uint32{0xb2e4d2, 0xa9d4ee}, Wallpaper: &model.ChatWallpaper{Colors: []uint32{0xd2e4ef, 0xbbe1c4, 0xf4edd2, 0xc9d5f1}}}
	dark := &model.ChatThemeStyle{Dark: true, Accent: 0x86c8e4, OutAccent: 0xa7dff0, Incoming: 0x20313eff, Text: 0xeff5faff, OutText: 0xf5faffff, Outgoing: []uint32{0x315964, 0x354c70}, Wallpaper: &model.ChatWallpaper{Colors: []uint32{0x182f3c, 0x24463f, 0x353550, 0x1d2839}}}
	return []model.ChatTheme{{ID: "🌿", Title: "🌿 Garden", Light: light, Dark: dark}}, nil
}
func (s *Store) ChatAppearance(ctx context.Context, chat int64) (model.ChatAppearance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return model.ChatAppearance{Theme: s.themes[chat]}, nil
}
func (s *Store) SetChatTheme(ctx context.Context, chat int64, id string) error {
	themes, _ := s.ChatThemes(ctx)
	theme := model.ChatTheme{}
	for _, t := range themes {
		if t.ID == id {
			theme = t
		}
	}
	return s.SetLocalChatTheme(ctx, chat, &theme)
}
func (s *Store) SetLocalChatTheme(ctx context.Context, chat int64, theme *model.ChatTheme) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.themes == nil {
		s.themes = map[int64]model.ChatTheme{}
	}
	if theme == nil {
		delete(s.themes, chat)
	} else {
		s.themes[chat] = *theme
	}
	return nil
}

func (s *Store) SharedCollection(ctx context.Context, chat int64, kind model.SharedKind, cursor string, limit int) (model.SharedCollectionPage, error) {
	p := model.SharedCollectionPage{}
	if kind != model.SharedGifts {
		return p, nil
	}
	p.Total = 9
	start, _ := strconv.Atoi(cursor)
	end := min(p.Total, start+limit)
	for i := start; i < end; i++ {
		g := &model.Gift{ID: strconv.Itoa(i), Title: "Demo Gift", Stars: 100, SenderID: 2, SenderName: "Анна Смирнова"}
		if i%2 == 1 {
			g.Unique = true
			g.Number = 1000 + i
			g.Model = "Aurora"
			g.Symbol = "Flower"
			g.Backdrop = "Lavender"
			g.ModelRarity = 35
			g.SymbolRarity = 20
			g.BackdropRarity = 12
			g.Issued = 950
			g.Total = 1000
			g.CenterColor = 0xbba2e0
			g.EdgeColor = 0x765caa
			g.PatternColor = 0xe5cfff
			g.TextColor = 0xffffff
			g.HasBackdrop = true
			g.Pattern = &model.MessageMedia{ID: "demo/tgs", MIMEType: "application/x-tgsticker", Width: 512, Height: 512}
		}
		p.Messages = append(p.Messages, model.Message{Key: model.MessageKey{AccountID: "demo", ChatID: chat, MessageID: model.MessageID(-i - 1)}, Date: time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC), Kind: model.MessageSticker, Media: &model.MessageMedia{ID: "demo/tgs", MIMEType: "application/x-tgsticker", Width: 512, Height: 512}, Gift: g, ContentRevision: 1})
	}
	if end < p.Total {
		p.Next = strconv.Itoa(end)
	}
	return p, nil
}
