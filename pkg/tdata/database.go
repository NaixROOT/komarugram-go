// SPDX-License-Identifier: Unlicense

package tdata

import (
	"database/sql"
	"encoding/hex"
	"os"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
)

const createTelethonQuery = `
CREATE TABLE entities (
    id       INTEGER PRIMARY KEY,
    hash     INTEGER NOT NULL,
    username TEXT,
    phone    INTEGER,
    name     TEXT,
    date     INTEGER
);
CREATE TABLE sent_files (
    md5_digest BLOB,
    file_size  INTEGER,
    type       INTEGER,
    id         INTEGER,
    hash       INTEGER,
    PRIMARY KEY (
        md5_digest,
        file_size,
        type
    )
);
CREATE TABLE sessions (
    dc_id          INTEGER PRIMARY KEY,
    server_address TEXT,
    port           INTEGER,
    auth_key       BLOB,
    takeout_id     INTEGER
);
CREATE TABLE update_state (
    id   INTEGER PRIMARY KEY,
    pts  INTEGER,
    qts  INTEGER,
    date INTEGER,
    seq  INTEGER
);
CREATE TABLE version (
    version INTEGER PRIMARY KEY
);
INSERT INTO version(version) VALUES(7);
`

const createPyrogramQuery = `
CREATE TABLE peers (
    id             INTEGER PRIMARY KEY,
    access_hash    INTEGER,
    type           INTEGER NOT NULL,
    username       TEXT,
    phone_number   TEXT,
    last_update_on INTEGER NOT NULL
                           DEFAULT (CAST (STRFTIME('%s', 'now') AS INTEGER) ) 
);
CREATE TABLE sessions (
    dc_id     INTEGER PRIMARY KEY,
    test_mode INTEGER,
    auth_key  BLOB,
    date      INTEGER NOT NULL,
    user_id   INTEGER,
    is_bot    INTEGER
);
CREATE TABLE version (
    number INTEGER PRIMARY KEY
);
INSERT INTO version(number) VALUES(2);
`

func ReadDatabase(source string) ([]TDataSession, error) {
	db, err := sql.Open("sqlite3", source)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	db.SetMaxOpenConns(1)

	probe, err := db.Query("SELECT name FROM pragma_table_info('sessions');")
	if err != nil {
		return nil, err
	}

	var hasUserId bool

	for probe.Next() {
		var name string
		probe.Scan(&name)
		if name == "user_id" {
			hasUserId = true
		}
	}

	var rows *sql.Rows
	if hasUserId {
		rows, err = db.Query(`SELECT auth_key, dc_id, user_id FROM sessions`)
	} else {
		rows, err = db.Query(`SELECT auth_key, dc_id FROM sessions`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TDataSession

	for rows.Next() {
		result := TDataSession{}
		var authKey []byte
		if hasUserId {
			err = rows.Scan(&authKey, &result.DC, &result.UserID)
		} else {
			err = rows.Scan(&authKey, &result.DC)
		}
		if err != nil {
			return results, err
		}
		result.AuthKey = hex.EncodeToString(authKey)
		results = append(results, result)
	}

	return results, err
}

func ReadDatabaseBytes(data []byte) ([]TDataSession, error) {
	tmpFile, err := os.CreateTemp("", "sqlite-*.db")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		return nil, err
	}
	tmpFile.Close()

	return ReadDatabase(tmpFile.Name())
}

func CreateTelethonDatabase(input []TDataSession, source string) error {
	db, err := sql.Open("sqlite3", source)
	if err != nil {
		return err
	}
	defer db.Close()

	db.SetMaxOpenConns(1)

	if _, err = db.Exec(createTelethonQuery); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO sessions(dc_id, server_address, port, auth_key) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, result := range input {
		serverAddress, err := GetServerAddress(result.DC)
		if err != nil {
			return err
		}
		authKeyBytes, err := hex.DecodeString(result.AuthKey)
		if err != nil {
			return err
		}
		if _, err := stmt.Exec(result.DC, serverAddress, 443, authKeyBytes); err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func CreateTelethonDatabaseBytes(input []TDataSession) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "sqlite-*.db")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if err = CreateTelethonDatabase(input, tmpFile.Name()); err != nil {
		return nil, err
	}

	return os.ReadFile(tmpFile.Name())
}

func CreatePyrogramDatabase(input []TDataSession, source string) error {
	db, err := sql.Open("sqlite3", source)
	if err != nil {
		return err
	}
	defer db.Close()

	db.SetMaxOpenConns(1)

	if _, err = db.Exec(createPyrogramQuery); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO sessions(dc_id, test_mode, auth_key, date, user_id) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, result := range input {
		authKeyBytes, err := hex.DecodeString(result.AuthKey)
		if err != nil {
			return err
		}
		if _, err := stmt.Exec(result.DC, 0, authKeyBytes, time.Now().Unix(), result.UserID); err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func CreatePyrogramDatabaseBytes(input []TDataSession) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "sqlite-*.db")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if err = CreatePyrogramDatabase(input, tmpFile.Name()); err != nil {
		return nil, err
	}

	return os.ReadFile(tmpFile.Name())
}
