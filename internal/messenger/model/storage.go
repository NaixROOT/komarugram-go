// SPDX-License-Identifier: Unlicense OR MIT

package model

import (
	"context"
	"strconv"
	"strings"
)

// StorageCategory is what a cached media object is, for Data and Storage.
// Previews and ranges take their parent's category.
type StorageCategory uint8

const (
	StorageOther StorageCategory = iota
	StoragePhotos
	StorageVideos
	StorageFiles
	StorageMusic
	StorageVoice
	StorageGIFs
	StorageStickers
	StorageProfilePhotos
	StorageStories
	storageCategories
)

// StorageCategories is the number of categories.
const StorageCategories = int(storageCategories)

// StorageCategoryOf is the category of media that message m shows; key is
// the cached object's own key, which tells avatars and emoji apart.
func StorageCategoryOf(m Message, key string) StorageCategory {
	switch {
	case strings.HasPrefix(key, "avatar/") || IsProfilePhoto(m.Key.MessageID):
		return StorageProfilePhotos
	case strings.HasPrefix(key, "emoji/"):
		return StorageStickers
	}
	switch m.Kind {
	case MessagePhoto:
		return StoragePhotos
	case MessageVideo:
		return StorageVideos
	case MessageFile:
		return StorageFiles
	case MessageMusic:
		return StorageMusic
	case MessageVoice:
		return StorageVoice
	case MessageGIF:
		return StorageGIFs
	case MessageSticker:
		return StorageStickers
	}
	return StorageOther
}

// AvatarChat is the chat whose avatar the cached object key is.
func AvatarChat(key string) (int64, bool) {
	rest, ok := strings.CutPrefix(key, "avatar/")
	if !ok {
		return 0, false
	}
	id, _, _ := strings.Cut(rest, "/")
	chat, err := strconv.ParseInt(id, 10, 64)
	return chat, err == nil && chat != 0
}

// CacheUsage is what one account's cache holds. Bytes are cached payload:
// what is stored, not what the media's full size would be.
type CacheUsage struct {
	// Media is every cached media byte, each object once.
	Media int64
	// Categories are Media by category, Unattributed aside.
	Categories [StorageCategories]int64
	// Unattributed is media cached before objects were recorded, whose
	// owner the backfill has not found (yet, while Backfilled is false).
	Unattributed int64
	Backfilled   bool
	// Chats are the chats objects are attributed to, largest first. An
	// object several chats show counts in each of them; Shared is the
	// bytes of such objects, and NoChat those of objects of no chat.
	Chats  []ChatUsage
	Shared int64
	NoChat int64
	// Database is the cache's files on disk; Free is the reusable space
	// inside them, which is not given back to the system by itself.
	Database int64
	Free     int64
}

// ChatUsage is one chat's part of the cache.
type ChatUsage struct {
	Chat       int64
	Bytes      int64
	Categories [StorageCategories]int64
}

// StorageUsageSource is a store that can tell what its cache holds.
type StorageUsageSource interface {
	StorageUsage(ctx context.Context) (CacheUsage, error)
}

// DiskUsage is the files under a path: their length, and the space the
// file system gave them where it tells (Allocated 0 otherwise).
type DiskUsage struct {
	Bytes, Allocated int64
	Files            int
}

// Volume is a file system KomaruGram keeps data on. Known is false where
// its capacity cannot be read; local totals are still right then.
type Volume struct {
	ID          string
	Path        string
	Total, Free int64
	Known       bool
	// Used is KomaruGram's data on it, each root once.
	Used int64
}

// StorageRoot is a directory KomaruGram owns and what it holds. Shared
// roots belong to no account.
type StorageRoot struct {
	Name   string
	Path   string
	Volume string
	Usage  DiskUsage
}

// AccountStorage is one account's data: its directory on disk and, when
// its store is open and unlocked, what its cache holds.
type AccountStorage struct {
	ID    string
	Files DiskUsage
	// Cache is known only with CacheKnown; Err tells why not, if it failed.
	Cache      CacheUsage
	CacheKnown bool
	Err        error
}

// StorageSnapshot is all of KomaruGram's local data at one time.
// Generation grows with every snapshot, so that a late one can be told.
type StorageSnapshot struct {
	Generation uint64
	Volumes    []Volume
	Roots      []StorageRoot
	Accounts   []AccountStorage
}

// StorageService gives snapshots of local data.
type StorageService interface {
	StorageSnapshot(ctx context.Context) (StorageSnapshot, error)
}
