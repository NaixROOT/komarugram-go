// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

// voiceMessage is voice message id of the user 5, not listened to, without
// a waveform.
func voiceMessage(id int) *tg.Message {
	return &tg.Message{ID: id, PeerID: &tg.PeerUser{UserID: 5}, Date: 10, MediaUnread: true, Media: &tg.MessageMediaDocument{
		Document: &tg.Document{ID: int64(id), MimeType: "audio/ogg", Attributes: []tg.DocumentAttributeClass{&tg.DocumentAttributeAudio{Voice: true, Duration: 3}}},
	}}
}

// listenedStore is a store that shows voice messages 17 and 18 of chat 5.
func listenedStore(t *testing.T) (*Store, int64) {
	t.Helper()
	s := testStore(t)
	chat := int64(5)
	s.history.peers[chat] = peerRecord{Kind: "user", ID: 5, Hash: 55}
	s.history.histories[chat] = &model.History{}
	ms, err := s.ingest(context.Background(), []tg.MessageClass{voiceMessage(17), voiceMessage(18)}, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	s.finishPage(chat, 0, ms, true, nil, 0)
	return s, chat
}

// shown is message id as the history shows it.
func shown(t *testing.T, s *Store, chat int64, id model.MessageID) model.Message {
	t.Helper()
	for _, m := range s.History(chat).Messages {
		if m.Key.MessageID == id {
			return m
		}
	}
	t.Fatalf("no message %d", id)
	return model.Message{}
}

// A voice message played loses its mark here at once, in the cache too;
// Telegram is told only with Ghost's leave.
func TestReadContents(t *testing.T) {
	s, chat := listenedStore(t)
	var mu sync.Mutex
	var told []string
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		mu.Lock()
		defer mu.Unlock()
		if r, ok := in.(*tg.MessagesReadMessageContentsRequest); ok {
			told = append(told, fmt.Sprint(r.ID))
			return &tg.MessagesAffectedMessages{}, nil
		}
		return nil, fmt.Errorf("unexpected %T", in)
	})
	m := shown(t, s, chat, 17)
	if m.Kind != model.MessageVoice || !m.MediaUnread {
		t.Fatalf("the voice message came as %+v", m)
	}
	s.ReadContents(m)
	if shown(t, s, chat, 17).MediaUnread || !shown(t, s, chat, 18).MediaUnread {
		t.Fatal("the mark was not taken off the one played alone")
	}
	cached, ok, err := s.Cache().Message(context.Background(), chat, 17)
	if err != nil || !ok || cached.MediaUnread {
		t.Fatalf("the cache keeps the mark: %v %v", ok, err)
	}
	// Nothing was sent off: without the leave the store does not ask.
	if len(told) != 0 {
		t.Fatalf("told %q without Ghost's leave", told)
	}

	s.SetGhost(model.Ghost{SendRead: true})
	s.ReadContents(shown(t, s, chat, 18))
	// Again is nothing: the mark is gone.
	s.ReadContents(shown(t, s, chat, 18))
	for deadline := time.Now().Add(5 * time.Second); ; time.Sleep(time.Millisecond) {
		mu.Lock()
		got := append([]string(nil), told...)
		mu.Unlock()
		if len(got) == 1 && got[0] == "[18]" {
			break
		}
		if time.Now().After(deadline) || len(got) > 1 {
			t.Fatalf("told %q, want message 18 once", got)
		}
	}
}

// Telegram tells when a voice message was listened to elsewhere, without
// naming the chat outside channels.
func TestContentsReadUpdate(t *testing.T) {
	s, chat := listenedStore(t)
	update := &tg.Updates{Updates: []tg.UpdateClass{&tg.UpdateReadMessagesContents{Messages: []int{18, 99}}}}
	if err := s.Handle(context.Background(), update); err != nil {
		t.Fatal(err)
	}
	if !shown(t, s, chat, 17).MediaUnread || shown(t, s, chat, 18).MediaUnread {
		t.Fatal("the update did not take the mark off message 18 alone")
	}
}

// The waveform worked out stays with the message, and with the cache; the
// copies handed out before keep theirs.
func TestKeepWaveform(t *testing.T) {
	s, chat := listenedStore(t)
	before := shown(t, s, chat, 17)
	s.KeepWaveform(before, []byte{1, 2, 3})
	if len(before.Media.Waveform) != 0 {
		t.Fatal("a copy handed out before changed")
	}
	if got := shown(t, s, chat, 17).Media.Waveform; string(got) != "\x01\x02\x03" {
		t.Fatalf("the message has the waveform %v", got)
	}
	cached, ok, err := s.Cache().Message(context.Background(), chat, 17)
	if err != nil || !ok || string(cached.Media.Waveform) != "\x01\x02\x03" {
		t.Fatalf("the cache has %+v, %v %v", cached.Media, ok, err)
	}
	// Telegram's own is not replaced.
	s.KeepWaveform(shown(t, s, chat, 17), []byte{9})
	if got := shown(t, s, chat, 17).Media.Waveform; string(got) != "\x01\x02\x03" {
		t.Fatalf("the waveform was replaced by %v", got)
	}
}
