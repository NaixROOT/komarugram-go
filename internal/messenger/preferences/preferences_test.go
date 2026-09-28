package preferences

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gio-mw/exp/powersave"

	"komarugram/pkg/miniapp"
	"komarugram/pkg/player"
)

func TestPersistsAndNotifies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	s, err := OpenPath(path)
	if err != nil {
		t.Fatal(err)
	}
	called := 0
	unsubscribe := s.Subscribe(func() { called++ })
	if err := s.SetTheme(ThemeDark); err != nil {
		t.Fatal(err)
	}
	if err := s.SetMotion(powersave.ModeOff, 25); err != nil {
		t.Fatal(err)
	}
	if err := s.SetMiniAppStorage(miniapp.PerApp); err != nil {
		t.Fatal(err)
	}
	if err := s.SetLanguage("en"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetLastAccount("account-b"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetComposer(ComposerClassic); err != nil {
		t.Fatal(err)
	}
	if err := s.SetComposerBlur(false); err != nil {
		t.Fatal(err)
	}
	if err := s.SetPlayer(player.VLC); err != nil {
		t.Fatal(err)
	}
	if err := s.SetPlayerPath(player.VLC, "/opt/vlc/vlc"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetBrowserPath("/opt/chromium/chrome"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetPlayer("totem"); err == nil {
		t.Fatal("an unknown player was accepted")
	}
	unsubscribe()
	if err := s.SetTheme(ThemeLight); err != nil {
		t.Fatal(err)
	}
	if called != 10 {
		t.Fatalf("notifications = %d, want 10", called)
	}

	loaded, err := OpenPath(path)
	if err != nil {
		t.Fatal(err)
	}
	want := Global{Theme: ThemeLight, Language: "en", LastAccountID: "account-b", MotionMode: powersave.ModeOff, LowBattery: 25, MiniAppStorage: miniapp.PerApp, Composer: ComposerClassic, Player: player.VLC, VLCPath: "/opt/vlc/vlc", BrowserPath: "/opt/chromium/chrome", Ghost: Ghost{ReadOnInteract: true}, Keep: Keep{Deleted: true, Edits: true}, Look: Look{BubbleRadius: BubbleRadiusMax, AvatarCorners: AvatarRound}}
	if got := loaded.Global(); !got.Equal(want) {
		t.Fatalf("loaded %+v, want %+v", got, want)
	}
}

func TestWindowLockOptionsPersistIndependently(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	s, err := OpenPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetWindowLock(17, true, false); err != nil {
		t.Fatal(err)
	}
	loaded, err := OpenPath(path)
	if err != nil {
		t.Fatal(err)
	}
	g := loaded.Global()
	if g.AutoLockMinutes != 17 || !g.LockOnMinimize || g.LockOnClose {
		t.Fatalf("settings: %+v", g)
	}
	if err := loaded.SetWindowLock(0, true, true); err != nil {
		t.Fatal(err)
	}
	g = loaded.Global()
	if g.AutoLockMinutes != 0 || !g.LockOnMinimize || !g.LockOnClose {
		t.Fatalf("independent triggers: %+v", g)
	}
	if err := loaded.SetWindowLock(121, false, false); err == nil {
		t.Fatal("invalid timeout accepted")
	}
}

func TestGlobalChangePreservesAccountNamespace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	original := `{"version":1,"global":{"theme":0,"language":"ru","motion_mode":0,"low_battery":20,"mini_app_storage":2},"accounts":{"a":{"compact":true}}}`
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	store, err := OpenPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetTheme(ThemeDark); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var saved fileData
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	var accountSettings map[string]bool
	if err := json.Unmarshal(saved.Accounts["a"], &accountSettings); err != nil {
		t.Fatal(err)
	}
	if !accountSettings["compact"] || len(accountSettings) != 1 {
		t.Fatalf("account settings were changed: %s", saved.Accounts["a"])
	}
}
