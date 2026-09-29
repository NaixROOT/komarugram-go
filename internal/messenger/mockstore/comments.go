package mockstore

import (
	"fmt"
	"time"

	"komarugram/internal/messenger/model"
)

// demoThreads are the comments chats opened, by post; their ids count down
// from the bottom of int64, below every chat's.
const demoThreadBase = -(int64(1) << 62)

func (s *Store) isChannel(chat int64) bool {
	for _, c := range s.chats {
		if c.ID == chat {
			return c.Kind == model.KindChannel
		}
	}
	return false
}

// OpenComments implements model.CommentsStore with made-up comments: as
// many as the post says, from the chat's people.
func (s *Store) OpenComments(post model.Message) model.Chat {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := demoThreadBase - int64(post.Key.ChatID)*100000 - int64(post.Key.MessageID)
	chat := model.Chat{ID: id, Kind: model.KindGroup}
	for _, c := range s.chats {
		if c.ID == post.Key.ChatID {
			chat.Title = c.Title
		}
	}
	if s.histories == nil {
		s.histories = map[int64]model.History{}
	}
	if _, ok := s.histories[id]; ok {
		return chat
	}
	root := post
	root.Key.ChatID = id
	// The group's copy of the post comes from the channel.
	root.SenderID, root.SenderName = post.Key.ChatID, chat.Title
	root.CommentsOpen, root.Comments, root.Commenters = false, 0, nil
	messages := []model.Message{root}
	people := []struct {
		id   int64
		name string
	}{{2, "Анна Смирнова"}, {5, "Игорь"}, {9, "Дмитрий Козлов"}}
	for i := range post.Comments {
		who := people[i%len(people)]
		m := model.Message{
			Key:        model.MessageKey{AccountID: "demo", ChatID: id, MessageID: post.Key.MessageID + model.MessageID(i+1)},
			Date:       post.Date.Add(time.Duration(i+1) * 7 * time.Minute),
			SenderID:   who.id,
			SenderName: who.name,
			Text:       fmt.Sprintf("Комментарий %d к посту.", i+1),
			// Each comment replies to the post, and every third to the
			// comment before.
			ReplyToMessageID: post.Key.MessageID,
			ContentRevision:  1,
		}
		if i%3 == 2 {
			m.ReplyToMessageID = m.Key.MessageID - 1
		}
		if i%4 == 1 {
			m.Reactions = []model.Reaction{{Emoji: "👍", Count: 2}}
		}
		messages = append(messages, m)
	}
	s.histories[id] = model.History{Messages: messages, Revision: 1, ThreadRoot: post.Key.MessageID}
	return chat
}
