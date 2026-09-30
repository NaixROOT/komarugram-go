// SPDX-License-Identifier: Unlicense OR MIT

package emojipacks

import (
	"bytes"
	"context"
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

// miniCells is how many cells miniList orders.
const miniCells = 43

// telegramSprite is a sprite of the layout of Telegram Desktop's sets with
// the rows given, each cell of its own color.
func telegramSprite(t *testing.T, rows int) []byte {
	t.Helper()
	cell, columns := TelegramLayout.Cell, TelegramLayout.Columns
	img := image.NewNRGBA(image.Rect(0, 0, columns*cell, rows*cell))
	for id := range rows * columns {
		at := image.Pt(id%columns, id/columns).Mul(cell)
		c := cellColor(id % 6)
		c.B = uint8(id)
		for y := range cell {
			for x := range cell {
				img.SetNRGBA(at.X+x, at.Y+y, c)
			}
		}
	}
	var out bytes.Buffer
	if err := (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(&out, img); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

// hosting serves the files given by their paths, and counts the requests.
func hosting(t *testing.T, files map[string][]byte) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		data, ok := files[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write(data)
	}))
	t.Cleanup(server.Close)
	return server, &requests
}

func TestPackOfAddressesIsInstalled(t *testing.T) {
	sprite := telegramSprite(t, 2)
	files := map[string][]byte{"/lib_ui/abc/emoji.txt": []byte(miniList), "/tdesktop/abc/emoji_1.webp": sprite}
	server, requests := hosting(t, files)
	list := OfficialFile{Name: "emoji.txt", URL: server.URL + "/lib_ui/abc/emoji.txt", Data: []byte(miniList)}
	images := []OfficialFile{{Name: "emoji_1.webp", URL: server.URL + "/tdesktop/abc/emoji_1.webp", Data: sprite}}
	p, err := OfficialPack(Pack{ID: "apple", Name: "Apple"}, list, images, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Telegram != nil || p.Sprites.List != "emoji.txt" || p.Sprites.Order != "" || p.DownloadSize() != int64(len(miniList)+len(sprite)) {
		t.Fatalf("the pack described: %+v", p)
	}
	// The catalog built in has no files: all of them come from the addresses.
	store := Open(filepath.Join(t.TempDir(), "emoji"))
	var last int64
	if err := store.Install(context.Background(), Official(), p, func(done, _ int64) { last = done }); err != nil {
		t.Fatal(err)
	}
	if n := requests.Load(); n != 2 {
		t.Errorf("%d requests, want the list and the sprite", n)
	}
	if last != p.DownloadSize() {
		t.Errorf("progress ended at %d of %d", last, p.DownloadSize())
	}
	set, err := store.Sprites("apple")
	if err != nil {
		t.Fatal(err)
	}
	if set.Count() != miniCells {
		t.Fatalf("%d emoji, want the %d of the list", set.Count(), miniCells)
	}
	// The order is that of the list: the thumbs up of the third tone is
	// the sixth cell.
	id, length := set.Match([]rune("👍🏽!"))
	if id != 5 || length != 2 {
		t.Fatalf("the thumbs up matched as %d, %d runes", id, length)
	}
	want := cellColor(5)
	want.B = 5
	if got := color.NRGBAModel.Convert(set.Image(id, 8).At(4, 4)); got != want {
		t.Errorf("its picture is %v, want the cell's %v", got, want)
	}

	// What an address gives is checked as what a catalog gives is.
	files["/tdesktop/abc/emoji_1.webp"] = telegramSprite(t, 1)
	if err := store.Install(context.Background(), Official(), p, nil); err == nil {
		t.Error("a sprite that is not the one described was installed")
	}
	delete(files, "/lib_ui/abc/emoji.txt")
	if err := store.Install(context.Background(), Official(), p, nil); err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("a list that is not there: %v", err)
	}
	if set, err := store.Sprites("apple"); err != nil || set.Count() != miniCells {
		t.Errorf("the pack installed before is gone after failures: %v", err)
	}
	if entries, _ := os.ReadDir(store.Dir()); len(entries) != 1 {
		t.Errorf("%d entries in the store after failures", len(entries))
	}
}

func TestPackOfAnArchiveAndAListIsInstalled(t *testing.T) {
	sprite := telegramSprite(t, 2)
	server, _ := hosting(t, map[string][]byte{"/emoji.txt": []byte(miniList)})
	list := OfficialFile{Name: "emoji.txt", URL: server.URL + "/emoji.txt", Data: []byte(miniList)}
	archive := zipOf(t, map[string][]byte{"config.json": []byte(`{"id": 2, "version": 8}`), "emoji_1.webp": sprite, "__MACOSX/._emoji_1.webp": []byte("junk")})
	p, err := OfficialPack(Pack{ID: "twemoji", Name: "Twemoji"}, list, nil, archive, &TelegramArchive{Channel: "tdhbcfiles", Post: 3224})
	if err != nil {
		t.Fatal(err)
	}
	if p.Telegram == nil || p.Telegram.Size != int64(len(archive)) || len(p.Sprites.Images) != 1 || p.DownloadSize() != int64(len(miniList)+len(archive)) {
		t.Fatalf("the pack described: %+v", p)
	}
	store := Open(filepath.Join(t.TempDir(), "emoji"))
	if err := store.Install(context.Background(), Official(), p, nil); !errors.Is(err, ErrNeedsTelegram) {
		t.Errorf("without a way to Telegram: %v", err)
	}
	if HasTelegram(Official()) || !HasTelegram(WithTelegram(Official(), serving(archive, nil))) {
		t.Error("HasTelegram does not tell a source with a way to Telegram from one without")
	}
	if err := store.Install(context.Background(), WithTelegram(Official(), serving(archive, nil)), p, nil); err != nil {
		t.Fatal(err)
	}
	set, err := store.Sprites("twemoji")
	if err != nil {
		t.Fatal(err)
	}
	if set.Count() != miniCells || set.Image(miniCells-1, 8) == nil {
		t.Errorf("%d emoji installed, want %d that draw", set.Count(), miniCells)
	}
}

func TestListOfAnotherVersionThanTheSprites(t *testing.T) {
	list := OfficialFile{Name: "emoji.txt", URL: "https://example.com/emoji.txt", Data: []byte(miniList)}
	image := func(rows int) []OfficialFile {
		return []OfficialFile{{Name: "emoji_1.webp", URL: "https://example.com/emoji_1.webp", Data: telegramSprite(t, rows)}}
	}
	// The list's emoji end in the second row: sprites of three rows are of
	// a version with more emoji, and those of one row of one with fewer.
	for _, rows := range []int{1, 3} {
		if _, err := OfficialPack(Pack{ID: "apple", Name: "Apple"}, list, image(rows), nil, nil); err == nil {
			t.Errorf("sprites of %d rows were taken for a list of %d emoji", rows, miniCells)
		}
	}
	if _, err := OfficialPack(Pack{ID: "apple", Name: "Apple"}, list, image(2), nil, nil); err != nil {
		t.Errorf("sprites of the list's version: %v", err)
	}
	other := list
	other.Data = []byte("hello")
	if _, err := OfficialPack(Pack{ID: "apple", Name: "Apple"}, other, image(2), nil, nil); err == nil {
		t.Error("a file that is not the list was taken")
	}
}

func TestCatalogBuiltIn(t *testing.T) {
	t.Setenv(CatalogEnv, "")
	src := SourceFromEnv()
	if src == nil || src != Official() {
		t.Fatalf("the catalog without the variable is %v, want the one built in", src)
	}
	packs, err := ReadIndex(context.Background(), src)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, p := range packs {
		ids = append(ids, p.ID)
	}
	if got := strings.Join(ids, " "); got != "apple android twemoji joypixels" {
		t.Fatalf("the packs built in: %s", got)
	}
	// One commit of each repository is pinned, by its full hash.
	commits := map[string]bool{}
	for _, p := range packs {
		if p.Kind != KindSprites || p.Sprites.List != "emoji.txt" {
			t.Errorf("%s: %+v", p.ID, p.Sprites)
		}
		if p.Sprites.Cell != TelegramLayout.Cell || p.Sprites.Columns != TelegramLayout.Columns || p.Sprites.Rows != TelegramLayout.Rows {
			t.Errorf("%s is not of the layout of Telegram Desktop's sets", p.ID)
		}
		if (p.ID == "apple") != (p.Telegram == nil) {
			t.Errorf("%s: its archive is %+v", p.ID, p.Telegram)
		}
		if p.Telegram != nil && p.Telegram.Channel != "tdhbcfiles" {
			t.Errorf("%s is downloaded from @%s", p.ID, p.Telegram.Channel)
		}
		for _, f := range p.Files {
			if f.Archive {
				continue
			}
			for _, repository := range []string{"https://raw.githubusercontent.com/telegramdesktop/tdesktop/", "https://raw.githubusercontent.com/desktop-app/lib_ui/"} {
				if rest, ok := strings.CutPrefix(f.URL, repository); ok {
					commit, _, _ := strings.Cut(rest, "/")
					if len(commit) != 40 {
						t.Errorf("%s: %s is not at a commit", p.ID, f.URL)
					}
					commits[repository+commit] = true
				}
			}
			if !strings.HasPrefix(f.URL, "https://raw.githubusercontent.com/") {
				t.Errorf("%s: %s is downloaded from %q", p.ID, f.Name, f.URL)
			}
		}
	}
	if len(commits) != 2 {
		t.Errorf("the files are at %d commits, want one of each of the two repositories: %v", len(commits), commits)
	}
	// Nothing but its index is in it.
	if _, err := src.Open(context.Background(), "apple/emoji_1.webp"); err == nil {
		t.Error("the catalog built in gave a file")
	}

	dir := t.TempDir()
	t.Setenv(CatalogEnv, dir)
	if got := SourceFromEnv(); got == nil || got.String() != dir {
		t.Errorf("the catalog with the variable is %v", got)
	}
}

func TestAddressesAndListsAreValidated(t *testing.T) {
	list := OfficialFile{Name: "emoji.txt", URL: "https://example.com/emoji.txt", Data: []byte(miniList)}
	images := []OfficialFile{{Name: "emoji_1.webp", URL: "https://example.com/emoji_1.webp", Data: telegramSprite(t, 2)}}
	p, err := OfficialPack(Pack{ID: "apple", Name: "Apple"}, list, images, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for name, change := range map[string]func(*Pack){
		"a file on this computer": func(p *Pack) { p.Files[1].URL = "file:///etc/passwd" },
		"an address of no host":   func(p *Pack) { p.Files[1].URL = "https:///emoji_1.webp" },
		"a path":                  func(p *Pack) { p.Files[1].URL = "../emoji_1.webp" },
		"an address and an archive": func(p *Pack) {
			p.Files[1].Archive = true
			p.Telegram = &TelegramArchive{Channel: "tdhbcfiles", Post: 1, Size: 1, SHA256: p.Files[1].SHA256}
		},
		"an order and a list":     func(p *Pack) { p.Sprites.Order = "emoji.txt" },
		"no order and no list":    func(p *Pack) { p.Sprites.List = "" },
		"a list not of its files": func(p *Pack) { p.Sprites.List = "other.txt" },
	} {
		bad := p
		bad.Files = append([]File(nil), p.Files...)
		sprites := *p.Sprites
		bad.Sprites = &sprites
		change(&bad)
		if bad.Validate() == nil {
			t.Errorf("%s: valid", name)
		}
	}
}
