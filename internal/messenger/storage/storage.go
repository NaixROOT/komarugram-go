// SPDX-License-Identifier: Unlicense OR MIT

// Package storage measures KomaruGram's local data: the directories it
// owns, the volumes they are on and every account's cache, for Data and
// Storage. It only reads; nothing here deletes.
package storage

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"komarugram/internal/messenger/model"
)

// Account is an account to report.
type Account struct {
	ID  string
	Dir string
	// Source is the account's open store; nil while it is locked or not
	// running, when only its files are measured.
	Source model.StorageUsageSource
}

// Service implements model.StorageService.
type Service struct {
	config, cache string
	accounts      func() []Account
	generation    atomic.Uint64
}

var _ model.StorageService = (*Service)(nil)

// New measures the data under config and cache, KomaruGram's directories
// in the user's configuration and cache directories, and the accounts
// accounts lists at the time of each snapshot.
func New(config, cache string, accounts func() []Account) *Service {
	return &Service{config: config, cache: cache, accounts: accounts}
}

// Dirs are KomaruGram's directories in the user's configuration and cache
// directories.
func Dirs() (config, cache string, err error) {
	c, err := os.UserConfigDir()
	if err != nil {
		return "", "", err
	}
	k, err := os.UserCacheDir()
	if err != nil {
		return "", "", err
	}
	return filepath.Join(c, "komarugram-go"), filepath.Join(k, "komarugram-go"), nil
}

type root struct {
	name, path string
	// top measures only the files right in path that it accepts.
	top func(name string) bool
}

// roots are the directories KomaruGram owns (docs/DATA_AND_STORAGE.md,
// "Storage owners"). Temporary files are not among them: they are a send's
// or a playback's, and found only by their owner.
func (s *Service) roots() []root {
	return []root{
		{name: "accounts", path: filepath.Join(s.config, "accounts")},
		{name: "settings", path: s.config, top: func(string) bool { return true }},
		{name: "wallpapers", path: filepath.Join(s.config, "wallpapers")},
		{name: "emoji", path: filepath.Join(s.config, "emoji")},
		{name: "miniapp", path: filepath.Join(s.config, "miniapp")},
		{name: "modules", path: s.cache, top: func(name string) bool { return strings.HasSuffix(name, ".wasm") }},
		{name: "crashes", path: filepath.Join(s.cache, "crashes")},
		{name: "locks", path: filepath.Join(s.cache, "locks")},
	}
}

// StorageSnapshot measures everything once.
func (s *Service) StorageSnapshot(ctx context.Context) (model.StorageSnapshot, error) {
	snap := model.StorageSnapshot{Generation: s.generation.Add(1)}
	volumes := map[string]int{}
	for _, r := range s.roots() {
		u, err := measure(ctx, r.path, r.top)
		if err != nil {
			return snap, err
		}
		v := volumeOf(r.path)
		i, ok := volumes[v.ID]
		if !ok {
			i = len(snap.Volumes)
			volumes[v.ID] = i
			snap.Volumes = append(snap.Volumes, v)
		}
		snap.Volumes[i].Used += u.Bytes
		snap.Roots = append(snap.Roots, model.StorageRoot{Name: r.name, Path: r.path, Volume: v.ID, Usage: u})
	}
	for _, a := range s.accounts() {
		as := model.AccountStorage{ID: a.ID}
		var err error
		if as.Files, err = measure(ctx, a.Dir, nil); err != nil {
			return snap, err
		}
		if a.Source != nil {
			as.Cache, as.Err = a.Source.StorageUsage(ctx)
			as.CacheKnown = as.Err == nil
		}
		snap.Accounts = append(snap.Accounts, as)
	}
	return snap, ctx.Err()
}

// volumeOf is the volume path is on, or would be: a directory not made yet
// is on its nearest parent's.
func volumeOf(path string) model.Volume {
	for p := path; ; {
		if _, err := os.Stat(p); err == nil {
			id, total, free, ok := volume(p)
			if !ok {
				return model.Volume{ID: p, Path: p}
			}
			return model.Volume{ID: id, Path: p, Total: total, Free: free, Known: true}
		}
		parent := filepath.Dir(p)
		if parent == p {
			return model.Volume{ID: path, Path: path}
		}
		p = parent
	}
}

// measure adds up the regular files under path, or the ones top accepts
// right in it. Links are not followed, and a missing path is empty.
func measure(ctx context.Context, path string, top func(string) bool) (model.DiskUsage, error) {
	var u model.DiskUsage
	add := func(info fs.FileInfo) {
		u.Files++
		u.Bytes += info.Size()
		u.Allocated += allocated(info)
	}
	if top != nil {
		entries, err := os.ReadDir(path)
		if errors.Is(err, fs.ErrNotExist) {
			return u, nil
		}
		for _, e := range entries {
			if e.Type().IsRegular() && top(e.Name()) {
				if info, err := e.Info(); err == nil {
					add(info)
				}
			}
		}
		return u, err
	}
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == path && errors.Is(err, fs.ErrNotExist) {
				return fs.SkipAll
			}
			// What cannot be read is left out, as a file that went away.
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.Type().IsRegular() {
			if info, err := d.Info(); err == nil {
				add(info)
			}
		}
		return nil
	})
	return u, err
}
