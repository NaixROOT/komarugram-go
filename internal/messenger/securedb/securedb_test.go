// SPDX-License-Identifier: Unlicense OR MIT

package securedb

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"komarugram/internal/messenger/security"
)

type testTPM struct{ secret []byte }

func (t *testTPM) Probe() error { return nil }
func (t *testTPM) Seal(secret, auth []byte) ([]byte, []byte, error) {
	t.secret = append([]byte(nil), secret...)
	return append([]byte(nil), auth...), []byte("device-bound"), nil
}
func (t *testTPM) Unseal(public, private, auth []byte) ([]byte, error) {
	if !bytes.Equal(public, auth) || string(private) != "device-bound" {
		return nil, errors.New("denied")
	}
	return append([]byte(nil), t.secret...), nil
}

func TestEncryptedDatabaseIsAccountScoped(t *testing.T) {
	root := t.TempDir()
	protection, err := security.OpenPath(filepath.Join(root, "security.json"), new(testTPM))
	if err != nil {
		t.Fatal(err)
	}
	if err := protection.Enable(context.Background(), "password"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "history.db")
	db, err := Open(protection, path, "account-a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE messages(body TEXT); INSERT INTO messages VALUES ('private-message-marker')`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("SQLite format 3")) || bytes.Contains(raw, []byte("private-message-marker")) {
		t.Fatal("database contains recognizable plaintext")
	}

	wrong, err := Open(protection, path, "account-b")
	if err != nil {
		return // A wrong key may be rejected while applying the first PRAGMA.
	}
	defer wrong.Close()
	var count int
	if err := wrong.QueryRow("SELECT count(*) FROM messages").Scan(&count); err == nil {
		t.Fatalf("database opened with another account key, rows=%d", count)
	}
}
