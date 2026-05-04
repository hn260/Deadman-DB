package engine

import (
	"context"
	"io"
)

// Engine defines the interface for database backup operations.
type Engine interface {
	// Backup runs the backup tool (e.g., pg_dump) and writes the output to the provided io.Writer.
	Backup(ctx context.Context, dbURL string, w io.Writer) error

	// Restore reads from the provided io.Reader and restores the data to the database.
	Restore(ctx context.Context, dbURL string, r io.Reader) error
}
