// SPDX-License-Identifier: Unlicense OR MIT

package mockstore

import (
	"fmt"
	"time"

	"komarugram/internal/messenger/model"
)

// DemoForum is the demo's forum, a group of topics.
const DemoForum = 16

// forumThreadBase is where the ids of the topics' chats count down from,
// below the comments' ones.
const forumThreadBase = demoThreadBase - (int64(1) << 40)

// demoTopics are the topics of the demo's forum, in the order the list
// shows them.
func demoTopics(now time.Time) []model.Topic {
	ago := func(d time.Duration) time.Time { return now.Add(-d) }
	return []model.Topic{
		{ID: 2, Title: "Новости релизов", IconColor: 0xFFD67E, Pinned: true, Closed: true, LastSender: "Игорь", LastMessage: "Go 1.27 вышел: заметки к релизу", LastTime: ago(3 * time.Hour)},
		{ID: 1, Title: "General", General: true, Pinned: true, Unread: 3, LastSender: "Анна", LastMessage: "Всем привет, это общий чат клуба", LastTime: ago(50 * time.Minute)},
		{ID: 3, Title: "Вопросы новичков", IconColor: 0x6FB9F0, Unread: 4, Mentions: 1, LastSender: "Дмитрий", LastMessage: "Как правильно завершать горутины?", LastTime: ago(35 * time.Minute)},
		{ID: 4, Title: "Оффтоп", IconColor: 0xCB86DB, Muted: true, Unread: 12, LastSender: "Вы", LastMessage: "Кто смотрел вчерашний матч?", LastTime: ago(2 * time.Hour)},
		{ID: 5, Title: "Вакансии", IconColor: 0x8EEE98, LastSender: "Ольга", LastMessage: "Ищем разработчика на Go в команду платформы", LastTime: ago(26 * time.Hour)},
		{ID: 6, Title: "Встречи", IconColor: 0xFF93B2, LastSender: "Павел", LastMessage: "Митап в четверг, регистрация открыта", LastTime: ago(3 * 24 * time.Hour)},
		{ID: 7, Title: "Архитектура и проектирование", IconColor: 0xFB6F5F, IconEmoji: 0, LastSender: "Сергей", LastMessage: "Обсуждаем, как делить сервис на модули", LastTime: ago(6 * 24 * time.Hour)},
	}
}

// OpenForum implements model.ForumSource; the demo's topics are there.
func (s *Store) OpenForum(int64) {}

// Topics implements model.ForumSource.
func (s *Store) Topics(chat int64) model.TopicList {
	if chat != DemoForum {
		return model.TopicList{}
	}
	return model.TopicList{Topics: demoTopics(s.created)}
}

// LoadMoreTopics implements model.ForumSource.
func (s *Store) LoadMoreTopics(int64) {}

// OpenTopic implements model.ForumSource with made-up messages: the topic's
// creation, then talk between the chat's people.
func (s *Store) OpenTopic(chat int64, topic model.Topic) model.Chat {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := forumThreadBase - chat*1000 - int64(topic.ID)
	out := model.Chat{ID: id, Kind: model.KindGroup, Title: topic.Title}
	if s.histories == nil {
		s.histories = map[int64]model.History{}
	}
	if _, ok := s.histories[id]; ok {
		return out
	}
	start := time.Now().Add(-48 * time.Hour)
	people := []struct {
		id   int64
		name string
	}{{2, "Анна Смирнова"}, {5, "Игорь"}, {9, "Дмитрий Козлов"}}
	key := func(n int) model.MessageKey {
		return model.MessageKey{AccountID: "demo", ChatID: id, MessageID: model.MessageID(topic.ID + n)}
	}
	var messages []model.Message
	if !topic.General {
		messages = append(messages, model.Message{
			Key: key(0), Kind: model.MessageService, Date: start, SenderID: 5, SenderName: "Игорь", ContentRevision: 1,
			Service: &model.ServiceAction{Kind: model.ServiceTopicCreate, Title: topic.Title},
		})
	}
	for n := 1; n <= 14; n++ {
		who := people[n%len(people)]
		m := model.Message{
			Key: key(n), Date: start.Add(time.Duration(n) * 37 * time.Minute), SenderID: who.id, SenderName: who.name, ContentRevision: 1,
			Text:             fmt.Sprintf("Сообщение %d в теме «%s».", n, topic.Title),
			ReplyToMessageID: model.MessageID(topic.ID),
			ForumTopic:       !topic.General,
		}
		if topic.General {
			m.ReplyToMessageID = 0
		}
		if n%5 == 0 {
			m.Outgoing, m.SenderID, m.SenderName = true, s.me.ID, ""
		}
		messages = append(messages, m)
	}
	s.histories[id] = model.History{Messages: messages, Revision: 1, ThreadRoot: model.MessageID(topic.ID)}
	return out
}
