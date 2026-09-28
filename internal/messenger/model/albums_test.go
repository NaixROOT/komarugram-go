package model

import "testing"

func TestAlbumsMergeAcrossPagesAndKeepMemberEdits(t *testing.T) {
	member := func(id int, group int64) Message {
		return Message{Key: MessageKey{ChatID: 1, MessageID: MessageID(id)}, GroupedID: group, Media: &MessageMedia{ID: "photo"}, ContentRevision: 1}
	}
	messages := []Message{member(1, 9), member(2, 9), member(3, 0), member(4, 9)}
	got := GroupAlbums(messages)
	if len(got) != 3 || len(got[0].Attachments) != 2 || got[0].Key.MessageID != 1 {
		t.Fatal(got)
	}
	if len(messages[0].Attachments) != 0 {
		t.Fatal("projection changed durable message")
	}
	old := got[0].ContentRevision
	messages[1].ContentRevision = 2
	if GroupAlbums(messages)[0].ContentRevision == old {
		t.Fatal("member edit did not invalidate album")
	}
	if got := GroupAlbums(messages[1:]); len(got[0].Attachments) != 0 || got[0].Key.MessageID != 2 {
		t.Fatal("delete/prepend boundary", got)
	}
	messages[1].Key.ChatID = 2
	if len(GroupAlbums(messages)) != 4 {
		t.Fatal("group merged across chats")
	}
}
