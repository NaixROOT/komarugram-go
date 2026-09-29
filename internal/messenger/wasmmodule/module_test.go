// SPDX-License-Identifier: Unlicense OR MIT

package wasmmodule

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

var testModule = Module{Name: "test", Env: "KOMARUGRAM_TEST_MODULE"}

func TestModuleFromPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "own.wasm")
	if err := os.WriteFile(path, []byte("own build"), 0o600); err != nil {
		t.Fatal(err)
	}
	module, err := testModule.load(context.Background(), path)
	if err != nil || string(module) != "own build" {
		t.Fatalf("got %q, %v", module, err)
	}
	if _, err := testModule.load(context.Background(), path+".missing"); err == nil {
		t.Fatal("a missing module loaded")
	}
}

func TestModuleDownload(t *testing.T) {
	body := []byte("module")
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Write(body)
	}))
	defer srv.Close()
	path := filepath.Join(t.TempDir(), "avcdec.wasm")
	if _, err := testModule.download(context.Background(), srv.URL, sha256Hex([]byte("other")), path); err == nil {
		t.Fatal("a module with another checksum was accepted")
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("a module with another checksum was kept")
	}
	module, err := testModule.download(context.Background(), srv.URL, sha256Hex(body), path)
	if err != nil || !bytes.Equal(module, body) {
		t.Fatalf("got %q, %v", module, err)
	}
	if kept, _ := os.ReadFile(path); !bytes.Equal(kept, body) {
		t.Fatalf("kept %q", kept)
	}

	// A module of the user's at a URL is fetched once per process.
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	requests = 0
	for range 2 {
		if module, err := testModule.load(context.Background(), srv.URL); err != nil || !bytes.Equal(module, body) {
			t.Fatalf("got %q, %v", module, err)
		}
	}
	if requests != 1 {
		t.Fatalf("%d requests", requests)
	}
}
