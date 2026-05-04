package storage

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalProvider struct {
	baseDir string
}

// NewLocalProvider creates a new local storage provider that saves files to baseDir.
func NewLocalProvider(baseDir string) (*LocalProvider, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local storage directory: %w", err)
	}
	return &LocalProvider{baseDir: baseDir}, nil
}

// Save reads from r, compresses using gzip, and writes to a local file.
// It returns the size of the uncompressed data read (or compressed size depending on metric desired).
// Here we return the compressed size written.
func (l *LocalProvider) Save(ctx context.Context, snapshotID string, r io.Reader) (int64, error) {
	filePath := filepath.Join(l.baseDir, snapshotID+".sql.gz")
	
	file, err := os.Create(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create snapshot file: %w", err)
	}
	defer file.Close()

	// Wrap the file with a gzip writer
	gw := gzip.NewWriter(file)
	defer gw.Close()

	// Create a writer that counts bytes written
	counter := &writeCounter{Writer: gw}

	// io.Copy handles the streaming efficiently
	if _, err := io.Copy(counter, r); err != nil {
		return 0, fmt.Errorf("failed to stream data to storage: %w", err)
	}

	if err := gw.Close(); err != nil {
		return 0, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return counter.written, nil
}

// Retrieve returns an io.ReadCloser for the uncompressed data.
func (l *LocalProvider) Retrieve(ctx context.Context, snapshotID string) (io.ReadCloser, error) {
	filePath := filepath.Join(l.baseDir, snapshotID+".sql.gz")
	
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open snapshot file: %w", err)
	}

	gr, err := gzip.NewReader(file)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	// We need to return a ReadCloser that closes both the gzip reader and the underlying file.
	return &multiCloser{
		Reader: gr,
		closers: []io.Closer{gr, file},
	}, nil
}

func (l *LocalProvider) Delete(ctx context.Context, snapshotID string) error {
	filePath := filepath.Join(l.baseDir, snapshotID+".sql.gz")
	return os.Remove(filePath)
}

// writeCounter wraps an io.Writer and counts bytes written.
type writeCounter struct {
	io.Writer
	written int64
}

func (w *writeCounter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	w.written += int64(n)
	return n, err
}

// multiCloser wraps an io.Reader and multiple io.Closers.
type multiCloser struct {
	io.Reader
	closers []io.Closer
}

func (m *multiCloser) Close() error {
	var firstErr error
	for _, c := range m.closers {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
