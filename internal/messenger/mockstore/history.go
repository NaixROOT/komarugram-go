package mockstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
	"strings"
	"time"
	"unicode/utf16"

	"komarugram/internal/messenger/model"
)

func (s *Store) OpenChat(id int64) {}
func (s *Store) LoadOlder(int64)   {}
func (s *Store) LoadNewer(int64)   {}
func (s *Store) Viewport(id int64) (model.Viewport, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.views[id]
	return v, ok
}
func (s *Store) SaveView(v model.Viewport, ls []model.MessageLayout) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.views == nil {
		s.views = map[int64]model.Viewport{}
	}
	s.views[v.ChatID] = v
}
func (s *Store) Layouts(int64, model.RenderEnvironment) []model.MessageLayout { return nil }
func (s *Store) History(chat int64) model.History {
	s.mu.Lock()
	defer s.mu.Unlock()
	if h, ok := s.histories[chat]; ok {
		return h
	}
	start := time.Date(2026, 9, 18, 12, 0, 0, 0, time.Local)
	messages := make([]model.Message, 0, 328)
	for i := 1; i <= 320; i++ {
		m := model.Message{Key: model.MessageKey{AccountID: "demo", ChatID: chat, MessageID: model.MessageID(i)}, Date: start.Add(time.Duration(i/80)*24*time.Hour + time.Duration(i)*time.Minute), SenderID: 42, SenderName: "Анна", Text: fmt.Sprintf("Сообщение %d. История сохраняет положение при переключении чатов. 👋", i), ContentRevision: 1, Outgoing: i%4 == 0}
		if i%7 == 0 {
			m.Text = strings.Repeat("Длинный абзац для проверки переноса строк и прокрутки. ", 5)
		}
		if i%10 == 0 {
			m.Reactions = []model.Reaction{{Emoji: "❤", Count: 3}, {Emoji: "👍", Count: 1, Chosen: i%20 == 0}}
		}
		if i%35 == 0 {
			n := i / 35
			m.Kind, m.Media, m.Text = model.MessagePhoto, demoPhoto(n), fmt.Sprintf("Фото %d — откройте, чтобы листать все фото чата", n)
		}
		messages = append(messages, m)
	}
	add := func(text string, kind model.MessageKind, media *model.MessageMedia, entities []model.Entity, buttons [][]model.MessageButton) {
		messages = append(messages, model.Message{Key: model.MessageKey{AccountID: "demo", ChatID: chat, MessageID: model.MessageID(len(messages) + 1)}, Date: start.Add(4*24*time.Hour + time.Duration(len(messages))*time.Minute), SenderID: 42, SenderName: "Демонстрация", Text: text, Kind: kind, Media: media, Entities: entities, Buttons: buttons, ContentRevision: 1})
	}

	rich := "👋 Привет! Жирный, курсив, код и спойлер. Telegram"
	var entities []model.Entity
	for _, pair := range [][2]string{{"bold", "Жирный"}, {"italic", "курсив"}, {"code", "код"}, {"spoiler", "спойлер"}, {"url", "Telegram"}} {
		i := strings.Index(rich, pair[1])
		e := model.Entity{Kind: pair[0], Offset: len(utf16.Encode([]rune(rich[:i]))), Length: len(utf16.Encode([]rune(pair[1])))}
		if e.Kind == "url" {
			e.URL = "https://telegram.org"
		}
		entities = append(entities, e)
	}
	add(rich, model.MessageText, nil, entities, nil)
	messages[len(messages)-1].WebPage = &model.WebPreview{URL: "https://telegram.org", Title: "Telegram", Description: "Telegram Messenger", Photo: &model.MessageMedia{ID: "demo/photo", MIMEType: "image/png", Width: 640, Height: 360}}
	add("", model.MessagePoll, nil, nil, nil)
	messages[len(messages)-1].Poll = &model.Poll{Question: "Какое оформление выберем?", Total: 10, Answers: []model.PollAnswer{{Text: "Светлое", Voters: 6}, {Text: "Тёмное", Voters: 4, Chosen: true}}}

	add("Изображение из локального кеша", model.MessagePhoto, &model.MessageMedia{ID: "demo/photo", MIMEType: "image/png", Width: 640, Height: 360}, nil, nil)
	add("GIF воспроизводится прямо в чате", model.MessageGIF, &model.MessageMedia{ID: "demo/gif", MIMEType: "image/gif", Width: 240, Height: 140}, nil, nil)
	add("", model.MessageSticker, &model.MessageMedia{ID: "demo/tgs", MIMEType: "application/x-tgsticker", Width: 512, Height: 512}, nil, nil)
	add("", model.MessageSticker, &model.MessageMedia{ID: "demo/webm", MIMEType: "video/webm", Width: 512, Height: 512}, nil, nil)
	add("Видео открывается в отдельном окне", model.MessageVideo, &model.MessageMedia{ID: "demo/video", MIMEType: "video/mp4", Width: 640, Height: 360, Size: 7789765, Thumbnail: &model.MessageMedia{ID: "demo/photo", MIMEType: "image/png", Width: 640, Height: 360}}, nil, nil)
	add("Кнопки бота: ссылки доступны, действия оставлены для будущей итерации.", model.MessageText, nil, nil, [][]model.MessageButton{{{Text: "Telegram", Kind: "url", URL: "https://telegram.org"}, {Text: "Обновить", Kind: "callback"}}})
	for i := 0; i < 3; i++ {
		add("Альбом: несколько вложений в одном сообщении", model.MessagePhoto, demoPhoto(len(demoPhotoSizes)-3+i), nil, nil)
		messages[len(messages)-1].GroupedID = 99
	}
	// A lone emoji is drawn large, without a bubble; two stay in one.
	add("🔥", model.MessageText, nil, nil, nil)
	add("❤", model.MessageText, nil, nil, nil)
	messages[len(messages)-1].Outgoing = true
	add("😀😀", model.MessageText, nil, nil, nil)
	add("", model.MessageSticker, &model.MessageMedia{ID: "demo/tgs", MIMEType: "application/x-tgsticker", Width: 512, Height: 512}, nil, nil)
	messages[len(messages)-1].ReplyToMessageID = messages[len(messages)-2].Key.MessageID
	add("Следующая дата остаётся у верхнего края при прокрутке.", model.MessageText, nil, nil, nil)
	// Service messages: what happened in the chat, worded by the UI.
	service := func(a model.ServiceAction, reply model.MessageID) {
		add("", model.MessageService, nil, nil, nil)
		messages[len(messages)-1].Service = &a
		messages[len(messages)-1].ReplyToMessageID = reply
	}
	service(model.ServiceAction{Kind: model.ServiceAddUser, Peers: []model.ServicePeer{{ID: 7, Name: "Ольга"}, {ID: 8, Name: "Павел"}}}, 0)
	service(model.ServiceAction{Kind: model.ServicePin}, messages[len(messages)-2].Key.MessageID)
	service(model.ServiceAction{Kind: model.ServiceEditTitle, Title: "Команда разработки"}, 0)
	service(model.ServiceAction{Kind: model.ServiceTTL, Count: 7 * 86400}, 0)
	service(model.ServiceAction{Kind: model.ServicePhoneCall, Count: 754}, 0)
	spoiler := "Спойлер на нескольких строках: " + strings.Repeat("Нажмите здесь — текст откроется волной от места клика. ", 4)
	add(spoiler, model.MessageText, nil, []model.Entity{{Kind: "spoiler", Offset: 0, Length: len(utf16.Encode([]rune(spoiler)))}}, nil)
	add("Выделите часть этого текста и нажмите Ctrl/Cmd+C. Для выборки сообщений проведите по свободному месту рядом с пузырьками. Escape отменяет выделение.", model.MessageText, nil, nil, nil)
	if s.isChannel(chat) {
		// A channel's posts have no sender and a discussion: the last ones
		// have comments, the others wait for the first one.
		for i := range messages {
			m := &messages[i]
			m.Post, m.Outgoing, m.SenderName, m.SenderID = true, false, "", 0
			m.Views = 1200 + 37*i
			m.CommentsOpen = true
			if i%3 != 0 {
				m.Comments = 1 + i%17
				m.Commenters = []int64{2, 5, 9}[:min(3, m.Comments)]
			}
		}
	}
	h := model.History{Messages: messages, Revision: 1}
	if s.histories == nil {
		s.histories = map[int64]model.History{}
	}
	s.histories[chat] = h
	return h
}
func (s *Store) Media(ctx context.Context, m model.Message) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if n, size, ok := parseDemoPhoto(m.Media.ID); ok {
		return demoPhotoJPEG(n, size.X, size.Y)
	}
	switch m.Media.ID {
	case "demo/tgs":
		return os.ReadFile("stickers/sample.tgs")
	case "demo/webm":
		return os.ReadFile("stickers/circle.webm")
	case "demo/video":
		return os.ReadFile("video.mp4")
	case "demo/photo":
		im := image.NewRGBA(image.Rect(0, 0, 640, 360))
		for y := 0; y < 360; y++ {
			for x := 0; x < 640; x++ {
				im.SetRGBA(x, y, color.RGBA{uint8(30 + x/4), uint8(80 + y/3), 150, 255})
			}
		}
		e := png.Encode(&b, im)
		return b.Bytes(), e
	case "demo/gif":
		g := &gif.GIF{}
		palette := color.Palette{color.RGBA{35, 48, 75, 255}, color.RGBA{130, 190, 250, 255}, color.RGBA{245, 184, 120, 255}}
		for n := 0; n < 24; n++ {
			im := image.NewPaletted(image.Rect(0, 0, 240, 140), palette)
			for y := 35; y < 105; y++ {
				for x := n * 8; x < min(n*8+48, 240); x++ {
					im.SetColorIndex(x, y, 1+uint8(n%2))
				}
			}
			g.Image = append(g.Image, im)
			g.Delay = append(g.Delay, 6)
		}
		e := gif.EncodeAll(&b, g)
		return b.Bytes(), e
	}
	return nil, errors.New("demo media unavailable")
}

// Deterministic fixtures also exercise profile images and the energy policy.
func (s *Store) Avatar(id int64) (model.Message, bool) {
	meta := &model.MessageMedia{ID: "demo/photo", MIMEType: "image/png", Width: 640, Height: 360}
	if id%3 == 0 {
		meta.Thumbnail = &model.MessageMedia{ID: "demo/gif", MIMEType: "image/gif", Width: 240, Height: 140}
	}
	return model.Message{Kind: model.MessagePhoto, Media: meta}, true
}

func (s *Store) HistorySince(chat int64, revision uint64) (model.History, bool) {
	h := s.History(chat)
	if revision == h.Revision {
		h.Messages = nil
		return h, false
	}
	return h, true
}
