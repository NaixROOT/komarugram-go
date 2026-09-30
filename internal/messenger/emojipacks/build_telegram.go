// SPDX-License-Identifier: Unlicense OR MIT

package emojipacks

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TelegramSprites takes the sprites of an emoji set of Telegram Desktop out
// of its zip into dir, and returns them in order with the version that its
// config.json gives: the order of the cells is of one version of the set.
func TelegramSprites(zipPath, dir string) (images []string, version int, err error) {
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, 0, err
	}
	defer archive.Close()
	byNumber := map[int]string{}
	for _, entry := range archive.File {
		var n int
		switch {
		case entry.Name == "config.json":
		case len(entry.Name) < 64 && strings.HasSuffix(entry.Name, ".webp"):
			if _, err := fmt.Sscanf(entry.Name, "emoji_%d.webp", &n); err != nil || n < 1 || fmt.Sprintf("emoji_%d.webp", n) != entry.Name {
				continue
			}
		default:
			continue
		}
		if entry.UncompressedSize64 > MaxArchiveSize {
			return nil, 0, fmt.Errorf("%s: %s is too large", zipPath, entry.Name)
		}
		r, err := entry.Open()
		if err != nil {
			return nil, 0, err
		}
		data, err := io.ReadAll(io.LimitReader(r, MaxArchiveSize))
		r.Close()
		if err != nil {
			return nil, 0, err
		}
		if n == 0 {
			var config struct {
				Version int `json:"version"`
			}
			if err := json.Unmarshal(data, &config); err != nil {
				return nil, 0, fmt.Errorf("%s: config.json: %w", zipPath, err)
			}
			version = config.Version
			continue
		}
		path := filepath.Join(dir, entry.Name)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return nil, 0, err
		}
		byNumber[n] = path
	}
	for n := 1; n <= len(byNumber); n++ {
		path, ok := byNumber[n]
		if !ok {
			return nil, 0, fmt.Errorf("%s: no emoji_%d.webp among its %d sprites", zipPath, n, len(byNumber))
		}
		images = append(images, path)
	}
	if len(images) == 0 {
		return nil, 0, fmt.Errorf("%s has no emoji_N.webp sprites", zipPath)
	}
	return images, version, nil
}

// BuildTelegramSprites is BuildSprites for an emoji set of Telegram
// Desktop, whose order is worked out and not given: a list of another
// version of the set than the sprites is caught by the emoji not ending in
// the last row of the sprites.
//
// With archive not nil the sprites stay in the cloud of Telegram: they are
// in the zip at zipPath, which is the file of the post that archive names.
// The catalog then has the order of the pack alone, and the client
// downloads the zip through the account of the user.
func BuildTelegramSprites(catalog string, p Pack, images []string, order [][]string, zipPath string, archive *TelegramArchive) (Pack, error) {
	built, err := BuildSprites(catalog, p, TelegramLayout, images, order)
	if err != nil {
		return Pack{}, err
	}
	set, err := OpenSprites(filepath.Join(catalog, built.ID), built)
	if err != nil {
		return Pack{}, err
	}
	cells := set.cellCount()
	if set.count <= cells-set.layout.Columns {
		RemoveFromCatalog(catalog, built.ID)
		return Pack{}, fmt.Errorf("%d emoji are listed for %d cells: the list is of another version of the set than the sprites", set.count, cells)
	}
	if archive == nil {
		return built, nil
	}
	data, err := os.ReadFile(zipPath)
	if err != nil {
		return Pack{}, err
	}
	sum := sha256.Sum256(data)
	t := *archive
	t.Size, t.SHA256 = int64(len(data)), hex.EncodeToString(sum[:])
	built.Telegram = &t
	for i, f := range built.Files {
		if f.Name == built.Sprites.Order {
			continue
		}
		built.Files[i].Archive = true
		if err := os.Remove(filepath.Join(catalog, built.ID, f.Name)); err != nil {
			return Pack{}, err
		}
	}
	return built, putPack(catalog, built)
}
