// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"fmt"
	"hash/fnv"
	"komarugram/internal/messenger/model"
	"strconv"

	"github.com/gotd/td/tg"
)

func giftRarity(raw tg.StarGiftAttributeRarityClass) int {
	if r, ok := raw.(*tg.StarGiftAttributeRarity); ok {
		return r.Permille
	}
	return 0
}
func (s *Store) savedGiftMessage(ctx context.Context, peer peerRecord, saved tg.SavedStarGift, cursor string, index int) (model.Message, error) {
	g := &model.Gift{SenderHidden: saved.NameHidden}
	var doc, pattern tg.DocumentClass
	if saved.FromID != nil && !saved.NameHidden {
		g.SenderID = peerID(saved.FromID)
	}
	switch gift := saved.Gift.(type) {
	case *tg.StarGift:
		g.ID = strconv.FormatInt(gift.ID, 10)
		g.Title = gift.Title
		g.Stars = gift.Stars
		g.Total = gift.AvailabilityTotal
		g.Issued = max(0, gift.AvailabilityTotal-gift.AvailabilityRemains)
		doc = gift.Sticker
	case *tg.StarGiftUnique:
		g.ID = strconv.FormatInt(gift.ID, 10)
		g.Title, g.Slug, g.Number, g.Unique = gift.Title, gift.Slug, gift.Num, true
		g.Total, g.Issued = gift.AvailabilityTotal, gift.AvailabilityIssued
		for _, raw := range gift.Attributes {
			switch a := raw.(type) {
			case *tg.StarGiftAttributeModel:
				g.Model = a.Name
				g.ModelRarity = giftRarity(a.Rarity)
				doc = a.Document
			case *tg.StarGiftAttributePattern:
				g.Symbol = a.Name
				g.SymbolRarity = giftRarity(a.Rarity)
				pattern = a.Document
			case *tg.StarGiftAttributeBackdrop:
				g.Backdrop = a.Name
				g.BackdropRarity = giftRarity(a.Rarity)
				g.HasBackdrop = true
				g.CenterColor = uint32(a.CenterColor) & 0xffffff
				g.EdgeColor = uint32(a.EdgeColor) & 0xffffff
				g.PatternColor = uint32(a.PatternColor) & 0xffffff
				g.TextColor = uint32(a.TextColor) & 0xffffff
			case *tg.StarGiftAttributeOriginalDetails:
				if g.SenderID == 0 && !g.SenderHidden && a.SenderID != nil {
					g.SenderID = peerID(a.SenderID)
				}
			}
		}
	}
	s.history.mu.Lock()
	g.SenderName = s.history.peers[g.SenderID].Name
	s.history.mu.Unlock()
	// Multiple identical standard gifts may have the same catalogue ID. Use the
	// saved instance identity, with an opaque page fallback for public profiles.
	identity := fmt.Sprintf("%s/%d/%d", g.ID, saved.SavedID, saved.MsgID)
	if saved.SavedID == 0 && saved.MsgID == 0 {
		identity += fmt.Sprintf("/%d/%d/%s/%d", saved.Date, g.SenderID, cursor, index)
	}
	h := fnv.New64a()
	h.Write([]byte(identity))
	id := -int(h.Sum64()&0x3fffffffffffffff) - 1
	m, err := s.collectionMessage(ctx, peer, &tg.Message{ID: id, Date: saved.Date, Media: &tg.MessageMediaDocument{Document: doc}})
	if err != nil {
		return m, err
	}
	m.Gift = g
	m.SenderID = g.SenderID
	m.SenderName = g.SenderName
	if pattern != nil {
		p, e := s.collectionMessage(ctx, peer, &tg.Message{ID: id, Media: &tg.MessageMediaDocument{Document: pattern}})
		if e != nil {
			return m, e
		}
		g.Pattern = p.Media
	}
	// Refreshing an expired gift uses the same Telegram collection, never the
	// unrelated story/message namespace that happens to share a numeric ID.
	for _, media := range []*model.MessageMedia{m.Media, g.Pattern} {
		if media == nil {
			continue
		}
		c := s.history
		c.mu.Lock()
		ref := c.refs[media.ID]
		ref.Gift = true
		ref.GiftChat = peerID(peer.output())
		ref.GiftOffset = cursor
		c.refs[media.ID] = ref
		cache := c.cache
		c.mu.Unlock()
		if cache != nil {
			if e := cache.Put(ctx, "ref/"+media.ID, ref); e != nil {
				return m, e
			}
		}
	}
	return m, nil
}
