// SPDX-License-Identifier: Unlicense OR MIT

// Package securedb opens SQLite databases through the pure-Go encrypted VFS:
// account-scoped ones for chat history, update-state and layout caches, and
// shared ones such as the list of accounts.
package securedb

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/vfs/adiantum"

	"komarugram/internal/messenger/security"
)

// CopyPlain makes a consistent, checked plaintext copy of an encrypted
// database. The caller serializes writes and changes to the new file before
// releasing that serialization lock.
func CopyPlain(source *sql.DB, target string) error {
	if _, err := os.Stat(target); err == nil {
		return nil // A completed copy may already have received live writes.
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	tmp := target + ".tmp"
	if err := os.Remove(tmp); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(tmp), RawQuery: "vfs=os"}
	if _, err := source.Exec(`VACUUM main INTO ?`, u.String()); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("securedb: copy to plaintext: %w", err)
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	check, err := sql.Open("sqlite3", u.String())
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	var result string
	err = check.QueryRow(`PRAGMA integrity_check`).Scan(&result)
	if closeErr := check.Close(); err == nil {
		err = closeErr
	}
	if err != nil || result != "ok" {
		_ = os.Remove(tmp)
		return fmt.Errorf("securedb: plaintext copy failed integrity check: %v, %s", err, result)
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// CopyEncrypted makes a checked encrypted copy of a plaintext history DB.
func CopyEncrypted(source *sql.DB, target, accountID string, protection *security.Manager) error {
	if accountID == "" {
		return errors.New("securedb: empty account id")
	}
	if _, err := os.Stat(target); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	tmp := target + ".tmp"
	if err := os.Remove(tmp); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	key, err := protection.DeriveKey("sqlite/" + accountID)
	if err != nil {
		return err
	}
	defer clear(key)
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(tmp)}
	query := u.Query()
	query.Set("vfs", "adiantum")
	query.Set("hexkey", hex.EncodeToString(key))
	u.RawQuery = query.Encode()
	if _, err := source.Exec(`VACUUM main INTO ?`, u.String()); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("securedb: copy to encrypted: %w", err)
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	check, err := Open(protection, tmp, accountID)
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	var result string
	err = check.QueryRow(`PRAGMA integrity_check`).Scan(&result)
	if closeErr := check.Close(); err == nil {
		err = closeErr
	}
	if err != nil || result != "ok" {
		_ = os.Remove(tmp)
		return fmt.Errorf("securedb: encrypted copy failed integrity check: %v, %s", err, result)
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// Open opens path with a key derived specifically for accountID. The key is
// supplied as raw hex so the VFS does not run a second password KDF. Callers
// must never log the returned driver's DSN (it exists only inside this call).
func Open(protection *security.Manager, path, accountID string) (*sql.DB, error) {
	if accountID == "" {
		return nil, errors.New("securedb: empty account id")
	}
	return open(protection, path, "sqlite/"+accountID)
}

// OpenShared opens a database that belongs to no single account, such as the
// list of accounts. Its key is derived for name in a namespace of its own, so
// it never equals an account database's key.
func OpenShared(protection *security.Manager, path, name string) (*sql.DB, error) {
	if name == "" {
		return nil, errors.New("securedb: empty database name")
	}
	return open(protection, path, "sqlite-shared/"+name)
}

func open(protection *security.Manager, path, purpose string) (*sql.DB, error) {
	if protection == nil {
		return nil, security.ErrNotProtected
	}
	key, err := protection.DeriveKey(purpose)
	if err != nil {
		return nil, err
	}
	defer clear(key)
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(abs)}
	query := u.Query()
	query.Set("vfs", "adiantum")
	query.Set("hexkey", hex.EncodeToString(key))
	u.RawQuery = query.Encode()
	db, err := sql.Open("sqlite3", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA temp_store = memory"); err != nil {
		db.Close()
		return nil, fmt.Errorf("securedb: initialize: %w", err)
	}
	return db, nil
}
