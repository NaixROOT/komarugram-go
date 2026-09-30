// SPDX-License-Identifier: Unlicense OR MIT

package emojipacks

import (
	"context"
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
// http(s) URL, over DefaultCatalog.
const CatalogEnv = "KOMARUGRAM_EMOJI_PACKS"

// DefaultCatalog is the catalog of a client that was not told of another:
// empty until the project has one published.
const DefaultCatalog = ""

// SourceFromEnv returns the source of the catalog the environment or the
// default names, nil for none.
func SourceFromEnv() Source {
	location := os.Getenv(CatalogEnv)
	if location == "" {
		location = DefaultCatalog
	}
	return NewSource(location)
}

// NewSource returns the source of the catalog at location: an http(s) URL
// or a directory. It is nil for an empty location.
func NewSource(location string) Source {
	switch {
	case location == "":
		return nil
	case strings.HasPrefix(location, "http://"), strings.HasPrefix(location, "https://"):
		return &urlSource{base: strings.TrimRight(location, "/"), client: &http.Client{Timeout: 10 * time.Minute}}
	}
	return dirSource(location)
}

type dirSource string

func (d dirSource) Open(_ context.Context, path string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(string(d), filepath.FromSlash(path)))
}

func (d dirSource) String() string { return string(d) }

type urlSource struct {
	base   string
	client *http.Client
}

func (u *urlSource) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	var escaped []string
	for _, part := range strings.Split(path, "/") {
		escaped = append(escaped, url.PathEscape(part))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.base+"/"+strings.Join(escaped, "/"), nil)
	if err != nil {
		return nil, err
	}
	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%s: %s", path, resp.Status)
	}
	return resp.Body, nil
}

func (u *urlSource) String() string { return u.base }

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

// fetcherOf returns the way to Telegram's archives that src has, or nil.
func fetcherOf(src Source) TelegramFetcher {
	if t, ok := src.(telegramSource); ok {
		return t.fetch
	}
	return nil
}
