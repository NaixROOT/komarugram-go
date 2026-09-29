// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"fmt"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// A message is translated by its id, a text selected by itself.
func TestTranslate(t *testing.T) {
	s := testStore(t)
	s.history.peers[5] = peerRecord{Kind: "user", ID: 5, Hash: 55}
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		r, ok := in.(*tg.MessagesTranslateTextRequest)
		if !ok || r.ToLang != "ru" {
			return nil, fmt.Errorf("unexpected %T %+v", in, in)
		}
		if len(r.ID) == 1 && r.ID[0] == 7 && r.Peer != nil {
			return &tg.MessagesTranslateResult{Result: []tg.TextWithEntities{{Text: "привет"}}}, nil
		}
		if len(r.Text) == 1 && r.Text[0].Text == "world" {
			return &tg.MessagesTranslateResult{Result: []tg.TextWithEntities{{Text: "мир"}}}, nil
		}
		return nil, fmt.Errorf("asked %+v", r)
	})
	ctx := context.Background()
	if got, err := s.Translate(ctx, 5, 7, "hello", "ru"); err != nil || got != "привет" {
		t.Fatalf("the message translates to %q, %v", got, err)
	}
	if got, err := s.Translate(ctx, 5, 0, "world", "ru"); err != nil || got != "мир" {
		t.Fatalf("the text translates to %q, %v", got, err)
	}
}
