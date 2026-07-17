package state

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

type DB struct {
	Conn *sql.DB
}

// Open initializes the SQLite database at the given path.
func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if _, err := db.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("failed to execute schema: %w", err)
	}

	// Dynamic column migrations for traffic tracking
	db.Exec("ALTER TABLE clients ADD COLUMN bytes_in INTEGER DEFAULT 0")
	db.Exec("ALTER TABLE clients ADD COLUMN bytes_out INTEGER DEFAULT 0")
	db.Exec("ALTER TABLE clients ADD COLUMN max_bytes INTEGER DEFAULT 0")
	db.Exec("ALTER TABLE clients ADD COLUMN download_speed TEXT DEFAULT ''")
	db.Exec("ALTER TABLE clients ADD COLUMN upload_speed TEXT DEFAULT ''")

	// Dynamic column migrations for LOPDP consent
	db.Exec("ALTER TABLE clients ADD COLUMN consent_given INTEGER DEFAULT 0")
	db.Exec("ALTER TABLE clients ADD COLUMN consent_timestamp INTEGER DEFAULT NULL")

	return &DB{Conn: db}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.Conn.Close()
}
