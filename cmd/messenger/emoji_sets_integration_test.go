package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"komarugram/internal/messenger/emojipacks"
	"komarugram/internal/messenger/security"
	"komarugram/internal/messenger/tgstore"

	"github.com/gotd/td/telegram"
)

// Opt-in: downloads the emoji sets Telegram Desktop offers, from the
// channel it takes them from, into the directory the variable names, as
// set-<post>.zip. It uses the existing account's exclusive lock and only
// reads: the channel's name is resolved, three of its posts are asked for
// and their files downloaded. cmd/emoji-pack makes packs of them. With
// KOMARUGRAM_EMOJI_PACKS naming a catalog of such packs, they are then
// installed from Telegram into the directory's "installed".
func TestLiveEmojiSets(t *testing.T) {
	dir := os.Getenv("KOMARUGRAM_TEST_LIVE_EMOJI_SETS")
	if dir == "" {
		t.Skip("opt-in authorized account check: set KOMARUGRAM_TEST_LIVE_EMOJI_SETS to a directory")
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	store := tgstore.New(func() {})
	if err = store.Configure(ctx, a.ID, filepath.Join(t.TempDir(), "history"), protection); err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	err = manager.RunClient(ctx, a, nil, func(ctx context.Context, client *telegram.Client) error {
		store.Attach(client)
		// The posts of Telegram Desktop's chat_helpers/emoji_sets_manager.cpp:
		// Android, Twemoji and JoyPixels.
		for _, post := range []int{3223, 3224, 3225} {
			data, err := store.ChannelFile(ctx, "tdhbcfiles", post, nil)
			if err != nil {
				return fmt.Errorf("post %d: %w", post, err)
			}
			path := filepath.Join(dir, fmt.Sprintf("set-%d.zip", post))
			if err := os.WriteFile(path, data, 0o644); err != nil {
				return err
			}
			t.Logf("post %d: %d bytes in %s", post, len(data), path)
		}
		// With a catalog, its packs whose sprites are in Telegram are
		// installed the way the settings install them, into the directory.
		src := emojipacks.WithTelegram(emojipacks.SourceFromEnv(), store.ChannelFile)
		if src == nil {
			return nil
		}
		packs, err := emojipacks.ReadIndex(ctx, src)
		if err != nil {
			return err
		}
		installed := emojipacks.Open(filepath.Join(dir, "installed"))
		for _, p := range packs {
			if p.Telegram == nil {
				continue
			}
			start := time.Now()
			if err := installed.Install(ctx, src, p, nil); err != nil {
				return fmt.Errorf("pack %s: %w", p.ID, err)
			}
			set, err := installed.Sprites(p.ID)
			if err != nil {
				return fmt.Errorf("pack %s: %w", p.ID, err)
			}
			if set.Image(0, 32) == nil {
				return fmt.Errorf("pack %s: no picture of its first emoji", p.ID)
			}
			t.Logf("pack %s: %d emoji installed from @%s/%d in %v", p.ID, set.Count(), p.Telegram.Channel, p.Telegram.Post, time.Since(start))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
