// SPDX-License-Identifier: Unlicense OR MIT

package account

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gotd/td/tg"
	_ "github.com/ncruces/go-sqlite3/driver"

	"komarugram/internal/messenger/securedb"
	"komarugram/internal/messenger/security"
)

// MaxCardPhoto bounds the avatar kept in a card. Telegram's small profile
// photo is 160×160 and a few kilobytes; anything much larger is not one.
const MaxCardPhoto = 256 << 10

// Card is what the account list shows about an account that is not
// connected: who it is and its small profile photo. It is refreshed whenever
// the account connects.
type Card struct {
	FirstName string
	LastName  string
	Username  string
	Phone     string
	// PhotoID is the profile photo Photo was downloaded from, 0 for none.
	// A card whose PhotoID matches the account's current photo need not
	// download it again.
	PhotoID int64
	// Photo is the small profile photo as the server sent it (JPEG).
	Photo []byte
	// Premium is set when the account has Telegram Premium.
	Premium bool
}

// Name returns the full name.
func (c Card) Name() string {
	if c.LastName == "" {
		return c.FirstName
	}
	return c.FirstName + " " + c.LastName
}

// userCard is the card of u without its photo.
func userCard(u *tg.User) Card {
	c := Card{FirstName: u.FirstName, LastName: u.LastName, Username: u.Username}
	if u.Phone != "" {
		c.Phone = "+" + u.Phone
	}
	return c
}

// record is one row of the registry.
type record struct {
	ID     string
	UserID int64
	DC     int
	KeyID  string
	Card   Card
}

const registrySchemaVersion = 2

const registrySchema = `
CREATE TABLE IF NOT EXISTS accounts(
	id         TEXT PRIMARY KEY,
	user_id    INTEGER NOT NULL UNIQUE,
	dc         INTEGER NOT NULL,
	key_id     TEXT NOT NULL UNIQUE,
	position   INTEGER NOT NULL,
	first_name TEXT NOT NULL DEFAULT '',
	last_name  TEXT NOT NULL DEFAULT '',
	username   TEXT NOT NULL DEFAULT '',
	phone      TEXT NOT NULL DEFAULT '',
	photo_id   INTEGER NOT NULL DEFAULT 0,
	photo      BLOB,
	added      INTEGER NOT NULL,
	updated    INTEGER NOT NULL DEFAULT 0,
	premium    INTEGER NOT NULL DEFAULT 0
) WITHOUT ROWID;`

// registryUpgrades bring a registry of schema version i+1 to i+2.
var registryUpgrades = []string{
	`ALTER TABLE accounts ADD COLUMN premium INTEGER NOT NULL DEFAULT 0`,
}

// registry is the SQLite database of every account of the client: who the
// account is, its home data center and the fingerprint of its auth key, and
// the card the account list shows. The auth keys are not in it: each stays in
// the session file of its account directory, as gotd reads and writes it.
//
// Without local-data protection it is a plaintext file (path.plain). With it,
// an Adiantum-encrypted one (path.secure) keyed for this database alone, so
// nothing about the accounts — not even which there are — can be read before
// the user unlocks. Turning protection on copies the rows over; the registry
// opens lazily, and while protection is locked it cannot open at all.
type registry struct {
	path       string
	protection *security.Manager

	mu        sync.Mutex
	db        *sql.DB
	encrypted bool
}

func (r *registry) plainPath() string  { return r.path + ".plain" }
func (r *registry) securePath() string { return r.path + ".secure" }

const recordColumns = `id, user_id, dc, key_id, first_name, last_name, username, phone, photo_id, photo, premium`

func scanRecord(rows *sql.Rows) (record, error) {
	var rec record
	c := &rec.Card
	err := rows.Scan(&rec.ID, &rec.UserID, &rec.DC, &rec.KeyID, &c.FirstName, &c.LastName, &c.Username, &c.Phone, &c.PhotoID, &c.Photo, &c.Premium)
	if len(c.Photo) > MaxCardPhoto {
		c.Photo, c.PhotoID = nil, 0
	}
	return rec, err
}

// list returns every account in the order they were added.
func (r *registry) list() ([]record, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	db, err := r.openLocked()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT ` + recordColumns + ` FROM accounts ORDER BY position`)
	if err != nil {
		return nil, fmt.Errorf("account: read account list: %w", err)
	}
	defer rows.Close()
	var records []record
	for rows.Next() {
		rec, err := scanRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("account: read account list: %w", err)
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

// insert adds rec after every other account. A user or key already listed
// is ErrAccountExists.
func (r *registry) insert(rec record) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	db, err := r.openLocked()
	if err != nil {
		return err
	}
	c := rec.Card
	_, err = db.Exec(`INSERT INTO accounts(id, user_id, dc, key_id, position, first_name, last_name, username, phone, photo_id, photo, premium, added)
		VALUES(?, ?, ?, ?, (SELECT coalesce(max(position), 0) + 1 FROM accounts), ?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.ID, rec.UserID, rec.DC, rec.KeyID, c.FirstName, c.LastName, c.Username, c.Phone, c.PhotoID, c.Photo, c.Premium, time.Now().Unix())
	if err != nil && strings.Contains(err.Error(), "UNIQUE") {
		return fmt.Errorf("%w: %s", ErrAccountExists, rec.ID)
	}
	if err != nil {
		return fmt.Errorf("account: add to account list: %w", err)
	}
	return nil
}

func (r *registry) saveCard(id string, c Card) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	db, err := r.openLocked()
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE accounts SET first_name = ?, last_name = ?, username = ?, phone = ?, photo_id = ?, photo = ?, premium = ?, updated = ? WHERE id = ?`,
		c.FirstName, c.LastName, c.Username, c.Phone, c.PhotoID, c.Photo, c.Premium, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("account: save card: %w", err)
	}
	return nil
}

func (r *registry) remove(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	db, err := r.openLocked()
	if err != nil {
		return err
	}
	if _, err := db.Exec(`DELETE FROM accounts WHERE id = ?`, id); err != nil {
		return fmt.Errorf("account: remove from account list: %w", err)
	}
	return nil
}

// protect moves the rows into the encrypted database. It runs as part of the
// protection migration, when the key is already available.
func (r *registry) protect() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, err := r.openLocked()
	return err
}

func (r *registry) unprotect() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	db, err := r.openLocked()
	if err != nil {
		return err
	}
	if !r.encrypted {
		return nil
	}
	if err := securedb.CopyPlain(db, r.plainPath()); err != nil {
		return err
	}
	_, err = r.openLocked()
	return err
}

func (r *registry) close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.db == nil {
		return nil
	}
	err := r.db.Close()
	r.db = nil
	return err
}

// openLocked returns the database in the form protection asks for now,
// moving a plaintext one into the encrypted form first. r.mu must be held.
func (r *registry) openLocked() (*sql.DB, error) {
	encrypted := r.protection != nil && r.protection.Enabled()
	if encrypted && r.protection.Disabling() {
		if _, err := os.Stat(r.plainPath()); err == nil {
			encrypted = false
		}
	}
	if r.db != nil && r.encrypted == encrypted {
		return r.db, nil
	}
	if r.db != nil {
		if err := r.db.Close(); err != nil {
			return nil, err
		}
		r.db = nil
	}
	var db *sql.DB
	var err error
	if encrypted {
		if err = r.migrateLocked(); err != nil {
			return nil, err
		}
		db, err = openRegistry(r.securePath(), r.protection)
	} else {
		db, err = openRegistry(r.plainPath(), nil)
	}
	if err != nil {
		return nil, err
	}
	r.db, r.encrypted = db, encrypted
	return db, nil
}

// migrateLocked copies a plaintext database into a new encrypted one and
// removes it. The copy is built beside the final file and renamed into place,
// so a crash leaves either the plaintext database or the complete encrypted
// one, and the migration runs again on the next open.
func (r *registry) migrateLocked() error {
	if _, err := os.Stat(r.plainPath()); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	if _, err := os.Stat(r.securePath()); err == nil {
		// The encrypted copy was completed before a crash.
		return removeDatabase(r.plainPath())
	}
	// Opening the plaintext registry brings its schema up to date, so that
	// its rows have the columns of the new one.
	plainDB, err := openRegistry(r.plainPath(), nil)
	if err != nil {
		return err
	}
	plainDB.Close()
	pending := r.securePath() + ".tmp"
	if err := removeDatabase(pending); err != nil {
		return err
	}
	// The encrypted database starts empty, so the plaintext one is attached
	// to it and copied over in one statement.
	secure, err := openRegistry(pending, r.protection)
	if err != nil {
		return err
	}
	// The attached file is named by URI to read it through the ordinary VFS
	// rather than the encrypting one of the connection.
	plain := url.URL{Scheme: "file", Path: filepath.ToSlash(r.plainPath()), RawQuery: "vfs=os&mode=ro"}
	_, err = secure.Exec(`ATTACH DATABASE ? AS plain`, plain.String())
	if err == nil {
		_, err = secure.Exec(`INSERT INTO main.accounts SELECT * FROM plain.accounts`)
	}
	if err == nil {
		_, err = secure.Exec(`DETACH DATABASE plain`)
	}
	if closeErr := secure.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(pending, r.securePath())
	}
	if err != nil {
		removeDatabase(pending)
		return fmt.Errorf("account: encrypt account list: %w", err)
	}
	return removeDatabase(r.plainPath())
}

// openRegistry opens the registry at path, encrypted when protection is set,
// and creates its table.
func openRegistry(path string, protection *security.Manager) (*sql.DB, error) {
	var db *sql.DB
	var err error
	if protection != nil {
		db, err = securedb.OpenShared(protection, path, "accounts")
	} else {
		// Create it private before SQLite does.
		var f *os.File
		if f, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600); err != nil {
			return nil, err
		}
		f.Close()
		db, err = sql.Open("sqlite3", path)
	}
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	var version int
	// Another copy of the client may be writing; wait for it rather than fail.
	if _, err = db.Exec(`PRAGMA busy_timeout = 5000`); err == nil {
		err = db.QueryRow(`PRAGMA user_version`).Scan(&version)
	}
	if err == nil && version > registrySchemaVersion {
		err = fmt.Errorf("written by a newer version (schema %d)", version)
	}
	if err == nil && version == 0 {
		_, err = db.Exec(registrySchema)
	} else {
		for v := version; err == nil && v < registrySchemaVersion; v++ {
			_, err = db.Exec(registryUpgrades[v-1])
		}
	}
	if err == nil {
		_, err = db.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, registrySchemaVersion))
	}
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("account: open account list: %w", err)
	}
	return db, nil
}

// removeDatabase removes a database file and what SQLite may have left
// beside it.
func removeDatabase(path string) error {
	for _, suffix := range []string{"", "-journal", "-wal", "-shm"} {
		if err := os.Remove(path + suffix); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}
