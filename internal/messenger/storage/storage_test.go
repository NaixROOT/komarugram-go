package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"komarugram/internal/messenger/model"
)

type fakeSource struct {
	usage model.CacheUsage
	err   error
}

func (f fakeSource) StorageUsage(context.Context) (model.CacheUsage, error) { return f.usage, f.err }

func write(t *testing.T, path string, n int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, n), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestSnapshotMeasuresOwnedRootsOnce(t *testing.T) {
	dir := t.TempDir()
	config, cache, outside := filepath.Join(dir, "config"), filepath.Join(dir, "cache"), filepath.Join(dir, "outside")
	write(t, filepath.Join(config, "settings.json"), 100)
	write(t, filepath.Join(config, "accounts.db"), 200)
	write(t, filepath.Join(config, "accounts", "a", "history.db.plain"), 3000)
	write(t, filepath.Join(config, "accounts", "a", "session.json"), 10)
	write(t, filepath.Join(config, "accounts", "b", "history.db.secure"), 4000)
	write(t, filepath.Join(config, "wallpapers", "w"), 500)
	write(t, filepath.Join(cache, "avcdec-0123.wasm"), 7000)
	write(t, filepath.Join(cache, "crashes", "fatal.txt"), 20)
	write(t, filepath.Join(cache, "unrelated.bin"), 9000)
	write(t, filepath.Join(outside, "big"), 1<<20)
	if runtime.GOOS != "windows" {
		if err := os.Symlink(outside, filepath.Join(config, "wallpapers", "dir")); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(outside, "big"), filepath.Join(config, "wallpapers", "file")); err != nil {
			t.Fatal(err)
		}
	}

	usage := model.CacheUsage{Media: 1234}
	s := New(config, cache, func() []Account {
		return []Account{
			{ID: "a", Dir: filepath.Join(config, "accounts", "a"), Source: fakeSource{usage: usage}},
			{ID: "b", Dir: filepath.Join(config, "accounts", "b")},
			{ID: "c", Dir: filepath.Join(config, "accounts", "c"), Source: fakeSource{err: errors.New("locked")}},
		}
	})
	snap, err := s.StorageSnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	roots := map[string]int64{}
	var total int64
	for _, r := range snap.Roots {
		roots[r.Name] = r.Usage.Bytes
		total += r.Usage.Bytes
	}
	want := map[string]int64{"accounts": 7010, "settings": 300, "wallpapers": 500, "modules": 7000, "crashes": 20}
	for name, n := range want {
		if roots[name] != n {
			t.Errorf("%s: %d, want %d", name, roots[name], n)
		}
	}
	if roots["emoji"] != 0 || roots["miniapp"] != 0 || total != 14830 {
		t.Error("roots", roots, total)
	}
	var used int64
	for _, v := range snap.Volumes {
		used += v.Used
		if runtime.GOOS == "linux" && (!v.Known || v.Total <= 0 || v.Free > v.Total) {
			t.Error("volume", v)
		}
	}
	if used != total || runtime.GOOS == "linux" && len(snap.Volumes) != 1 {
		t.Error("volumes", snap.Volumes)
	}
	if len(snap.Accounts) != 3 {
		t.Fatal(snap.Accounts)
	}
	a, b, c := snap.Accounts[0], snap.Accounts[1], snap.Accounts[2]
	if a.Files.Bytes != 3010 || a.Files.Files != 2 || !a.CacheKnown || a.Cache.Media != 1234 {
		t.Error("a", a)
	}
	if b.Files.Bytes != 4000 || b.CacheKnown || b.Err != nil {
		t.Error("b", b)
	}
	if c.Files.Bytes != 0 || c.CacheKnown || c.Err == nil {
		t.Error("c", c)
	}
	if next, _ := s.StorageSnapshot(context.Background()); next.Generation <= snap.Generation {
		t.Error("generation", next.Generation)
	}
}
