// SPDX-License-Identifier: Unlicense OR MIT

package emojipacks

import (
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

// testLayout is a grid of two columns and two rows of cells of 8 pixels.
var testLayout = Sprites{Cell: 8, Columns: 2, Rows: 2}

// cellColor is the color of cell id of the test's sprites.
func cellColor(id int) color.NRGBA {
	return color.NRGBA{R: uint8(40 * (id + 1)), G: uint8(255 - 40*id), B: 0x20, A: 0xff}
}

// writeSprites writes the sprites of a set of n emoji, each cell of its
// own color, into dir, and returns their paths.
func writeSprites(t *testing.T, dir string, n int) []string {
	t.Helper()
	per := testLayout.Columns * testLayout.Rows
	var paths []string
	for first := 0; first < n; first += per {
		rows := testLayout.Rows
		if left := n - first; left < per {
			rows = (left + testLayout.Columns - 1) / testLayout.Columns
		}
		img := image.NewNRGBA(image.Rect(0, 0, testLayout.Columns*testLayout.Cell, rows*testLayout.Cell))
		for id := first; id < min(n, first+per); id++ {
			at := image.Pt((id-first)%testLayout.Columns, (id-first)/testLayout.Columns).Mul(testLayout.Cell)
			for y := range testLayout.Cell {
				for x := range testLayout.Cell {
					img.SetNRGBA(at.X+x, at.Y+y, cellColor(id))
				}
			}
		}
		path := filepath.Join(dir, "emoji_"+string(rune('1'+len(paths)))+".png")
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		f.Close()
		paths = append(paths, path)
	}
	return paths
}

// testOrder are the emoji of the test's set: a face, a heart that is an
// emoji on its own, a thumbs up and its tone, a keycap, a sign that is an
// emoji only when asked to be, and a pair with another spelling.
var testOrder = [][]string{
	{"😀"},
	{"❤"},
	{"👍"},
	{"👍🏽"},
	{"1⃣"},
	{"©"},
	{"👫🏻", "👩🏻‍🤝‍👨🏻"},
}

// testCatalog makes a catalog of a sprite pack and a font pack, and
// returns its directory and the packs.
func testCatalog(t *testing.T) (dir string, sprites, font Pack) {
	t.Helper()
	dir = t.TempDir()
	images := writeSprites(t, t.TempDir(), len(testOrder))
	sprites, err := BuildSprites(dir, Pack{ID: "sprites", Name: "Sprites", License: "CC0"}, testLayout, images, testOrder)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "My Emoji (v2).ttf")
	if err := os.WriteFile(file, []byte("a font, for all the store knows"), 0o600); err != nil {
		t.Fatal(err)
	}
	font, err = BuildFont(dir, Pack{ID: "font", Name: "Font"}, file)
	if err != nil {
		t.Fatal(err)
	}
	return dir, sprites, font
}

func TestCatalogIsReadBack(t *testing.T) {
	dir, sprites, font := testCatalog(t)
	packs, err := ReadIndex(context.Background(), NewSource(dir))
	if err != nil {
		t.Fatal(err)
	}
	if len(packs) != 2 || packs[0].ID != "font" || packs[1].ID != "sprites" {
		t.Fatalf("packs read: %+v", packs)
	}
	if packs[1].Revision() != sprites.Revision() || packs[0].Revision() != font.Revision() {
		t.Error("the packs read are not the packs built")
	}
	if font.Font != "My-Emoji--v2-.ttf" {
		t.Errorf("the font's file is named %q", font.Font)
	}
	if len(sprites.Files) != 3 || sprites.Sprites.Order != "order.txt" || len(sprites.Sprites.Images) != 2 {
		t.Errorf("the sprite pack: %+v", sprites)
	}
	// A pack built again takes the place of the one of its name.
	other := filepath.Join(t.TempDir(), "other.ttf")
	os.WriteFile(other, []byte("another font"), 0o600)
	if _, err := BuildFont(dir, Pack{ID: "font", Name: "Font 2"}, other); err != nil {
		t.Fatal(err)
	}
	packs, _ = ReadIndex(context.Background(), NewSource(dir))
	if len(packs) != 2 || packs[0].Name != "Font 2" || packs[0].Font != "other.ttf" {
		t.Errorf("packs after one was built again: %+v", packs)
	}
	if _, err := os.Stat(filepath.Join(dir, "font", font.Font)); !errors.Is(err, os.ErrNotExist) {
		t.Error("a pack built again kept the file it had")
	}
	if err := RemoveFromCatalog(dir, "sprites"); err != nil {
		t.Fatal(err)
	}
	packs, _ = ReadIndex(context.Background(), NewSource(dir))
	for _, p := range packs {
		if p.ID == "sprites" {
			t.Error("a pack removed from the catalog is still in its index")
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "sprites")); !errors.Is(err, os.ErrNotExist) {
		t.Error("a pack removed from the catalog left its files")
	}
}

func TestIndexOfAnotherVersionOrBadPacks(t *testing.T) {
	dir := t.TempDir()
	write := func(index Index) {
		data, _ := json.Marshal(index)
		os.WriteFile(filepath.Join(dir, "index.json"), data, 0o600)
	}
	if _, err := ReadIndex(context.Background(), NewSource(dir)); err == nil {
		t.Error("no error for a catalog without an index")
	}
	if _, err := ReadIndex(context.Background(), nil); !errors.Is(err, ErrNoCatalog) {
		t.Errorf("without a catalog: %v", err)
	}
	write(Index{Version: IndexVersion + 1})
	if _, err := ReadIndex(context.Background(), NewSource(dir)); err == nil {
		t.Error("no error for an index of a newer version")
	}
	good := File{Name: "a.ttf", Size: 1, SHA256: strings.Repeat("0", 64)}
	write(Index{Version: IndexVersion, Packs: []Pack{
		{ID: "ok", Name: "OK", Kind: KindFont, Font: "a.ttf", Files: []File{good}},
		{ID: "ok", Name: "Twice", Kind: KindFont, Font: "a.ttf", Files: []File{good}},
		{ID: "../up", Name: "Up", Kind: KindFont, Font: "a.ttf", Files: []File{good}},
		{ID: "path", Name: "Path", Kind: KindFont, Font: "../a.ttf", Files: []File{{Name: "../a.ttf", Size: 1, SHA256: good.SHA256}}},
		{ID: "sub", Name: "Sub", Kind: KindFont, Font: "d/a.ttf", Files: []File{{Name: "d/a.ttf", Size: 1, SHA256: good.SHA256}}},
		{ID: "manifest", Name: "Manifest", Kind: KindFont, Font: "pack.json", Files: []File{{Name: "pack.json", Size: 1, SHA256: good.SHA256}}},
		{ID: "nohash", Name: "No hash", Kind: KindFont, Font: "a.ttf", Files: []File{{Name: "a.ttf", Size: 1}}},
		{ID: "huge", Name: "Huge", Kind: KindFont, Font: "a.ttf", Files: []File{{Name: "a.ttf", Size: MaxPackSize + 1, SHA256: good.SHA256}}},
		{ID: "kind", Name: "Kind", Kind: "video", Files: []File{good}},
		{ID: "nofont", Name: "No font", Kind: KindFont, Font: "b.ttf", Files: []File{good}},
		{ID: "nosprites", Name: "No sprites", Kind: KindSprites, Files: []File{good}},
		{ID: "noname", Kind: KindFont, Font: "a.ttf", Files: []File{good}},
	}})
	packs, err := ReadIndex(context.Background(), NewSource(dir))
	if err != nil {
		t.Fatal(err)
	}
	if len(packs) != 1 || packs[0].Name != "OK" {
		t.Errorf("packs taken from an index of bad ones: %+v", packs)
	}
}

func TestInstallRemoveAndInstallAgain(t *testing.T) {
	catalog, sprites, font := testCatalog(t)
	src := NewSource(catalog)
	store := Open(filepath.Join(t.TempDir(), "emoji"))
	if got := store.Installed(); len(got) != 0 {
		t.Fatalf("installed before anything: %v", got)
	}
	var last, total int64
	if err := store.Install(context.Background(), src, sprites, func(done, all int64) { last, total = done, all }); err != nil {
		t.Fatal(err)
	}
	if last != sprites.Size() || total != sprites.Size() {
		t.Errorf("progress ended at %d of %d, the pack is %d", last, total, sprites.Size())
	}
	if err := store.Install(context.Background(), src, font, nil); err != nil {
		t.Fatal(err)
	}
	installed := store.Installed()
	if len(installed) != 2 || installed[0].ID != "font" || installed[1].ID != "sprites" {
		t.Fatalf("installed: %+v", installed)
	}
	if data, err := os.ReadFile(store.Path("font", font.Font)); err != nil || string(data) != "a font, for all the store knows" {
		t.Errorf("the installed font: %q, %v", data, err)
	}
	set, err := store.Sprites("sprites")
	if err != nil {
		t.Fatal(err)
	}
	if again, _ := store.Sprites("sprites"); again != set {
		t.Error("the sprites were opened twice")
	}
	if _, err := store.Sprites("font"); err == nil {
		t.Error("a font pack gave sprites")
	}
	// Another window's store of the directory is the same store.
	if Open(store.Dir()) != store {
		t.Error("the directory was opened as another store")
	}

	if err := store.Remove("sprites"); err != nil {
		t.Fatal(err)
	}
	if _, ok := store.Pack("sprites"); ok {
		t.Error("a removed pack is still installed")
	}
	if _, err := store.Sprites("sprites"); err == nil {
		t.Error("a removed pack still gives sprites")
	}
	if err := store.Install(context.Background(), src, sprites, nil); err != nil {
		t.Fatalf("installing a removed pack again: %v", err)
	}
	if again, err := store.Sprites("sprites"); err != nil || again == set {
		t.Errorf("sprites of a pack installed again: the old ones %v, %v", again == set, err)
	}
	if entries, _ := os.ReadDir(store.Dir()); len(entries) != 2 {
		t.Errorf("the store's directory has %d entries, want the two packs", len(entries))
	}
}

func TestInstallChecksWhatItDownloads(t *testing.T) {
	catalog, sprites, _ := testCatalog(t)
	src := NewSource(catalog)
	dir := filepath.Join(t.TempDir(), "emoji")
	store := Open(dir)
	if err := store.Install(context.Background(), src, sprites, nil); err != nil {
		t.Fatal(err)
	}
	order := filepath.Join(catalog, "sprites", "order.txt")
	original, _ := os.ReadFile(order)
	for name, damage := range map[string]func(){
		"another content": func() { os.WriteFile(order, []byte(strings.Repeat("x", len(original))), 0o600) },
		"a longer file":   func() { os.WriteFile(order, append(original, 'x'), 0o600) },
		"a shorter file":  func() { os.WriteFile(order, original[:len(original)-1], 0o600) },
		"a missing file":  func() { os.Remove(order) },
	} {
		damage()
		if err := store.Install(context.Background(), src, sprites, nil); err == nil {
			t.Errorf("%s: installed", name)
		}
		// The pack installed before stays, and nothing is left behind.
		if _, ok := store.Pack("sprites"); !ok {
			t.Errorf("%s: the pack installed before is gone", name)
		}
		if entries, _ := os.ReadDir(dir); len(entries) != 1 {
			t.Errorf("%s: %d entries in the store, want the one pack", name, len(entries))
		}
		os.WriteFile(order, original, 0o600)
	}
	// What the catalog says of a pack is checked before anything is read.
	bad := sprites
	bad.Files = append([]File(nil), sprites.Files...)
	bad.Files[0].Name = "../outside.png"
	if err := store.Install(context.Background(), src, bad, nil); err == nil {
		t.Error("a pack with a file outside its directory was installed")
	}
	if err := store.Install(context.Background(), nil, sprites, nil); !errors.Is(err, ErrNoCatalog) {
		t.Errorf("without a catalog: %v", err)
	}
	if err := Open("").Install(context.Background(), src, sprites, nil); err == nil {
		t.Error("a store without a directory installed a pack")
	}
}

func TestInstallOverHTTPAndCancel(t *testing.T) {
	catalog, sprites, _ := testCatalog(t)
	var requests atomic.Int32
	files := http.FileServer(http.Dir(catalog))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		files.ServeHTTP(w, r)
	}))
	defer server.Close()
	src := NewSource(server.URL + "/")
	packs, err := ReadIndex(context.Background(), src)
	if err != nil || len(packs) != 2 {
		t.Fatalf("the index over HTTP: %v, %v", packs, err)
	}
	store := Open(filepath.Join(t.TempDir(), "emoji"))
	if err := store.Install(context.Background(), src, sprites, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Sprites("sprites"); err != nil {
		t.Errorf("sprites installed over HTTP: %v", err)
	}
	if n := requests.Load(); n != int32(1+len(sprites.Files)) {
		t.Errorf("%d requests, want the index and the %d files", n, len(sprites.Files))
	}
	missing := sprites
	missing.ID = "gone"
	if err := store.Install(context.Background(), src, missing, nil); err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("a pack the server does not have: %v", err)
	}

	// A download stopped leaves the pack as it was.
	ctx, cancel := context.WithCancel(context.Background())
	err = store.Install(ctx, src, sprites, func(done, total int64) { cancel() })
	if !errors.Is(err, context.Canceled) {
		t.Errorf("a cancelled download: %v", err)
	}
	if _, ok := store.Pack("sprites"); !ok {
		t.Error("a cancelled download removed the pack installed before")
	}
	if entries, _ := os.ReadDir(store.Dir()); len(entries) != 1 {
		t.Errorf("a cancelled download left %d entries in the store", len(entries))
	}
}

func TestOnePackIsInstalledAtATime(t *testing.T) {
	catalog, sprites, _ := testCatalog(t)
	store := Open(filepath.Join(t.TempDir(), "emoji"))
	started, release := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	first := true
	go func() {
		done <- store.Install(context.Background(), NewSource(catalog), sprites, func(int64, int64) {
			if first {
				first = false
				close(started)
				<-release
			}
		})
	}()
	<-started
	if err := store.Install(context.Background(), NewSource(catalog), sprites, nil); !errors.Is(err, ErrBusy) {
		t.Errorf("a second download of the pack: %v", err)
	}
	if err := store.Remove("sprites"); !errors.Is(err, ErrBusy) {
		t.Errorf("removing the pack while it downloads: %v", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if err := store.Remove("sprites"); err != nil {
		t.Errorf("removing the pack once it is installed: %v", err)
	}
}

func openTestSprites(t *testing.T) *SpriteSet {
	t.Helper()
	catalog, sprites, _ := testCatalog(t)
	set, err := OpenSprites(filepath.Join(catalog, "sprites"), sprites)
	if err != nil {
		t.Fatal(err)
	}
	return set
}

func TestSpritesMatch(t *testing.T) {
	set := openTestSprites(t)
	if set.Count() != len(testOrder) {
		t.Fatalf("%d emoji, want %d", set.Count(), len(testOrder))
	}
	for _, c := range []struct {
		text       string
		id, length int
	}{
		{"😀", 0, 1},
		{"😀abc", 0, 1},
		{"😀️", 0, 2},
		// A heart is an emoji with and without the selector.
		{"❤", 1, 1},
		{"❤️!", 1, 2},
		// The longest emoji wins: the thumbs up with its tone.
		{"👍", 2, 1},
		{"👍🏽", 3, 2},
		{"👍🏾", 2, 1},
		// A keycap, with the selector inside it or without.
		{"1️⃣", 4, 3},
		{"1⃣", 4, 2},
		{"1", 0, 0},
		{"12", 0, 0},
		// © is an emoji only with the selector.
		{"©", 0, 0},
		{"© 2026", 0, 0},
		{"©️", 5, 2},
		// Both spellings of the pair are one cell.
		{"👫🏻", 6, 2},
		{"👩🏻‍🤝‍👨🏻", 6, 7},
		// A sequence begun and not finished is not an emoji of the set.
		{"👩🏻‍🤝", 0, 0},
		{"abc", 0, 0},
		{"", 0, 0},
		{"️😀", 0, 0},
	} {
		id, length := set.Match([]rune(c.text))
		if length != c.length || (length > 0 && id != c.id) {
			t.Errorf("%q: emoji %d of %d runes, want %d of %d", c.text, id, length, c.id, c.length)
		}
	}
}

func TestSpritesImage(t *testing.T) {
	set := openTestSprites(t)
	for id := range testOrder {
		for _, size := range []int{4, 8, 20} {
			img := set.Image(id, size)
			if img == nil {
				t.Fatalf("emoji %d at %d pixels: no picture", id, size)
			}
			if img.Bounds() != image.Rect(0, 0, size, size) {
				t.Errorf("emoji %d at %d pixels is %v", id, size, img.Bounds())
			}
			// The middle of the picture is the cell's color, whatever the
			// scaling does at the edges.
			got := color.NRGBAModel.Convert(img.At(size/2, size/2)).(color.NRGBA)
			if want := cellColor(id); got != want {
				t.Errorf("emoji %d at %d pixels is %v, want its cell's %v", id, size, got, want)
			}
		}
	}
	for _, id := range []int{-1, len(testOrder), 1000} {
		if set.Image(id, 8) != nil {
			t.Errorf("a picture of emoji %d, which the set does not have", id)
		}
	}
	if set.Image(0, 0) != nil {
		t.Error("a picture of no size")
	}
	// Pictures decoded are kept, the last two of them.
	if set.decoded[0].img == nil || set.decoded[1].img == nil {
		t.Error("the sprites used last are not kept decoded")
	}
}

func TestSpritesTurnDownWrongFiles(t *testing.T) {
	catalog, sprites, _ := testCatalog(t)
	dir := filepath.Join(catalog, "sprites")
	order := filepath.Join(dir, "order.txt")
	original, _ := os.ReadFile(order)
	for name, text := range map[string]string{
		"more emoji than cells": strings.Repeat("😀\n", 9),
		"an emoji listed twice": "😀\n😀\n",
		"an empty line":         "😀\n\n👍\n",
		"no emoji":              "",
	} {
		if name == "more emoji than cells" {
			// Each line another emoji.
			var b strings.Builder
			for r := rune(0x1F600); r < 0x1F609; r++ {
				b.WriteString(string(r) + "\n")
			}
			text = b.String()
		}
		os.WriteFile(order, []byte(text), 0o600)
		if _, err := OpenSprites(dir, sprites); err == nil {
			t.Errorf("%s: the set opened", name)
		}
	}
	os.WriteFile(order, original, 0o600)
	wrong := sprites
	layout := *sprites.Sprites
	layout.Cell = 7
	wrong.Sprites = &layout
	if _, err := OpenSprites(dir, wrong); err == nil {
		t.Error("sprites that are not a grid of the cells described opened")
	}
	os.WriteFile(filepath.Join(dir, sprites.Sprites.Images[0]), []byte("not a picture"), 0o600)
	if _, err := OpenSprites(dir, sprites); err == nil {
		t.Error("a sprite that is not a picture opened")
	}
}

func TestReadDir(t *testing.T) {
	catalog, sprites, _ := testCatalog(t)
	p, err := ReadDir(filepath.Join(catalog, "sprites"))
	if err != nil || p.Revision() != sprites.Revision() {
		t.Errorf("a pack of a catalog: %v, %v", p.ID, err)
	}
	store := Open(filepath.Join(t.TempDir(), "emoji"))
	if err := store.Install(context.Background(), NewSource(catalog), sprites, nil); err != nil {
		t.Fatal(err)
	}
	p, err = ReadDir(store.PackDir("sprites"))
	if err != nil || p.Revision() != sprites.Revision() {
		t.Errorf("an installed pack: %v, %v", p.ID, err)
	}
	if _, err := ReadDir(t.TempDir()); err == nil {
		t.Error("a directory that is not a pack was read as one")
	}
}

func TestInstalledSpritesAreReadFromCells(t *testing.T) {
	catalog, sprites, _ := testCatalog(t)
	store := Open(filepath.Join(t.TempDir(), "emoji"))
	if err := store.Install(context.Background(), NewSource(catalog), sprites, nil); err != nil {
		t.Fatal(err)
	}
	set, err := store.Sprites("sprites")
	if err != nil {
		t.Fatal(err)
	}
	if set.cells == nil {
		t.Fatal("an installed pack has no cells cut out")
	}
	// The pictures come from the cells: the sprites are not decoded, and
	// may be gone.
	for _, name := range sprites.Sprites.Images {
		os.Remove(store.Path("sprites", name))
	}
	for id := range testOrder {
		img := set.Image(id, 8)
		if img == nil {
			t.Fatalf("emoji %d: no picture", id)
		}
		if got, want := color.NRGBAModel.Convert(img.At(4, 4)).(color.NRGBA), cellColor(id); got != want {
			t.Errorf("emoji %d is %v, want its cell's %v", id, got, want)
		}
	}
	if set.decoded[0].img != nil {
		t.Error("a sprite was decoded for a pack with cells")
	}
	// The catalog's own directory is left as published: no cells there.
	published, err := OpenSprites(filepath.Join(catalog, "sprites"), sprites)
	if err != nil || published.cells != nil {
		t.Errorf("a pack of a catalog: cells %v, %v", published.cells != nil, err)
	}
	if _, err := os.Stat(filepath.Join(catalog, "sprites", cellsName)); !errors.Is(err, os.ErrNotExist) {
		t.Error("cells were written into the catalog")
	}
}

func TestCellsOfAnotherPackAreNotUsed(t *testing.T) {
	catalog, sprites, _ := testCatalog(t)
	dir := filepath.Join(catalog, "sprites")
	if err := buildCells(dir, sprites); err != nil {
		t.Fatal(err)
	}
	if set, err := OpenSprites(dir, sprites); err != nil || set.cells == nil {
		t.Fatalf("cells just made: %v", err)
	}
	// The pack described otherwise is another pack.
	other := sprites
	other.Files = append([]File(nil), sprites.Files...)
	other.Files[0].SHA256 = strings.Repeat("0", 64)
	if set, err := OpenSprites(dir, other); err != nil || set.cells != nil {
		t.Errorf("cells of another revision: used %v, %v", set != nil && set.cells != nil, err)
	}
	// A file cut short is not cells, and the set draws from its sprites.
	path := filepath.Join(dir, cellsName)
	data, _ := os.ReadFile(path)
	os.WriteFile(path, data[:len(data)-3], 0o600)
	set, err := OpenSprites(dir, sprites)
	if err != nil || set.cells != nil {
		t.Fatalf("cells cut short: used %v, %v", set != nil && set.cells != nil, err)
	}
	if got, want := color.NRGBAModel.Convert(set.Image(2, 8).At(4, 4)).(color.NRGBA), cellColor(2); got != want {
		t.Errorf("without cells emoji 2 is %v, want %v", got, want)
	}
	// Cells that do not unpack leave the picture to the sprites too.
	for i := len(data) - 20; i < len(data); i++ {
		data[i] ^= 0xff
	}
	os.WriteFile(path, data, 0o600)
	set, err = OpenSprites(dir, sprites)
	if err != nil || set.cells == nil {
		t.Fatalf("cells with a bad end: %v", err)
	}
	last := len(testOrder) - 1
	if got, want := color.NRGBAModel.Convert(set.Image(last, 8).At(4, 4)).(color.NRGBA), cellColor(last); got != want {
		t.Errorf("with a cell that does not unpack emoji %d is %v, want %v", last, got, want)
	}
}
