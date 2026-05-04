package engine

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

type PostgresEngine struct{}

// NewPostgresEngine returns a new instance of PostgresEngine.
func NewPostgresEngine() *PostgresEngine {
	return &PostgresEngine{}
}

// Backup runs pg_dump and writes its stdout directly to the provided writer.
func (p *PostgresEngine) Backup(ctx context.Context, dbURL string, w io.Writer) error {
	// Use standard pg_dump format to dump everything to stdout.
	// Clean format, no owner information (to make restore easier on other users).
	cmd := exec.CommandContext(ctx, "pg_dump", "--clean", "--no-owner", "--no-privileges", dbURL)
	
	// Stream standard output directly to the writer (which will likely be gzip)
	cmd.Stdout = w
	
	// We could capture stderr for better error reporting if needed,
	// but for now let's just run it.
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump failed: %w", err)
	}
	
	return nil
}

// Restore runs pg_restore, reading from the provided reader.
func (p *PostgresEngine) Restore(ctx context.Context, dbURL string, r io.Reader) error {
	// We read standard SQL from stdin because pg_dump was run without a specific format (defaults to plain text SQL)
	cmd := exec.CommandContext(ctx, "psql", dbURL)
	cmd.Stdin = r
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("psql restore failed: %w", err)
	}
	
	return nil
}
