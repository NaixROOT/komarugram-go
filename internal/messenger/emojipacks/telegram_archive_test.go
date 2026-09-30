// SPDX-License-Identifier: Unlicense OR MIT

package emojipacks

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// zipOf makes a zip of the files given, by name.
func zipOf(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for name, data := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		f.Write(data)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

// archivedCatalog makes a catalog whose sprite pack has its sprites in an
// archive of Telegram's cloud, as the sets of Telegram Desktop are: the
// catalog keeps the order alone. It returns the catalog, the pack and the
// archive.
func archivedCatalog(t *testing.T) (catalog string, p Pack, archive []byte) {
	t.Helper()
	catalog, p, _ = testCatalog(t)
	files := map[string][]byte{"config.json": []byte(`{"id": 1, "version": 8}`), "__MACOSX/._emoji_1.png": []byte("junk")}
	for i, f := range p.Files {
		if f.Name == p.Sprites.Order {
			continue
		}
		path := filepath.Join(catalog, p.ID, f.Name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		files[f.Name] = data
		p.Files[i].Archive = true
		os.Remove(path)
	}
	archive = zipOf(t, files)
	sum := sha256.Sum256(archive)
	p.Telegram = &TelegramArchive{Channel: "tdhbcfiles", Post: 3224, Size: int64(len(archive)), SHA256: hex.EncodeToString(sum[:])}
	if err := putPack(catalog, p); err != nil {
		t.Fatal(err)
	}
	return catalog, p, archive
}

// serving returns a fetcher that gives archive for the post of the pack,
// telling of its progress in two parts, and counts what it was asked.
func serving(archive []byte, asked *[]string) TelegramFetcher {
	return func(_ context.Context, channel string, post int, progress func(done, total int64)) ([]byte, error) {
		if asked != nil {
			*asked = append(*asked, channel)
		}
		if channel != "tdhbcfiles" || post != 3224 {
			return nil, errors.New("no such post")
		}
		if progress != nil {
			progress(int64(len(archive)/2), int64(len(archive)))
			progress(int64(len(archive)), int64(len(archive)))
		}
		return archive, nil
	}
}

func TestPackOfATelegramArchive(t *testing.T) {
	catalog, p, archive := archivedCatalog(t)
	// The catalog is read back with the archive, and says what is to be
	// downloaded: the order and the archive, not the sprites unpacked.
	packs, err := ReadIndex(context.Background(), NewSource(catalog))
	if err != nil {
		t.Fatal(err)
	}
	var read Pack
	for _, c := range packs {
		if c.ID == p.ID {
			read = c
		}
	}
	if read.Telegram == nil || read.Telegram.Post != 3224 {
		t.Fatalf("the pack read back: %+v", read)
	}
	var order int64
	for _, f := range p.Files {
		if !f.Archive {
			order += f.Size
		}
	}
	if got, want := read.DownloadSize(), order+int64(len(archive)); got != want {
		t.Errorf("to download: %d bytes, want the order and the archive, %d", got, want)
	}

	store := Open(filepath.Join(t.TempDir(), "emoji"))
	var asked []string
	var last, total int64
	steps := 0
	src := WithTelegram(NewSource(catalog), serving(archive, &asked))
	if err := store.Install(context.Background(), src, read, func(done, all int64) { last, total, steps = done, all, steps+1 }); err != nil {
		t.Fatal(err)
	}
	if len(asked) != 1 {
		t.Errorf("Telegram was asked %d times, want once", len(asked))
	}
	if last != read.DownloadSize() || total != read.DownloadSize() || steps < 3 {
		t.Errorf("progress ended at %d of %d in %d steps, the download is %d", last, total, steps, read.DownloadSize())
	}
	set, err := store.Sprites(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if set.Count() != len(testOrder) || set.Image(2, 8) == nil {
		t.Error("the installed pack does not draw")
	}
	// Only the pack's files were taken out of the archive.
	entries, _ := os.ReadDir(store.PackDir(p.ID))
	for _, e := range entries {
		if e.Name() == "config.json" || e.IsDir() {
			t.Errorf("%s was taken out of the archive", e.Name())
		}
	}
}

func TestTelegramArchiveIsChecked(t *testing.T) {
	catalog, p, archive := archivedCatalog(t)
	store := Open(filepath.Join(t.TempDir(), "emoji"))
	src := NewSource(catalog)
	// Without an account there is no way to the archive.
	if err := store.Install(context.Background(), src, p, nil); !errors.Is(err, ErrNeedsTelegram) {
		t.Errorf("without a way to Telegram: %v", err)
	}
	if err := store.Install(context.Background(), WithTelegram(src, nil), p, nil); !errors.Is(err, ErrNeedsTelegram) {
		t.Errorf("with no fetcher: %v", err)
	}
	sprite := p.Sprites.Images[0]
	var spriteFile File
	for _, f := range p.Files {
		if f.Name == sprite {
			spriteFile = f
		}
	}
	for name, other := range map[string][]byte{
		"another archive":     zipOf(t, map[string][]byte{sprite: []byte("x")}),
		"not an archive":      []byte(strings.Repeat("x", len(archive))),
		"an archive cut off":  archive[:len(archive)-10],
		"an archive appended": append(append([]byte(nil), archive...), 0),
	} {
		if err := store.Install(context.Background(), WithTelegram(src, serving(other, nil)), p, nil); err == nil {
			t.Errorf("%s: installed", name)
		}
	}
	// An archive that is the one described, with a file that is not: the
	// catalog's word of the archive is taken, of the file it is not.
	for name, content := range map[string][]byte{
		"a file of another content": bytes.Repeat([]byte("x"), int(spriteFile.Size)),
		"a file of another size":    []byte("short"),
	} {
		other := zipOf(t, map[string][]byte{sprite: content, p.Sprites.Images[1]: []byte("y")})
		sum := sha256.Sum256(other)
		bad := p
		bad.Telegram = &TelegramArchive{Channel: "tdhbcfiles", Post: 3224, Size: int64(len(other)), SHA256: hex.EncodeToString(sum[:])}
		err := store.Install(context.Background(), WithTelegram(src, serving(other, nil)), bad, nil)
		if err == nil || !strings.Contains(err.Error(), sprite) {
			t.Errorf("%s: %v", name, err)
		}
	}
	missing := zipOf(t, map[string][]byte{"other.png": []byte("x")})
	sum := sha256.Sum256(missing)
	bad := p
	bad.Telegram = &TelegramArchive{Channel: "tdhbcfiles", Post: 3224, Size: int64(len(missing)), SHA256: hex.EncodeToString(sum[:])}
	if err := store.Install(context.Background(), WithTelegram(src, serving(missing, nil)), bad, nil); err == nil {
		t.Error("an archive without the pack's files: installed")
	}
	if got := store.Installed(); len(got) != 0 {
		t.Errorf("installed after failures: %v", got)
	}
	if entries, _ := os.ReadDir(store.Dir()); len(entries) != 0 {
		t.Errorf("%d entries left in the store after failures", len(entries))
	}
	// A failure of Telegram names the post.
	failing := func(context.Context, string, int, func(int64, int64)) ([]byte, error) {
		return nil, errors.New("FLOOD_WAIT")
	}
	if err := store.Install(context.Background(), WithTelegram(src, failing), p, nil); err == nil || !strings.Contains(err.Error(), "@tdhbcfiles/3224") || !strings.Contains(err.Error(), "FLOOD_WAIT") {
		t.Errorf("a failure of Telegram: %v", err)
	}
}

func TestArchiveIsDescribedWhole(t *testing.T) {
	_, p, _ := archivedCatalog(t)
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, change := range map[string]func(*Pack){
		"no archive for its files": func(p *Pack) { p.Telegram = nil },
		"no channel":               func(p *Pack) { t := *p.Telegram; t.Channel = ""; p.Telegram = &t },
		"a channel that is a link": func(p *Pack) { t := *p.Telegram; t.Channel = "t.me/x/../y"; p.Telegram = &t },
		"no post":                  func(p *Pack) { t := *p.Telegram; t.Post = 0; p.Telegram = &t },
		"no size":                  func(p *Pack) { t := *p.Telegram; t.Size = 0; p.Telegram = &t },
		"too large":                func(p *Pack) { t := *p.Telegram; t.Size = MaxArchiveSize + 1; p.Telegram = &t },
		"no hash":                  func(p *Pack) { t := *p.Telegram; t.SHA256 = ""; p.Telegram = &t },
		"a hash that is not hex":   func(p *Pack) { t := *p.Telegram; t.SHA256 = strings.Repeat("z", 64); p.Telegram = &t },
	} {
		bad := p
		change(&bad)
		if bad.Validate() == nil {
			t.Errorf("%s: valid", name)
		}
	}
}

func TestTelegramSpritesOfAZip(t *testing.T) {
	write := func(files map[string][]byte) string {
		path := filepath.Join(t.TempDir(), "set.zip")
		if err := os.WriteFile(path, zipOf(t, files), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	set := map[string][]byte{
		"config.json":             []byte(`{"id": 2, "version": 8}`),
		"emoji_2.webp":            []byte("two"),
		"emoji_1.webp":            []byte("one"),
		"emoji_10.webp":           []byte("ten"),
		"__MACOSX/._emoji_1.webp": []byte("junk"),
		"../emoji_3.webp":         []byte("outside"),
		"emoji_03.webp":           []byte("not a number as written"),
		"readme.txt":              []byte("x"),
	}
	for n := 3; n <= 9; n++ {
		set["emoji_"+string(rune('0'+n))+".webp"] = []byte{byte(n)}
	}
	dir := t.TempDir()
	images, version, err := TelegramSprites(write(set), dir)
	if err != nil {
		t.Fatal(err)
	}
	if version != 8 || len(images) != 10 {
		t.Fatalf("version %d, %d sprites", version, len(images))
	}
	// In the order of their numbers, not of their names.
	for i, want := range []string{"emoji_1.webp", "emoji_2.webp", "emoji_3.webp"} {
		if filepath.Base(images[i]) != want {
			t.Errorf("sprite %d is %s, want %s", i, filepath.Base(images[i]), want)
		}
	}
	if filepath.Base(images[9]) != "emoji_10.webp" {
		t.Errorf("the last sprite is %s", filepath.Base(images[9]))
	}
	if data, _ := os.ReadFile(images[0]); string(data) != "one" {
		t.Errorf("the first sprite holds %q", data)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 10 {
		t.Errorf("%d files were taken out, want the 10 sprites", len(entries))
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "emoji_3.webp")); err == nil {
		t.Error("a file was written outside the directory")
	}

	delete(set, "emoji_5.webp")
	if _, _, err := TelegramSprites(write(set), t.TempDir()); err == nil {
		t.Error("a set with a sprite missing was taken")
	}
	if _, _, err := TelegramSprites(write(map[string][]byte{"config.json": []byte("{}")}), t.TempDir()); err == nil {
		t.Error("a zip without sprites was taken")
	}
}
