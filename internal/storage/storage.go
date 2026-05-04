package storage

import (
	"context"
	"io"
)

// Provider defines the interface for saving and retrieving snapshots.
type Provider interface {
	// Save stores a new snapshot, reading from the provided io.Reader.
	// It should handle compression internally if configured.
	Save(ctx context.Context, snapshotID string, r io.Reader) (int64, error)

	// Retrieve returns an io.ReadCloser to read a snapshot's contents.
	// It should handle decompression internally if necessary.
	Retrieve(ctx context.Context, snapshotID string) (io.ReadCloser, error)
	
	// Delete removes a snapshot.
	Delete(ctx context.Context, snapshotID string) error
}
