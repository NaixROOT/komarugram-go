// SPDX-License-Identifier: Unlicense OR MIT

package emojipacks

import (
	"bytes"
	"context"
	_ "embed" // the catalog built in
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Source is where a catalog is read from.
type Source interface {
	// Open opens a file of the catalog by its path in it, with '/'.
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	// String names the source for the user and the log.
	String() string
}

// CatalogEnv names the variable that points at a catalog, a directory or an
// http(s) URL, in place of the one built in.
const CatalogEnv = "KOMARUGRAM_EMOJI_PACKS"

// SourceFromEnv returns the source of the catalog the environment names,
// and the one built in without it.
func SourceFromEnv() Source {
	if location := os.Getenv(CatalogEnv); location != "" {
		return NewSource(location)
	}
	return Official()
}

// official.json is made by cmd/emoji-pack's "official".
//
//go:embed official.json
var officialIndex []byte

// Official returns the catalog built into the client: the emoji sets of
// Telegram Desktop. It has no files of its own; every file of its packs is
// at an address or in an archive of Telegram's cloud.
func Official() Source { return officialSource{} }

type officialSource struct{}

func (officialSource) Open(_ context.Context, path string) (io.ReadCloser, error) {
	if path != "index.json" {
		return nil, fmt.Errorf("%s is not in the catalog built in", path)
	}
	return io.NopCloser(bytes.NewReader(officialIndex)), nil
}

func (officialSource) String() string { return "built in" }

// downloads is the client of what is downloaded over HTTP.
var downloads = &http.Client{Timeout: 10 * time.Minute}

// openURL opens what is at an address.
func openURL(ctx context.Context, address string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	resp, err := downloads.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%s: %s", address, resp.Status)
	}
	return resp.Body, nil
}

// NewSource returns the source of the catalog at location: an http(s) URL
// or a directory. It is nil for an empty location.
func NewSource(location string) Source {
	switch {
	case location == "":
		return nil
	case strings.HasPrefix(location, "http://"), strings.HasPrefix(location, "https://"):
		return urlSource(strings.TrimRight(location, "/"))
	}
	return dirSource(location)
}

type dirSource string

func (d dirSource) Open(_ context.Context, path string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(string(d), filepath.FromSlash(path)))
}

func (d dirSource) String() string { return string(d) }

type urlSource string

func (u urlSource) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	var escaped []string
	for _, part := range strings.Split(path, "/") {
		escaped = append(escaped, url.PathEscape(part))
	}
	return openURL(ctx, string(u)+"/"+strings.Join(escaped, "/"))
}

func (u urlSource) String() string { return string(u) }

// maxIndexSize bounds an index.json.
const maxIndexSize = 4 << 20

// ReadIndex reads the catalog's index and returns its packs that this
// client can install; the others, of kinds it does not know or described
// wrongly, are passed over.
func ReadIndex(ctx context.Context, src Source) ([]Pack, error) {
	if src == nil {
		return nil, ErrNoCatalog
	}
	r, err := src.Open(ctx, "index.json")
	if err != nil {
		return nil, err
	}
	defer r.Close()
	data, err := io.ReadAll(io.LimitReader(r, maxIndexSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxIndexSize {
		return nil, fmt.Errorf("the catalog's index is over %d MB", maxIndexSize>>20)
	}
	var index Index
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("the catalog's index: %w", err)
	}
	if index.Version != IndexVersion {
		return nil, fmt.Errorf("the catalog is of version %d, this client reads %d", index.Version, IndexVersion)
	}
	var packs []Pack
	seen := map[string]bool{}
	for _, p := range index.Packs {
		if p.Validate() == nil && !seen[p.ID] {
			seen[p.ID] = true
			packs = append(packs, p)
		}
	}
	return packs, nil
}

// TelegramFetcher downloads the file of a post of a public channel through
// the user's account. progress, if not nil, is told the bytes received and
// the file's size.
type TelegramFetcher func(ctx context.Context, channel string, post int, progress func(done, total int64)) ([]byte, error)

// WithTelegram returns src with a way to the archives of Telegram's cloud
// beside it, for the packs that have one. A nil fetch is none.
func WithTelegram(src Source, fetch TelegramFetcher) Source {
	if src == nil || fetch == nil {
		return src
	}
	return telegramSource{Source: src, fetch: fetch}
}

type telegramSource struct {
	Source
	fetch TelegramFetcher
}

// HasTelegram reports whether src has a way to the archives of Telegram's
// cloud.
func HasTelegram(src Source) bool { return fetcherOf(src) != nil }

// fetcherOf returns the way to Telegram's archives that src has, or nil.
func fetcherOf(src Source) TelegramFetcher {
	if t, ok := src.(telegramSource); ok {
		return t.fetch
	}
	return nil
}
