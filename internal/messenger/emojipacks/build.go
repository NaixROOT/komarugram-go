// SPDX-License-Identifier: Unlicense OR MIT

package emojipacks

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// The making of a catalog, for cmd/emoji-pack: a pack's files are copied
// into the catalog's directory, measured and hashed, and the pack is put
// into the index in place of the one of its name.

// addFile copies the file at from into the pack's directory of the catalog
// under name, and describes it.
func addFile(catalog, id, name, from string) (File, error) {
	if !validFileName(name) {
		return File{}, fmt.Errorf("file name %q: only letters, digits, '-', '_' and '.'", name)
	}
	in, err := os.Open(from)
	if err != nil {
		return File{}, err
	}
	defer in.Close()
	dir := filepath.Join(catalog, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return File{}, err
	}
	out, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return File{}, err
	}
	sum := sha256.New()
	size, err := io.Copy(io.MultiWriter(out, sum), in)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return File{}, err
	}
	return File{Name: name, Size: size, SHA256: hex.EncodeToString(sum.Sum(nil))}, nil
}

// BuildFont puts an emoji font into the catalog as pack p, whose ID, Name,
// License and Source are given.
func BuildFont(catalog string, p Pack, font string) (Pack, error) {
	if !validID(p.ID) {
		return Pack{}, fmt.Errorf("pack id %q: lower-case letters, digits, '-' and '_'", p.ID)
	}
	if err := os.RemoveAll(filepath.Join(catalog, p.ID)); err != nil {
		return Pack{}, err
	}
	f, err := addFile(catalog, p.ID, plainName(filepath.Base(font)), font)
	if err != nil {
		return Pack{}, err
	}
	p.Kind, p.Files, p.Font, p.Sprites = KindFont, []File{f}, f.Name, nil
	return p, putPack(catalog, p)
}

// BuildSprites puts a set of sprites into the catalog as pack p: the
// pictures at images, in order, of the layout given, and the emoji of
// their cells, each with the other sequences drawn the same.
func BuildSprites(catalog string, p Pack, layout Sprites, images []string, order [][]string) (Pack, error) {
	if !validID(p.ID) {
		return Pack{}, fmt.Errorf("pack id %q: lower-case letters, digits, '-' and '_'", p.ID)
	}
	if len(images) == 0 || len(order) == 0 {
		return Pack{}, errors.New("no pictures or no emoji")
	}
	if err := os.RemoveAll(filepath.Join(catalog, p.ID)); err != nil {
		return Pack{}, err
	}
	p.Kind, p.Files, p.Font = KindSprites, nil, ""
	layout.Images, layout.Order = nil, "order.txt"
	for _, image := range images {
		f, err := addFile(catalog, p.ID, filepath.Base(image), image)
		if err != nil {
			return Pack{}, err
		}
		p.Files = append(p.Files, f)
		layout.Images = append(layout.Images, f.Name)
	}
	var list strings.Builder
	for _, sequences := range order {
		list.WriteString(strings.Join(sequences, " "))
		list.WriteByte('\n')
	}
	tmp := filepath.Join(catalog, p.ID, ".order")
	if err := os.WriteFile(tmp, []byte(list.String()), 0o644); err != nil {
		return Pack{}, err
	}
	f, err := addFile(catalog, p.ID, layout.Order, tmp)
	os.Remove(tmp)
	if err != nil {
		return Pack{}, err
	}
	p.Files = append(p.Files, f)
	p.Sprites = &layout
	// The pack is opened as the client will open it: the pictures are of
	// the layout, and hold the emoji listed.
	if err := p.Validate(); err != nil {
		return Pack{}, err
	}
	if _, err := OpenSprites(filepath.Join(catalog, p.ID), p); err != nil {
		return Pack{}, err
	}
	return p, putPack(catalog, p)
}

// plainName makes a file's name one a pack may have: what is not a letter,
// a digit, '-', '_' or '.' becomes '-'.
func plainName(name string) string {
	out := []rune(name)
	for i, c := range out {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '-' && c != '_' && c != '.' {
			out[i] = '-'
		}
	}
	name = strings.TrimLeft(string(out), ".")
	for strings.Contains(name, "..") {
		name = strings.ReplaceAll(name, "..", ".")
	}
	return name
}

// putPack writes p into the catalog's index, in place of the pack of its
// name.
func putPack(catalog string, p Pack) error {
	if err := p.Validate(); err != nil {
		return err
	}
	index := Index{Version: IndexVersion}
	path := filepath.Join(catalog, "index.json")
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &index); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if index.Version != IndexVersion {
			return fmt.Errorf("%s is of version %d, not %d", path, index.Version, IndexVersion)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	packs := index.Packs[:0]
	for _, old := range index.Packs {
		if old.ID != p.ID {
			packs = append(packs, old)
		}
	}
	index.Packs = append(packs, p)
	sort.Slice(index.Packs, func(i, j int) bool { return index.Packs[i].Name < index.Packs[j].Name })
	data, err := json.MarshalIndent(index, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// RemoveFromCatalog takes a pack out of the catalog: its files and its
// entry of the index.
func RemoveFromCatalog(catalog, id string) error {
	if !validID(id) {
		return fmt.Errorf("pack id %q", id)
	}
	path := filepath.Join(catalog, "index.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var index Index
	if err := json.Unmarshal(data, &index); err != nil {
		return err
	}
	packs := index.Packs[:0]
	for _, old := range index.Packs {
		if old.ID != id {
			packs = append(packs, old)
		}
	}
	index.Packs = packs
	if data, err = json.MarshalIndent(index, "", "\t"); err != nil {
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(catalog, id))
}
