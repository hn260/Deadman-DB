package state

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// Snapshot represents a backup snapshot in the database.
type Snapshot struct {
	ID        string `db:"id"`
	DBName    string `db:"db_name"`
	Timestamp int64  `db:"timestamp"`
	Size      int64  `db:"size"`
	Status    string `db:"status"`
	FilePath  string `db:"file_path"`
}

const schema = `
CREATE TABLE IF NOT EXISTS snapshots (
    id TEXT PRIMARY KEY,
    db_name TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    size INTEGER NOT NULL,
    status TEXT NOT NULL,
    file_path TEXT NOT NULL
);
`

// InitDB initializes the SQLite database at the given path.
func InitDB(dbPath string) (*sqlx.DB, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	// Create tables
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return db, nil
}
