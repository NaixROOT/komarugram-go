package model

import "context"

// StickerSet is a pack preview returned by the account store.
type StickerSet struct {
	Ref       StickerSetRef
	Title     string
	AuthorID  int64
	Count     int
	Items     []PickerItem
	Installed bool
	Emoji     bool
}

// StickerSetStore is optional so offline and demo stores can still show chats.
type StickerSetStore interface {
	StickerSet(context.Context, StickerSetRef) (StickerSet, error)
	SetStickerSetInstalled(context.Context, StickerSetRef, bool) error
}

// StickerSetCache shows a set at once as it was last fetched, while
// StickerSet brings it up to date; the set may also be shown offline.
type StickerSetCache interface {
	CachedStickerSet(context.Context, StickerSetRef) (StickerSet, bool)
}

// StickerSetAuthorSource resolves a creator when Telegram allows access.
type StickerSetAuthorSource interface {
	StickerSetAuthor(context.Context, int64) (Chat, error)
}

// StickerSetCreatorID decodes the creator hint carried by known pack ID
// formats. Telegram does not document this as an ownership guarantee (packs
// can also be transferred). Unknown formats deliberately return no hint.
func StickerSetCreatorID(id int64) int64 {
	if id == 0 {
		return 0
	}
	bits := uint64(id)
	high, serial := bits>>32, uint32(bits)
	switch {
	case serial >= 0x003f0000 && serial < 0x00400000:
		return int64(high | 1<<31)
	case serial < 1<<24:
		return int64(high)
	case serial >= 0xff000000 && serial < 0xff400000:
		return int64(high | 1<<31 | 1<<32)
	case serial >= 0xff400000 && serial < 0xff800000:
		return int64(high | 1<<32)
	default:
		return 0
	}
}
