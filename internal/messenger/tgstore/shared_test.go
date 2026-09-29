// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"fmt"
	"komarugram/internal/messenger/model"
	"testing"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func TestSharedSearchIndependentPaginationAndReferences(t *testing.T) {
	s := testStore(t)
	chat := peerID(&tg.PeerChannel{ChannelID: 8})
	s.history.peers[chat] = peerRecord{ID: 8, Hash: 9, Kind: "channel"}
	calls := 0
	s.history.api = tg.NewClient(telegram.InvokeFunc(func(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
		req, ok := in.(*tg.MessagesSearchRequest)
		if !ok {
			t.Fatalf("shared media fetched history: %T", in)
		}
		calls++
		if _, ok := req.Filter.(*tg.InputMessagesFilterPhotos); !ok {
			t.Fatalf("filter %T", req.Filter)
		}
		peer := req.Peer.(*tg.InputPeerChannel)
		if peer.ChannelID != 8 || peer.AccessHash != 9 {
			t.Fatal(peer)
		}
		if calls == 1 && req.OffsetID != 0 || calls == 2 && req.OffsetID != 80 {
			t.Fatalf("offset %d", req.OffsetID)
		}
		raw := []tg.MessageClass{}
		if calls == 1 {
			for _, id := range []int{90, 80} {
				raw = append(raw, &tg.Message{ID: id, PeerID: &tg.PeerChannel{ChannelID: 8}, Date: 1, Media: &tg.MessageMediaPhoto{Photo: &tg.Photo{ID: int64(id), AccessHash: 7, FileReference: []byte{1}, Sizes: []tg.PhotoSizeClass{&tg.PhotoSize{Type: "m", W: 320, H: 240, Size: 500}, &tg.PhotoSize{Type: "x", W: 800, H: 600, Size: 2000}}}}})
			}
		}
		out.(*tg.MessagesMessagesBox).Messages = &tg.MessagesChannelMessages{Count: 2, Messages: raw}
		return nil
	}))
	page, err := s.SharedMedia(context.Background(), chat, model.SharedPhotos, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Messages) != 2 || page.Next != 80 || !page.More || page.Total != 2 {
		t.Fatalf("page %+v", page)
	}
	if len(s.history.histories) != 0 {
		t.Fatal("search opened chat history")
	}
	m := page.Messages[0]
	if _, ok := s.history.refs[m.Media.Variants[0].ID]; !ok {
		t.Fatal("photo variant reference lost")
	}
	next, err := s.SharedMedia(context.Background(), chat, model.SharedPhotos, page.Next, 2)
	if err != nil || next.More || len(next.Messages) != 0 {
		t.Fatal(next, err)
	}
}
func TestSharedSavedUsesSelfAndSourcePeer(t *testing.T) {
	s := testStore(t)
	s.history.peers[42] = peerRecord{ID: 42, Hash: 9, Kind: "user"}
	s.history.api = tg.NewClient(telegram.InvokeFunc(func(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
		req := in.(*tg.MessagesSearchRequest)
		if _, ok := req.Peer.(*tg.InputPeerSelf); !ok {
			t.Fatal("saved search does not target Self")
		}
		source, ok := req.GetSavedPeerID()
		if !ok || source.(*tg.InputPeerUser).UserID != 42 {
			t.Fatal("saved source scope missing")
		}
		out.(*tg.MessagesMessagesBox).Messages = &tg.MessagesMessagesSlice{Count: 1, Messages: []tg.MessageClass{&tg.Message{ID: 900, PeerID: &tg.PeerUser{UserID: 1}, Message: "saved"}}}
		return nil
	}))
	p, e := s.SharedMedia(context.Background(), 42, model.SharedSaved, 0, 60)
	if e != nil || len(p.Messages) != 1 || p.Messages[0].Key.ChatID != 1 {
		t.Fatal(p, e)
	}
}
func TestSharedSearchDoesNotResurrectDeletedOrStallOnEmpty(t *testing.T) {
	s := testStore(t)
	s.history.peers[3] = peerRecord{ID: 3, Kind: "user"}
	s.history.api = tg.NewClient(telegram.InvokeFunc(func(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
		if err := s.Handle(ctx, &tg.UpdateShort{Update: &tg.UpdateDeleteMessages{Messages: []int{10}}}); err != nil {
			return err
		}
		out.(*tg.MessagesMessagesBox).Messages = &tg.MessagesMessagesSlice{Count: 5, Messages: []tg.MessageClass{&tg.Message{ID: 10, PeerID: &tg.PeerUser{UserID: 3}}, &tg.MessageEmpty{ID: 9}}}
		return nil
	}))
	p, e := s.SharedMedia(context.Background(), 3, model.SharedPhotos, 0, 2)
	if e != nil || len(p.Messages) != 0 || p.Next != 9 || !p.More {
		t.Fatal(p, e)
	}
}
func TestSharedCountsAndOptionalCollections(t *testing.T) {
	s := testStore(t)
	s.history.peers[3] = peerRecord{ID: 3, Kind: "user"}
	s.history.api = tg.NewClient(telegram.InvokeFunc(func(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
		switch req := in.(type) {
		case *tg.MessagesGetSearchCountersRequest:
			box := out.(*tg.MessagesSearchCounterVector)
			for i, f := range req.Filters {
				box.Elems = append(box.Elems, tg.MessagesSearchCounter{Filter: f, Count: i + 10})
			}
		case *tg.MessagesSearchRequest:
			out.(*tg.MessagesMessagesBox).Messages = &tg.MessagesMessagesSlice{Count: 5}
		case *tg.StoriesGetPinnedStoriesRequest:
			out.(*tg.StoriesStories).Count = 17
		case *tg.PaymentsGetSavedStarGiftsRequest:
			out.(*tg.PaymentsSavedStarGifts).Count = 16
		case *tg.MessagesGetCommonChatsRequest:
			out.(*tg.MessagesChatsBox).Chats = &tg.MessagesChatsSlice{Count: 3}
		default:
			return fmt.Errorf("unexpected %T", req)
		}
		return nil
	}))
	counts, e := s.SharedCounts(context.Background(), 3)
	if e != nil {
		t.Fatal(e)
	}
	if counts[model.SharedPhotos] != 10 || counts[model.SharedStories] != 17 || counts[model.SharedGifts] != 16 || counts[model.SharedGroups] != 3 || counts[model.SharedSaved] != 5 {
		t.Fatal(counts)
	}
}
func TestPollConversionAndMusic(t *testing.T) {
	poll := &tg.MessageMediaPoll{Poll: tg.Poll{Question: tg.TextWithEntities{Text: "Question?"}, Answers: []tg.PollAnswerClass{&tg.PollAnswer{Text: tg.TextWithEntities{Text: "Yes"}, Option: []byte{1}}}}, Results: tg.PollResults{TotalVoters: 3, Results: []tg.PollAnswerVoters{{Option: []byte{1}, Voters: 3, Chosen: true}}}}
	m, _ := convertMessage("a", &tg.Message{ID: 1, PeerID: &tg.PeerUser{UserID: 3}, Media: poll}, nil)
	if m.Kind != model.MessagePoll || m.Poll == nil || m.Poll.Answers[0].Voters != 3 || !m.Poll.Answers[0].Chosen {
		t.Fatal(m)
	}
	k, meta, _ := documentMedia(&tg.Document{ID: 1, Attributes: []tg.DocumentAttributeClass{&tg.DocumentAttributeAudio{Title: "Song", Performer: "Artist", Duration: 123}}})
	if k != model.MessageMusic || meta.Performer != "Artist" {
		t.Fatal(k, meta)
	}
}

// A voice message keeps the waveform Telegram sends with it, for the
// history to draw.
func TestVoiceKeepsWaveform(t *testing.T) {
	waveform := []byte{1, 2, 3, 4, 5}
	k, meta, _ := documentMedia(&tg.Document{ID: 1, Attributes: []tg.DocumentAttributeClass{&tg.DocumentAttributeAudio{Voice: true, Duration: 7, Waveform: waveform}}})
	if k != model.MessageVoice || string(meta.Waveform) != string(waveform) || meta.Duration != 7*time.Second {
		t.Fatal(k, meta)
	}
}
