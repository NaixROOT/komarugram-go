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
	"strings"

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
	uri, err := FileURI(tmp, url.Values{"vfs": {"os"}})
	if err != nil {
		return err
	}
	if _, err := source.Exec(`VACUUM main INTO ?`, uri); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("securedb: copy to plaintext: %w", err)
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	check, err := sql.Open("sqlite3", uri)
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
	hexKey := hex.EncodeToString(key)
	uri, err := FileURI(tmp, url.Values{"vfs": {"adiantum"}, "hexkey": {hexKey}})
	if err != nil {
		return err
	}
	if _, err := source.Exec(`VACUUM main INTO ?`, uri); err != nil {
		_ = os.Remove(tmp)
		// SQLite names the file it could not open by the whole URI.
		return fmt.Errorf("securedb: copy to encrypted: %w", redactedError{err, hexKey})
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
	hexKey := hex.EncodeToString(key)
	uri, err := FileURI(path, url.Values{"vfs": {"adiantum"}, "hexkey": {hexKey}})
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", uri)
	if err != nil {
		return nil, redactedError{err, hexKey}
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA temp_store = memory"); err != nil {
		db.Close()
		return nil, fmt.Errorf("securedb: initialize: %w", redactedError{err, hexKey})
	}
	return db, nil
}

// FileURI returns a "file:" URI that names path, made absolute, with query.
// The URI has no authority part: SQLite is built with
// SQLITE_ALLOW_URI_AUTHORITY, so file://C:/x would name the UNC path //C:/x
// on Windows, and without SQLITE_OS_WIN nothing drops the slash of
// file:///C:/x.
func FileURI(path string, query url.Values) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(abs), OmitHost: true, RawQuery: query.Encode()}
	return u.String(), nil
}

// redactedError hides a database key that SQLite put in an error message
// together with the URI it was given.
type redactedError struct {
	err    error
	hexKey string
}

func (e redactedError) Error() string {
	return strings.ReplaceAll(e.err.Error(), e.hexKey, "[key]")
}

func (e redactedError) Unwrap() error { return e.err }
