// SPDX-License-Identifier: Unlicense OR MIT

package emojipacks

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// The making of the catalog built into the client, for cmd/emoji-pack: the
// emoji sets of Telegram Desktop, described by where their files are and
// what they hold. None of the files is kept.

// OfficialFile is a file of a set: its name in the pack, what it holds,
// and the address it is downloaded from; a file without one is in the
// archive of the set.
type OfficialFile struct {
	Name string
	URL  string
	Data []byte
}

func (f OfficialFile) describe() File {
	sum := sha256.Sum256(f.Data)
	return File{Name: f.Name, Size: int64(len(f.Data)), SHA256: hex.EncodeToString(sum[:]), URL: f.URL, Archive: f.URL == ""}
}

// OfficialPack describes an emoji set of Telegram Desktop as pack p, whose
// ID, Name, License and Source are given: its sprites, in order, and the
// emoji list of the sources of its version. With an archive the sprites are
// taken out of it, and are those of its emoji_N.webp; archive then names
// the post it is the file of.
func OfficialPack(p Pack, list OfficialFile, images []OfficialFile, archive []byte, post *TelegramArchive) (Pack, error) {
	if post != nil {
		sprites, err := archiveSprites(archive)
		if err != nil {
			return Pack{}, err
		}
		images = sprites
		sum := sha256.Sum256(archive)
		t := *post
		t.Size, t.SHA256 = int64(len(archive)), hex.EncodeToString(sum[:])
		p.Telegram = &t
	}
	layout := TelegramLayout
	layout.List = list.Name
	p.Kind, p.Font, p.Files = KindSprites, "", []File{list.describe()}
	// The pack is opened as the client will open it: the sprites are of the
	// layout, and the list is of their version.
	dir, err := os.MkdirTemp("", "emoji-pack")
	if err != nil {
		return Pack{}, err
	}
	defer os.RemoveAll(dir)
	for _, f := range append([]OfficialFile{list}, images...) {
		if !validFileName(f.Name) {
			return Pack{}, fmt.Errorf("file name %q", f.Name)
		}
		if err := os.WriteFile(filepath.Join(dir, f.Name), f.Data, 0o600); err != nil {
			return Pack{}, err
		}
	}
	for _, f := range images {
		p.Files = append(p.Files, f.describe())
		layout.Images = append(layout.Images, f.Name)
	}
	p.Sprites = &layout
	if err := p.Validate(); err != nil {
		return Pack{}, err
	}
	if _, err := OpenSprites(dir, p); err != nil {
		return Pack{}, fmt.Errorf("%s: %w", p.ID, err)
	}
	return p, nil
}

// archiveSprites returns the emoji_N.webp of the zip of a set, in order.
func archiveSprites(archive []byte) ([]OfficialFile, error) {
	r, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, err
	}
	byNumber := map[int]OfficialFile{}
	for _, entry := range r.File {
		var n int
		if _, err := fmt.Sscanf(entry.Name, "emoji_%d.webp", &n); err != nil || n < 1 || fmt.Sprintf("emoji_%d.webp", n) != entry.Name {
			continue
		}
		if entry.UncompressedSize64 > MaxArchiveSize {
			return nil, fmt.Errorf("%s is too large", entry.Name)
		}
		f, err := entry.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(f, MaxArchiveSize))
		f.Close()
		if err != nil {
			return nil, err
		}
		byNumber[n] = OfficialFile{Name: entry.Name, Data: data}
	}
	var images []OfficialFile
	for n := 1; n <= len(byNumber); n++ {
		f, ok := byNumber[n]
		if !ok {
			return nil, fmt.Errorf("no emoji_%d.webp among the %d sprites of the archive", n, len(byNumber))
		}
		images = append(images, f)
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("the archive has no emoji_N.webp sprites")
	}
	return images, nil
}

// WriteIndex writes the index of a catalog of the packs given, in their
// order, to path.
func WriteIndex(path string, packs []Pack) error {
	data, err := json.MarshalIndent(Index{Version: IndexVersion, Packs: packs}, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
