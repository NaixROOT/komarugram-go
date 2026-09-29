// SPDX-License-Identifier: Unlicense OR MIT

// Package wasmmodule fetches the WebAssembly decoders the client does not
// carry in its binary: those whose license asks for their sources beside
// them, or leaves them for a distribution to leave out — FFmpeg's H.264
// decoder (LGPL), fdk-aac (patents). Each lives in a repository of its own
// with its sources, and is fetched the first time it is needed, from the
// commit this version was tested with, checked against its SHA-256 and
// kept in the cache directory. A variable of the environment names a build
// of the user's instead: a path, or an http(s) URL.
package wasmmodule

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Module is one fetched module.
type Module struct {
	// Name names its cache files, as avcdec.
	Name string
	// Env is the variable that names a build of the user's.
	Env string
	// URL is where the tested build is, at a pinned commit; SHA256 is its
	// checksum.
	URL, SHA256 string
}

// maxModule bounds a download.
const maxModule = 64 << 20

// fetch serializes fetches, which the windows may start at once. fetched
// holds the URLs of the user's modules downloaded by this process, which
// are not checked and so are fetched again on the next start.
var (
	fetch   sync.Mutex
	fetched = map[string]string{}
)

// Load returns the module: the build the environment names, or the tested
// one, from the cache or downloaded.
func (m Module) Load(ctx context.Context) ([]byte, error) {
	return m.load(ctx, os.Getenv(m.Env))
}

// load returns the module named by source, m.Env's value, or the tested
// one when it is empty.
func (m Module) load(ctx context.Context, source string) ([]byte, error) {
	custom := source != ""
	if !custom {
		source = m.URL
	}
	if custom && !strings.HasPrefix(source, "https://") && !strings.HasPrefix(source, "http://") {
		module, err := os.ReadFile(source)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", m.Env, err)
		}
		return module, nil
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(cache, "komarugram-go")
	fetch.Lock()
	defer fetch.Unlock()
	if custom {
		if path, ok := fetched[source]; ok {
			return os.ReadFile(path)
		}
		sum := sha256.Sum256([]byte(source))
		path := filepath.Join(dir, m.Name+"-custom-"+hex.EncodeToString(sum[:8])+".wasm")
		module, err := m.download(ctx, source, "", path)
		if err == nil {
			fetched[source] = path
		}
		return module, err
	}
	path := filepath.Join(dir, m.Name+"-"+m.SHA256[:16]+".wasm")
	if module, err := os.ReadFile(path); err == nil && sha256Hex(module) == m.SHA256 {
		return module, nil
	}
	return m.download(ctx, source, m.SHA256, path)
}

// download fetches url, checks it against want unless want is empty, and
// keeps it at path.
func (m Module) download(ctx context.Context, url, want, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", m.Name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: %s", m.Name, resp.Status)
	}
	module, err := io.ReadAll(io.LimitReader(resp.Body, maxModule+1))
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", m.Name, err)
	}
	if len(module) > maxModule {
		return nil, fmt.Errorf("download %s: too large", m.Name)
	}
	if want != "" && sha256Hex(module) != want {
		return nil, fmt.Errorf("download %s: checksum mismatch", m.Name)
	}
	// A failure to keep it only means fetching it again next time.
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err == nil {
		tmp := path + ".tmp"
		if os.WriteFile(tmp, module, 0o600) == nil && os.Rename(tmp, path) != nil {
			os.Remove(tmp)
		}
	}
	return module, nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// AVCDec is FFmpeg's H.264 decoder, LGPL, from libavcodec-wasm, for GIFs and
// animated avatars without an FFmpeg. KOMARUGRAM_AVCDEC names a build of
// the user's instead, as the LGPL lets them.
var AVCDec = Module{
	Name:   "avcdec",
	Env:    "KOMARUGRAM_AVCDEC",
	URL:    "https://raw.githubusercontent.com/komarugif/libavcodec-wasm/51a13afa567d51a5f3f6d469de6a2761a1cd2aec/avcdec.wasm",
	SHA256: "f4a66e79f3aa20b12b264bb86413e1ad8be82f5455a588f6bf79d6a862e8088e",
}

// AACDec is the decoder of fdk-aac, from fdk-aac-wasm, for voice messages
// and music in M4A. Its license grants no patent rights on AAC: it is
// fetched, not carried, so that a distribution can leave it out, and
// KOMARUGRAM_AACDEC names another build, or another decoder with its
// exports.
var AACDec = Module{
	Name:   "aacdec",
	Env:    "KOMARUGRAM_AACDEC",
	URL:    "https://raw.githubusercontent.com/komarugif/fdk-aac-wasm/68c53cefce6ccf620e31ea96281623a63f175e7e/aacdec.wasm",
	SHA256: "26d138ef82075b7de3ab838002b0f7000082736f8ba698b2aa0366fdefb5c4d7",
}
