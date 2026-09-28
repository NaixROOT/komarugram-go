// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"gio-mw/exp"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
	"time"

	"gio-mw/defaults"
	"gio-mw/defaults/schemes"
	"gio-mw/exp/appearance"
	"gio-mw/wdk"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
	"komarugram/internal/messenger/preferences"
	"komarugram/internal/miniappprefs"
	"komarugram/internal/motion"
	"komarugram/pkg/miniapp"
	"komarugram/pkg/player"
)

type staticAccounts []model.AccountInfo

func (a staticAccounts) All() []model.AccountInfo { return a }
func (staticAccounts) Open(string)                {}
func (staticAccounts) Add()                       {}
func (staticAccounts) LogOut(string)              {}
func (staticAccounts) Subscribe(func()) func()    { return func() {} }

// TestRenderSettingsAccounts draws the main settings page with a list of
// saved accounts and saves a screenshot, for looking at it. SETTINGS_SECTION=appearance,
// privacy or integrations draws that section instead:
//
//	SETTINGS_PNG=/tmp/settings.png go test ./internal/messenger/ui -run RenderSettingsAccounts
func TestRenderSettingsAccounts(t *testing.T) {
	out := os.Getenv("SETTINGS_PNG")
	if out == "" {
		t.Skip("set SETTINGS_PNG to a file name")
	}
	photo := image.NewRGBA(image.Rect(0, 0, 160, 160))
	for y := 0; y < 160; y++ {
		for x := 0; x < 160; x++ {
			photo.Set(x, y, color.RGBA{uint8(x), 90, uint8(y), 255})
		}
	}
	accounts := staticAccounts{
		{ID: "1", UserID: 1, Name: "Ада Лавлейс", Username: "ada", Avatar: photo},
		{ID: "2", UserID: 2, Name: "Чарльз Бэббидж", Phone: "+44 20 0000 0000", Open: true},
		{ID: "3", UserID: 3, Name: "Аккаунт 3"},
	}
	p := newSettingsPage(motion.New(func() {}), miniappprefs.New(miniapp.Ephemeral), nil, func() {}, accounts,
		func() string { return "1" }, themeAuto, "ru", func(themeMode) {}, func(string) {})
	var images imageOps
	p.images = &images
	p.composerStyle = func() preferences.ComposerStyle { return preferences.ComposerFloating }
	p.composerBlur = func() bool { return true }
	size := image.Pt(900, 700)
	if os.Getenv("SETTINGS_SECTION") == "appearance" {
		p.section = settingsAppearance
		size.Y = 1000
	}
	if os.Getenv("SETTINGS_SECTION") == "privacy" {
		p.section = settingsPrivacy
		size.Y = 1600
		ghost := preferences.Ghost{ReadOnInteract: true}
		p.ghost = func() preferences.Ghost { return ghost }
		p.setGhost = func(g preferences.Ghost) { ghost = g }
		keep := preferences.Keep{Deleted: true, Edits: true}
		p.keep = func() preferences.Keep { return keep }
		p.setKeep = func(k preferences.Keep) { keep = k }
	}
	if os.Getenv("SETTINGS_SECTION") == "integrations" {
		p.section = settingsIntegrations
		size.Y = 1100
		// A VLC the user pointed at, and a file that is not mpv.
		custom := map[player.Kind]string{player.VLC: "/var/lib/flatpak/exports/bin/org.videolan.VLC", player.MPV: "/usr/bin/ls"}
		p.players.paths = func() map[player.Kind]string { return custom }
		p.invalidate = func() {}
		p.refreshPrograms()
		// The programs answer in the background.
		ops := new(op.Ops)
		for range 100 {
			ops.Reset()
			gtx := layout.Context{Ops: ops, Now: time.Now(), Constraints: layout.Exact(size), Metric: unit.Metric{PxPerDp: 1.25, PxPerSp: 1.25}, Values: map[string]any{}}
			wdk.InitMaterialThemeInContext(gtx, defaults.NewTheme(gtx, schemes.SchemeBaselineLight()))
			p.Update(gtx, themeAuto, "ru")
			if p.browser.detected != nil && p.players.programs[player.VLC].customAbout != nil && p.players.programs[player.MPV].customAbout != nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
	win, err := headless.NewWindow(size.X, size.Y)
	if err != nil {
		t.Fatal(err)
	}
	defer win.Release()
	ops := new(op.Ops)
	gtx := layout.Context{Ops: ops, Now: time.Now(), Constraints: layout.Exact(size), Metric: unit.Metric{PxPerDp: 1.25, PxPerSp: 1.25}, Values: map[string]any{}}
	wdk.InitMaterialThemeInContext(gtx, defaults.NewTheme(gtx, schemes.SchemeBaselineLight()))
	images.BeginFrame()
	// The root surface, which sections drawn by exp widgets read.
	exp.Background(gtx)
	p.Update(gtx, themeAuto, "ru")
	p.Layout(gtx, themeAuto, appearance.Light, false, localization.For("ru"))
	images.EndFrame()
	if err := win.Frame(ops); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rectangle{Max: size})
	if err := win.Screenshot(img); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
