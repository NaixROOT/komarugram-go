// SPDX-License-Identifier: Unlicense OR MIT

package chatmedia

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"komarugram/pkg/h264"
)

// AVCDecEnv names the variable that replaces the H.264 decoder module with
// one of the user's: a path to an avcdec.wasm, or an http(s) URL of one. The
// module is LGPL, and this is how a build of the user's takes its place.
const AVCDecEnv = "KOMARUGRAM_AVCDEC"

// The H.264 decoder is not in the binary. It is fetched, the first time a
// GIF plays in the sandbox, from the commit of libavcodec-wasm this version
// was tested with, and kept in the cache directory.
const (
	avcdecCommit = "51a13afa567d51a5f3f6d469de6a2761a1cd2aec"
	avcdecSHA256 = "f4a66e79f3aa20b12b264bb86413e1ad8be82f5455a588f6bf79d6a862e8088e"
	avcdecURL    = "https://raw.githubusercontent.com/komarugif/libavcodec-wasm/" + avcdecCommit + "/avcdec.wasm"
	// avcdecMax bounds a download; the module is about 2 MiB.
	avcdecMax = 64 << 20
)

// avcdecFetch serializes fetches, which the managers of every window may
// start at once. fetched holds the URLs of the user's modules already
// downloaded by this process, which are not verified and so are fetched
// again on the next start.
var (
	avcdecFetch sync.Mutex
	fetched     = map[string]string{}
)

// newAVCRuntime compiles the H.264 decoder, fetching it first if needed.
func newAVCRuntime(ctx context.Context) (*h264.Runtime, error) {
	module, err := loadAVCDec(ctx, os.Getenv(AVCDecEnv))
	if err != nil {
		return nil, err
	}
	return h264.NewRuntime(ctx, module)
}

// loadAVCDec returns the module named by source, AVCDecEnv's value, or the
// published one when it is empty.
func loadAVCDec(ctx context.Context, source string) ([]byte, error) {
	custom := source != ""
	if !custom {
		source = avcdecURL
	}
	if custom && !strings.HasPrefix(source, "https://") && !strings.HasPrefix(source, "http://") {
		module, err := os.ReadFile(source)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", AVCDecEnv, err)
		}
		return module, nil
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(cache, "komarugram-go")
	avcdecFetch.Lock()
	defer avcdecFetch.Unlock()
	if custom {
		if path, ok := fetched[source]; ok {
			return os.ReadFile(path)
		}
		sum := sha256.Sum256([]byte(source))
		path := filepath.Join(dir, "avcdec-custom-"+hex.EncodeToString(sum[:8])+".wasm")
		module, err := downloadAVCDec(ctx, source, "", path)
		if err == nil {
			fetched[source] = path
		}
		return module, err
	}
	path := filepath.Join(dir, "avcdec-"+avcdecSHA256[:16]+".wasm")
	if module, err := os.ReadFile(path); err == nil && sha256Hex(module) == avcdecSHA256 {
		return module, nil
	}
	return downloadAVCDec(ctx, source, avcdecSHA256, path)
}

// downloadAVCDec fetches url, checks it against want unless want is empty,
// and keeps it at path.
func downloadAVCDec(ctx context.Context, url, want, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download H.264 decoder: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download H.264 decoder: %s", resp.Status)
	}
	module, err := io.ReadAll(io.LimitReader(resp.Body, avcdecMax+1))
	if err != nil {
		return nil, fmt.Errorf("download H.264 decoder: %w", err)
	}
	if len(module) > avcdecMax {
		return nil, errors.New("download H.264 decoder: too large")
	}
	if want != "" && sha256Hex(module) != want {
		return nil, errors.New("download H.264 decoder: checksum mismatch")
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
