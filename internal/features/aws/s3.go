package awsfeat

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/mbeniwal-imwe/ark/internal/storage"
)

// S3Service handles S3 operations
type S3Service struct {
	Client *Client
	S3     *s3.Client
}

// NewS3Service creates a new S3 service for a profile
func NewS3Service(ctx context.Context, db *storage.Database, profileName string) (*S3Service, error) {
	client, err := NewClient(ctx, db, profileName)
	if err != nil {
		return nil, err
	}
	return &S3Service{Client: client, S3: s3.NewFromConfig(client.Config)}, nil
}

// ListBuckets lists S3 buckets
func (s *S3Service) ListBuckets(ctx context.Context) ([]types.Bucket, error) {
	out, err := s.S3.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("list buckets failed: %w", err)
	}
	return out.Buckets, nil
}

// ListObjects lists objects under a bucket/prefix
func (s *S3Service) ListObjects(ctx context.Context, bucket, prefix string) ([]types.Object, error) {
	out, err := s.S3.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("list objects failed: %w", err)
	}
	return out.Contents, nil
}

// UploadFile uploads a local file to s3://bucket/key
func (s *S3Service) UploadFile(ctx context.Context, localPath, bucket, key string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = s.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   f,
	})
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	return nil
}

// DownloadFile downloads s3://bucket/key to localPath (directory or file)
func (s *S3Service) DownloadFile(ctx context.Context, bucket, key, localPath string) error {
	out, err := s.S3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer out.Body.Close()

	dstPath := localPath
	if fi, err := os.Stat(localPath); err == nil && fi.IsDir() {
		dstPath = filepath.Join(localPath, filepath.Base(key))
	}
	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, out.Body)
	return err
}
