package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"komarugram/internal/messenger/model"
	"komarugram/internal/messenger/security"
	"komarugram/internal/messenger/tgstore"

	"github.com/gotd/td/telegram"
)

// Opt-in: uses the existing account's exclusive lock; never signs in, sends,
// marks history read, runs callbacks or changes online status.
func TestLiveReadonlyHistory(t *testing.T) {
	if os.Getenv("KOMARUGRAM_TEST_LIVE_HISTORY") != "1" {
		t.Skip("opt-in authorized account check")
	}
	protection, err := security.Open()
	if err != nil {
		t.Fatal(err)
	}
	if protection.Enabled() && !protection.State().Unlocked {
		t.Skip("account needs interactive TPM unlock")
	}
	manager, err := newManager(os.Getenv("KOMARUGRAM_PROXY"), protection)
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.Load(); err != nil {
		t.Fatal(err)
	}
	accounts := manager.Accounts()
	if len(accounts) == 0 {
		t.Fatal("no existing accounts")
	}
	a := accounts[0]
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	changed := make(chan struct{}, 1)
	store := tgstore.New(func() {
		select {
		case changed <- struct{}{}:
		default:
		}
	})
	if err = store.Configure(ctx, a.ID, filepath.Join(t.TempDir(), "history"), protection); err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	err = manager.RunClient(ctx, a, nil, func(ctx context.Context, client *telegram.Client) error {
		store.Attach(client)
		if err := store.Load(ctx, client.API()); err != nil {
			return err
		}
		chosen := map[model.ChatKind]bool{}
		tested := 0
		photos, animations, emoji := 0, 0, 0
		streams := 0
		channel := os.Getenv("KOMARUGRAM_TEST_MEDIA_CHANNEL")
		for _, chat := range store.Chats() {
			if channel != "" {
				if !strings.EqualFold(chat.Title, channel) {
					continue
				}
			} else if chosen[chat.Kind] || chat.Kind == model.KindGroup {
				continue
			}
			chosen[chat.Kind] = true
			store.OpenChat(chat.ID)
			for {
				h := store.History(chat.ID)
				if !h.LoadingOlder {
					if h.Err != nil {
						return h.Err
					}
					t.Logf("kind=%d messages=%d older=%v newer=%v", chat.Kind, len(h.Messages), h.HasOlder, h.HasNewer)
					for _, msg := range h.Messages {
						if msg.Kind == model.MessageVideo && msg.Media != nil && msg.Media.Size > 0 && streams == 0 {
							reader, size, close, e := store.MediaStream(ctx, msg)
							if e != nil {
								return e
							}
							nbytes := min(int64(65536), size)
							head, tail := make([]byte, nbytes), make([]byte, nbytes)
							_, e = reader.ReadAt(head, 0)
							if e == nil {
								_, e = reader.ReadAt(tail, size-nbytes)
							}
							close()
							if e != nil {
								return e
							}
							streams++
							if msg.Media.Thumbnail != nil {
								thumb := msg
								thumb.Media = msg.Media.Thumbnail
								thumb.Kind = model.MessagePhoto
								if _, e = store.Media(ctx, thumb); e != nil {
									return e
								}
							}
						}

						if msg.Media != nil && msg.Media.Size > 0 && msg.Media.Size < 8<<20 && (photos+animations < 2 || (msg.Kind == model.MessagePhoto && photos < 1)) {
							if _, e := store.Media(ctx, msg); e != nil {
								if channel != "" {
									return fmt.Errorf("media kind=%d size=%d: %w", msg.Kind, msg.Media.Size, e)
								}
								t.Logf("media kind=%d unavailable: %v", msg.Kind, e)
							} else {
								if msg.Kind == model.MessagePhoto {
									photos++
								} else {
									animations++
								}
							}
						}
						for _, entity := range msg.Entities {
							if entity.Kind == "emoji" && emoji == 0 {
								em, e := store.CustomEmoji(ctx, entity.DocumentID)
								if e == nil {
									_, e = store.Media(ctx, em)
								}
								if e != nil {
									t.Logf("custom emoji unavailable: %v", e)
								} else {
									emoji++
								}
							}
						}
					}
					break
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-changed:
				}
			}
			tested++
			if tested >= 3 || channel != "" {
				break
			}
		}
		if channel != "" && (tested == 0 || photos+animations+streams == 0) {
			t.Fatal("requested channel or media not found")
		}
		t.Logf("checked dialogs=%d photos=%d other_media=%d custom_emoji=%d streamed_videos=%d", tested, photos, animations, emoji, streams)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
