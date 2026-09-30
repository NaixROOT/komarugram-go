// SPDX-License-Identifier: Unlicense OR MIT

package emojipacks

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// manifestName is the file of an installed pack that describes it: the
// pack's entry of the catalog it came from.
const manifestName = "pack.json"

// ErrBusy is the error of installing or removing a pack that is being
// installed.
var ErrBusy = errors.New("the pack is being installed")

// Store keeps the installed packs in a directory, one directory a pack. It
// is used by all windows at once.
type Store struct {
	dir string

	mu sync.Mutex
	// busy are the packs being installed or removed.
	busy map[string]bool
	// sprites are the sprite sets opened, by pack.
	sprites map[string]*SpriteSet
}

var (
	storesMu sync.Mutex
	stores   = map[string]*Store{}
)

// Open returns the store of a directory, the same one for the same
// directory: the windows of a process share it. An empty directory is a
// store that keeps nothing.
func Open(dir string) *Store {
	if dir == "" {
		return &Store{busy: map[string]bool{}, sprites: map[string]*SpriteSet{}}
	}
	storesMu.Lock()
	defer storesMu.Unlock()
	if s := stores[dir]; s != nil {
		return s
	}
	s := &Store{dir: dir, busy: map[string]bool{}, sprites: map[string]*SpriteSet{}}
	stores[dir] = s
	return s
}

// Installed returns the packs installed, by name.
func (s *Store) Installed() []Pack {
	if s == nil || s.dir == "" {
		return nil
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil
	}
	var packs []Pack
	for _, e := range entries {
		if p, ok := s.Pack(e.Name()); ok {
			packs = append(packs, p)
		}
	}
	sort.Slice(packs, func(i, j int) bool { return packs[i].Name < packs[j].Name })
	return packs
}

// Pack returns an installed pack.
func (s *Store) Pack(id string) (Pack, bool) {
	if s == nil || s.dir == "" || !validID(id) {
		return Pack{}, false
	}
	data, err := os.ReadFile(filepath.Join(s.dir, id, manifestName))
	if err != nil {
		return Pack{}, false
	}
	var p Pack
	if json.Unmarshal(data, &p) != nil || p.ID != id || p.Validate() != nil {
		return Pack{}, false
	}
	return p, true
}

// Path returns where a file of an installed pack is.
func (s *Store) Path(id, name string) string {
	return filepath.Join(s.dir, id, name)
}

func (s *Store) begin(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.busy[id] {
		return ErrBusy
	}
	s.busy[id] = true
	return nil
}

func (s *Store) end(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.busy, id)
	// What was opened is of the files as they were.
	delete(s.sprites, id)
}

// Install downloads a pack of the catalog at src and keeps it, in place of
// the one installed under its name. Every file is checked against the size
// and the hash the catalog gives, and the pack replaces the old one only
// once all of them are there. progress, if not nil, is told the bytes
// read so far and the pack's size.
func (s *Store) Install(ctx context.Context, src Source, p Pack, progress func(done, total int64)) (err error) {
	if s == nil || s.dir == "" {
		return errors.New("no place to keep emoji packs")
	}
	if src == nil {
		return ErrNoCatalog
	}
	if err := p.Validate(); err != nil {
		return err
	}
	if err := s.begin(p.ID); err != nil {
		return err
	}
	defer s.end(p.ID)
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(s.dir, ".tmp-"+p.ID+"-")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(tmp)
		}
	}()
	total, done := p.DownloadSize(), int64(0)
	tell := func(n int64) {
		done += n
		if progress != nil {
			progress(done, total)
		}
	}
	if p.Telegram != nil {
		if err := unpackArchive(ctx, fetcherOf(src), p, tmp, tell); err != nil {
			return err
		}
	}
	for _, f := range p.Files {
		if f.Archive {
			continue
		}
		open := func() (io.ReadCloser, error) { return src.Open(ctx, p.ID+"/"+f.Name) }
		if f.URL != "" {
			open = func() (io.ReadCloser, error) { return openURL(ctx, f.URL) }
		}
		if err := fetch(ctx, open, filepath.Join(tmp, f.Name), f, tell); err != nil {
			return fmt.Errorf("%s: %w", f.Name, err)
		}
	}
	if p.Kind == KindSprites {
		// The cells are cut out here, off the frame: a second or two for a
		// set of Telegram Desktop's. A pack whose sprites are not what it
		// says they are is turned down.
		if err := buildCells(tmp, p); err != nil {
			return err
		}
	}
	manifest, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmp, manifestName), manifest, 0o600); err != nil {
		return err
	}
	final := filepath.Join(s.dir, p.ID)
	if err := os.RemoveAll(final); err != nil {
		return err
	}
	return os.Rename(tmp, final)
}

// unpackArchive downloads the pack's archive from Telegram and takes the
// pack's files that are in it out into dir. The archive and each file are
// checked against what the catalog says of them; what else the archive has
// is left in it.
func unpackArchive(ctx context.Context, download TelegramFetcher, p Pack, dir string, read func(n int64)) error {
	if download == nil {
		return ErrNeedsTelegram
	}
	t := p.Telegram
	var told int64
	data, err := download(ctx, t.Channel, t.Post, func(done, _ int64) {
		// Parts come in any order; what is told is how much came.
		if done > told && done <= t.Size {
			read(done - told)
			told = done
		}
	})
	if err != nil {
		return fmt.Errorf("@%s/%d: %w", t.Channel, t.Post, err)
	}
	if told < t.Size {
		read(t.Size - told)
	}
	sum := sha256.Sum256(data)
	if int64(len(data)) != t.Size || hex.EncodeToString(sum[:]) != t.SHA256 {
		return fmt.Errorf("@%s/%d is not the archive the catalog describes", t.Channel, t.Post)
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("@%s/%d: %w", t.Channel, t.Post, err)
	}
	entries := map[string]*zip.File{}
	for _, entry := range archive.File {
		entries[entry.Name] = entry
	}
	for _, f := range p.Files {
		if !f.Archive {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		entry := entries[f.Name]
		if entry == nil || entry.UncompressedSize64 != uint64(f.Size) {
			return fmt.Errorf("%s: not in the archive as the catalog describes it", f.Name)
		}
		r, err := entry.Open()
		if err != nil {
			return fmt.Errorf("%s: %w", f.Name, err)
		}
		// One more byte would tell a file that is larger than it says.
		content, err := io.ReadAll(io.LimitReader(r, f.Size+1))
		r.Close()
		if err != nil {
			return fmt.Errorf("%s: %w", f.Name, err)
		}
		sum := sha256.Sum256(content)
		if int64(len(content)) != f.Size || hex.EncodeToString(sum[:]) != f.SHA256 {
			return fmt.Errorf("%s: not the file the catalog describes", f.Name)
		}
		if err := os.WriteFile(filepath.Join(dir, f.Name), content, 0o600); err != nil {
			return err
		}
	}
	return nil
}

// fetch copies a file of a pack, as open gives it, to path, and fails unless
// it is of the size and the hash of f.
func fetch(ctx context.Context, open func() (io.ReadCloser, error), path string, f File, read func(n int64)) error {
	r, err := open()
	if err != nil {
		return err
	}
	defer r.Close()
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	sum := sha256.New()
	buf := make([]byte, 64<<10)
	var size int64
	for {
		if err := ctx.Err(); err != nil {
			out.Close()
			return err
		}
		n, err := r.Read(buf)
		if n > 0 {
			size += int64(n)
			if size > f.Size {
				out.Close()
				return errors.New("the file is larger than the catalog says")
			}
			sum.Write(buf[:n])
			if _, werr := out.Write(buf[:n]); werr != nil {
				out.Close()
				return werr
			}
			read(int64(n))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			out.Close()
			return err
		}
	}
	if err := out.Close(); err != nil {
		return err
	}
	if size != f.Size {
		return errors.New("the file is smaller than the catalog says")
	}
	if hex.EncodeToString(sum.Sum(nil)) != f.SHA256 {
		return errors.New("the file is not the one the catalog describes")
	}
	return nil
}

// Remove deletes an installed pack.
func (s *Store) Remove(id string) error {
	if s == nil || s.dir == "" || !validID(id) {
		return nil
	}
	if err := s.begin(id); err != nil {
		return err
	}
	defer s.end(id)
	return os.RemoveAll(filepath.Join(s.dir, id))
}

// Sprites opens the sprite set of an installed pack, once for all who ask.
func (s *Store) Sprites(id string) (*SpriteSet, error) {
	p, ok := s.Pack(id)
	if !ok {
		return nil, fmt.Errorf("emoji pack %s is not installed", id)
	}
	if p.Kind != KindSprites {
		return nil, fmt.Errorf("emoji pack %s has no sprites", id)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if set := s.sprites[id]; set != nil && set.revision == p.Revision() {
		return set, nil
	}
	set, err := OpenSprites(filepath.Join(s.dir, id), p)
	if err != nil {
		return nil, err
	}
	s.sprites[id] = set
	return set, nil
}

// ReadDir describes the pack in a directory: an installed one, by its
// manifest, or one of a catalog, by the catalog's index beside it.
func ReadDir(dir string) (Pack, error) {
	var p Pack
	if data, err := os.ReadFile(filepath.Join(dir, manifestName)); err == nil {
		if err := json.Unmarshal(data, &p); err != nil {
			return Pack{}, err
		}
		return p, p.Validate()
	}
	packs, err := ReadIndex(context.Background(), NewSource(filepath.Dir(filepath.Clean(dir))))
	if err != nil {
		return Pack{}, err
	}
	for _, p := range packs {
		if p.ID == filepath.Base(filepath.Clean(dir)) {
			return p, nil
		}
	}
	return Pack{}, fmt.Errorf("%s is not a pack of the catalog beside it", dir)
}

// Dir is the directory of the store, and PackDir that of an installed
// pack in it.
func (s *Store) Dir() string { return s.dir }

func (s *Store) PackDir(id string) string { return filepath.Join(s.dir, id) }
