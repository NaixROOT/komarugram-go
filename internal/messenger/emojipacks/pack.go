// SPDX-License-Identifier: Unlicense OR MIT

// Package emojipacks keeps the emoji the user chose to draw with: packs
// from a catalog, downloaded once and kept beside the settings. A pack is
// an emoji font, or a set of sprites as Telegram Desktop's emoji sets are.
//
// A catalog is a directory, on disk or behind a URL:
//
//	index.json          the packs: Index
//	<pack>/<file>       the files of a pack, as its entry lists them
//
// cmd/emoji-pack makes one.
package emojipacks

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// IndexVersion is the version of the catalog's format this client reads.
const IndexVersion = 1

// Index is a catalog's index.json.
type Index struct {
	Version int    `json:"version"`
	Packs   []Pack `json:"packs"`
}

// Kind is what a pack draws emoji with.
type Kind string

const (
	// KindFont is an emoji font: one file.
	KindFont Kind = "font"
	// KindSprites is a set of sprites: pictures of a grid of emoji, and the
	// list of the emoji in the order of the cells.
	KindSprites Kind = "sprites"
)

// Pack is a pack of a catalog, or one installed.
type Pack struct {
	// ID names the pack's directory: lower-case letters, digits, '-', '_'.
	ID string `json:"id"`
	// Name is shown to the user.
	Name string `json:"name"`
	Kind Kind   `json:"kind"`
	// License and Source tell where the pack is from, for the user.
	License string `json:"license,omitempty"`
	Source  string `json:"source,omitempty"`
	// Files are all the files of the pack.
	Files []File `json:"files"`
	// Font is the file of a font pack.
	Font string `json:"font,omitempty"`
	// Sprites describes a sprite pack.
	Sprites *Sprites `json:"sprites,omitempty"`
	// Telegram is the archive that has the files of the pack marked
	// Archive; the catalog has only the others.
	Telegram *TelegramArchive `json:"telegram,omitempty"`
}

// TelegramArchive is a zip kept in Telegram's cloud: the file of a post of
// a public channel, as the emoji sets of Telegram Desktop are. It is
// downloaded through the user's account, which only reads the channel.
type TelegramArchive struct {
	// Channel is the channel's username, and Post the post's number.
	Channel string `json:"channel"`
	Post    int    `json:"post"`
	// Size and SHA256 are of the zip.
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// MaxArchiveSize is the largest archive downloaded from Telegram.
const MaxArchiveSize = 64 << 20

// File is a file of a pack: its name in the pack's directory, its size and
// the SHA-256 of its contents, in hex.
type File struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
	// Archive marks a file that is in the pack's Telegram archive, under
	// its name, and not in the catalog.
	Archive bool `json:"archive,omitempty"`
}

// Sprites is the layout of a sprite pack.
type Sprites struct {
	// Cell is the side of an emoji in the pictures, in pixels; Columns and
	// Rows are how many a full picture has.
	Cell    int `json:"cell"`
	Columns int `json:"columns"`
	Rows    int `json:"rows"`
	// Images are the pictures, in order; the last may have fewer rows.
	Images []string `json:"images"`
	// Order is the file that lists the emoji in the order of the cells,
	// one cell a line: the emoji, and after spaces other sequences drawn
	// the same. Sequences are without U+FE0F.
	Order string `json:"order"`
}

// Size is the size of all the pack's files.
func (p Pack) Size() int64 {
	var n int64
	for _, f := range p.Files {
		n += f.Size
	}
	return n
}

// DownloadSize is what installing the pack downloads: the files of the
// catalog, and the archive that has the others.
func (p Pack) DownloadSize() int64 {
	var n int64
	for _, f := range p.Files {
		if !f.Archive {
			n += f.Size
		}
	}
	if p.Telegram != nil {
		n += p.Telegram.Size
	}
	return n
}

// Revision tells one state of the pack's files from another: a catalog
// with another revision of an installed pack has an update.
func (p Pack) Revision() string {
	files := append([]File(nil), p.Files...)
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	sum := sha256.New()
	for _, f := range files {
		fmt.Fprintf(sum, "%s\x00%d\x00%s\x00", f.Name, f.Size, f.SHA256)
	}
	return hex.EncodeToString(sum.Sum(nil))[:16]
}

// MaxPackSize is the largest pack installed: a catalog is someone's
// repository, and what it says is checked.
const MaxPackSize = 256 << 20

// Validate checks that the pack can be installed: nothing in it names a
// place outside its directory, and its description is whole.
func (p Pack) Validate() error {
	if !validID(p.ID) {
		return fmt.Errorf("pack id %q", p.ID)
	}
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("pack %s has no name", p.ID)
	}
	if len(p.Files) == 0 {
		return fmt.Errorf("pack %s has no files", p.ID)
	}
	names := map[string]bool{}
	for _, f := range p.Files {
		if !validFileName(f.Name) || names[f.Name] {
			return fmt.Errorf("pack %s: file name %q", p.ID, f.Name)
		}
		names[f.Name] = true
		if f.Size <= 0 || len(f.SHA256) != sha256.Size*2 {
			return fmt.Errorf("pack %s: file %s has no size or hash", p.ID, f.Name)
		}
		if _, err := hex.DecodeString(f.SHA256); err != nil {
			return fmt.Errorf("pack %s: file %s: hash: %w", p.ID, f.Name, err)
		}
	}
	if p.Size() > MaxPackSize {
		return fmt.Errorf("pack %s is over %d MB", p.ID, MaxPackSize>>20)
	}
	archived := false
	for _, f := range p.Files {
		archived = archived || f.Archive
	}
	if t := p.Telegram; t != nil {
		if !validUsername(t.Channel) || t.Post <= 0 || t.Size <= 0 || t.Size > MaxArchiveSize || len(t.SHA256) != sha256.Size*2 {
			return fmt.Errorf("pack %s: its Telegram archive is not described", p.ID)
		}
		if _, err := hex.DecodeString(t.SHA256); err != nil {
			return fmt.Errorf("pack %s: its archive's hash: %w", p.ID, err)
		}
	} else if archived {
		return fmt.Errorf("pack %s has files of an archive, and no archive", p.ID)
	}
	switch p.Kind {
	case KindFont:
		if !names[p.Font] {
			return fmt.Errorf("pack %s: its font %q is not among its files", p.ID, p.Font)
		}
	case KindSprites:
		s := p.Sprites
		if s == nil || s.Cell <= 0 || s.Cell > 512 || s.Columns <= 0 || s.Columns > 256 || s.Rows <= 0 || s.Rows > 256 || len(s.Images) == 0 {
			return fmt.Errorf("pack %s: its sprites are not described", p.ID)
		}
		for _, name := range append([]string{s.Order}, s.Images...) {
			if !names[name] {
				return fmt.Errorf("pack %s: %q is not among its files", p.ID, name)
			}
		}
	default:
		return fmt.Errorf("pack %s is of an unknown kind %q", p.ID, p.Kind)
	}
	return nil
}

func validID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	for _, c := range id {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' && c != '_' {
			return false
		}
	}
	return true
}

// validUsername reports whether name can be a Telegram username.
func validUsername(name string) bool {
	if len(name) < 4 || len(name) > 32 {
		return false
	}
	for _, c := range name {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' {
			return false
		}
	}
	return true
}

// validFileName reports whether name is a plain file name: no directories,
// nothing a file system gives a meaning of its own.
func validFileName(name string) bool {
	if name == "" || len(name) > 128 || name[0] == '.' || name == manifestName {
		return false
	}
	for _, c := range name {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '-' && c != '_' && c != '.' {
			return false
		}
	}
	return !strings.Contains(name, "..")
}

// ErrNoCatalog is the error of asking for a catalog when none is set.
var ErrNoCatalog = errors.New("no emoji catalog is set")

// ErrNeedsTelegram is the error of installing a pack whose files are in
// Telegram's cloud without an account that is connected.
var ErrNeedsTelegram = errors.New("the pack is downloaded from Telegram, and no account is connected")
