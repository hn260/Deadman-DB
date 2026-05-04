package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Provider struct {
	client     *s3.Client
	bucketName string
	prefix     string
}

// NewS3Provider initializes a new S3 storage provider.
// It uses the default AWS credentials chain (env vars, ~/.aws/credentials, etc.)
func NewS3Provider(ctx context.Context, region, bucketName, prefix string) (*S3Provider, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Provider{
		client:     client,
		bucketName: bucketName,
		prefix:     prefix,
	}, nil
}

// Save streams the data to S3 using the Uploader manager which handles multipart uploads efficiently.
func (s *S3Provider) Save(ctx context.Context, snapshotID string, r io.Reader) (int64, error) {
	key := fmt.Sprintf("%s/%s.sql.gz", s.prefix, snapshotID)

	uploader := manager.NewUploader(s.client)

	// Note: We don't have the size since we stream, but Uploader handles chunking automatically.
	// We wrap the reader to count bytes if we want to return the size.
	counter := &writeCounterR{Reader: r}

	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   counter,
	})

	if err != nil {
		return 0, fmt.Errorf("failed to upload to s3: %w", err)
	}

	return counter.read, nil
}

// Retrieve downloads the file stream from S3.
func (s *S3Provider) Retrieve(ctx context.Context, snapshotID string) (io.ReadCloser, error) {
	key := fmt.Sprintf("%s/%s.sql.gz", s.prefix, snapshotID)

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get object from s3: %w", err)
	}

	return resp.Body, nil
}

func (s *S3Provider) Delete(ctx context.Context, snapshotID string) error {
	key := fmt.Sprintf("%s/%s.sql.gz", s.prefix, snapshotID)

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	return err
}

// writeCounterR wraps an io.Reader and counts bytes read.
type writeCounterR struct {
	io.Reader
	read int64
}

func (w *writeCounterR) Read(p []byte) (int, error) {
	n, err := w.Reader.Read(p)
	w.read += int64(n)
	return n, err
}
