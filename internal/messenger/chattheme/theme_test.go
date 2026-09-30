// SPDX-License-Identifier: Unlicense OR MIT

package chattheme

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"image"
	"image/color"
	"image/png"
	"komarugram/internal/messenger/model"
	"testing"
)

func TestDesktopThemeArchiveAliasesAndTile(t *testing.T) {
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	w, _ := z.Create("colors.tdesktop-theme")
	w.Write([]byte("// color aliases\nbase: #102030; msgInBg: base; historyTextInFg: #ffffff; msgOutBg: #405060; historyBg: #203040;"))
	w, _ = z.Create("tiled.png")
	im := image.NewRGBA(image.Rect(0, 0, 2, 2))
	im.Set(0, 0, color.White)
	png.Encode(w, im)
	z.Close()
	theme, err := ImportDesktop(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if theme.Light.Incoming != 0x102030ff || !theme.Light.Dark || !theme.Light.Wallpaper.Tile || len(theme.Light.Wallpaper.Image) == 0 {
		t.Fatal(theme.Light)
	}
	for _, bad := range []string{"msgInBg: other;other: msgInBg;", "msgInBg: unknown;", "no palette"} {
		if _, e := ImportDesktop([]byte(bad)); e == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
	var missing bytes.Buffer
	zw := zip.NewWriter(&missing)
	zw.Create("../../colors.tdesktop-theme")
	zw.Close()
	if _, e := ImportDesktop(missing.Bytes()); e == nil {
		t.Fatal("accepted non-root palette")
	}
}
func TestWallpaperGradientRotationAndPattern(t *testing.T) {
	a := Gradient([]uint32{0xff0000, 0x0000ff}, 5, 5, 0, 0)
	if a.RGBAAt(2, 0).R != 255 || a.RGBAAt(2, 4).B != 255 {
		t.Fatal("gradient endpoints")
	}
	b := Gradient([]uint32{0xff0000, 0x0000ff}, 5, 5, 180, 0)
	if b.RGBAAt(2, 0).B != 255 {
		t.Fatal("gradient rotation")
	}
	pattern := image.NewRGBA(image.Rect(0, 0, 2, 2))
	pattern.SetRGBA(0, 0, color.RGBA{A: 255})
	var encoded bytes.Buffer
	png.Encode(&encoded, pattern)
	render := func(intensity int) *image.RGBA {
		b, err := Prepare(&model.ChatWallpaper{Colors: []uint32{0x808080}, Pattern: true, Intensity: intensity}, encoded.Bytes())
		if err != nil {
			t.Fatal(err)
		}
		return b.Render(image.Pt(2, 2))
	}
	// A shape darkens the colors under it as a soft light does; over dark
	// wallpapers the colors show only through the shapes.
	positive, inverse := render(100), render(-100)
	if positive.RGBAAt(0, 0).R != 64 || positive.RGBAAt(1, 1).R != 128 || inverse.RGBAAt(0, 0).R != 128 || inverse.RGBAAt(1, 1).R != 0 {
		t.Fatal("pattern intensity/inversion", positive.Pix, inverse.Pix)
	}
}
func TestTGVSafelyDecodesGzipSVG(t *testing.T) {
	var b bytes.Buffer
	z := gzip.NewWriter(&b)
	z.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 10"><path d="M0 0H20V10H0Z" fill="#000000"/></svg>`))
	z.Close()
	im, err := Decode(b.Bytes())
	if err != nil || im.Bounds().Dx() != 20 {
		t.Fatal(im, err)
	}
	if _, err := Decode([]byte(`<svg viewBox="0 0 0 1"></svg>`)); err == nil {
		t.Fatal("zero SVG dimensions")
	}
	if _, err := Decode(make([]byte, MaxBytes+1)); err == nil {
		t.Fatal("unbounded input")
	}
}

func TestAccentColorsThemeButGrays(t *testing.T) {
	const pink = 0xd46c99
	day := Style(PresetDay, 0xff000000|pink)
	if day.Incoming != 0xffffffff || day.Text != 0x000000ff {
		t.Fatal("the accent changed white bubbles or text", day)
	}
	// Day's outgoing bubble is blue, as its accent: it turns pink.
	if h := hsvOf(day.Outgoing[0]).h; h < 300 && h > 20 {
		t.Fatalf("outgoing hue %v", h)
	}
	// Tinted's incoming bubble is of its accent's hue too.
	tinted := Style(PresetTinted, 0xff000000|pink)
	if h := hsvOf(tinted.Incoming >> 8).h; h < 300 && h > 20 {
		t.Fatalf("incoming hue %v", h)
	}
	// Classic's green bubbles are far from its blue accent.
	if classic := Style(PresetClassic, 0xff000000|pink); classic.Outgoing[0] != 0xeffdde || classic.Wallpaper == nil || len(classic.Wallpaper.Image) == 0 {
		t.Fatal("Classic", classic)
	}
	if Style(PresetApp, pink) != nil {
		t.Fatal("the application's colors have a style")
	}
}

func TestThemeStyleKeepsTextReadable(t *testing.T) {
	dark := ThemeStyle(PresetDay, 0x112233, 0, []uint32{0x102030, 0x304050}, false)
	light := ThemeStyle(PresetDay, 0x112233, 0, []uint32{0xfff0cb}, false)
	if dark.OutText != 0xffffffff || light.OutText != 0x000000ff || dark.Wallpaper != nil {
		t.Fatal(dark, light)
	}
}
