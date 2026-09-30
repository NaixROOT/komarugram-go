package crash

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGuardReportsOncePerSite(t *testing.T) {
	// os.UserCacheDir reads LOCALAPPDATA on Windows, XDG_CACHE_HOME elsewhere.
	cache := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cache)
	t.Setenv("LOCALAPPDATA", cache)
	var nilMap map[string]int
	for range 3 {
		p := Guard("test site", func() { nilMap["x"] = 1 })
		if p == nil || !strings.Contains(p.Error(), "test site") || len(p.Stack) == 0 {
			t.Fatalf("panic not recovered: %v", p)
		}
	}
	if Guard("test site", func() {}) != nil {
		t.Fatal("reported a panic that did not happen")
	}
	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(dir, "panic-*.txt"))
	if len(files) != 1 {
		t.Fatalf("want one report for a repeating panic, got %d", len(files))
	}
	body, _ := os.ReadFile(files[0])
	if !strings.Contains(string(body), "assignment to entry in nil map") {
		t.Fatalf("report without the panic value:\n%s", body)
	}
}

func TestTakeFatalKeepsReportOfDeadRun(t *testing.T) {
	dir := t.TempDir()
	if got := takeFatal(dir); got != "" {
		t.Fatalf("report %q of a run that left none", got)
	}
	path := filepath.Join(dir, fatalName)
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if got := takeFatal(dir); got != "" {
		t.Fatalf("report %q of a run that ended well", got)
	}
	if err := os.WriteFile(path, []byte("fatal error: demonstration"), 0o600); err != nil {
		t.Fatal(err)
	}
	kept := takeFatal(dir)
	body, err := os.ReadFile(kept)
	if err != nil || string(body) != "fatal error: demonstration" {
		t.Fatalf("report not kept: %q, %v", body, err)
	}
	if matched, _ := filepath.Match("panic-*.txt", filepath.Base(kept)); !matched {
		t.Fatalf("report %q is not among those pruned", kept)
	}
	if got := takeFatal(dir); got != "" {
		t.Fatalf("report %q told twice", got)
	}
}

func TestReporterNotifiesWithSavedReport(t *testing.T) {
	// os.UserCacheDir reads LOCALAPPDATA on Windows, XDG_CACHE_HOME elsewhere.
	cache := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cache)
	t.Setenv("LOCALAPPDATA", cache)
	var got *Panic
	stop := Subscribe(func(p *Panic) { got = p })
	defer stop()
	p := Guard("notification test", func() { panic("demonstration") })
	if got != p || p.Path == "" {
		t.Fatalf("notifier did not receive the saved panic: %+v", got)
	}
	data, err := os.ReadFile(p.Path)
	if err != nil || string(data) != p.Text() {
		t.Fatalf("saved report differs from copy text: %v", err)
	}
}
