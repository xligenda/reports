package minio

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client wraps the minio.Client with multi-bucket support.
type Client struct {
	client *minio.Client
}

// NewMinioClient creates a new Minio client with the given configuration.
// Unlike the previous version, this client can work with multiple buckets.
func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) (*Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &Client{
		client: minioClient,
	}, nil
}

// PutObject uploads an object to the specified Minio bucket.
func (c *Client) PutObject(ctx context.Context, bucket, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	if err := validateBucketName(bucket); err != nil {
		return minio.UploadInfo{}, err
	}
	return c.client.PutObject(ctx, bucket, objectName, reader, objectSize, opts)
}

// GetObject downloads an object from the specified Minio bucket.
func (c *Client) GetObject(ctx context.Context, bucket, objectName string, opts minio.GetObjectOptions) (io.ReadCloser, error) {
	if err := validateBucketName(bucket); err != nil {
		return nil, err
	}
	obj, err := c.client.GetObject(ctx, bucket, objectName, opts)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// ListObjects lists objects in the specified Minio bucket.
func (c *Client) ListObjects(ctx context.Context, bucket string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	if err := validateBucketName(bucket); err != nil {
		// Return empty channel if bucket is invalid
		ch := make(chan minio.ObjectInfo)
		close(ch)
		return ch
	}
	return c.client.ListObjects(ctx, bucket, opts)
}

// RemoveObject deletes an object from the specified Minio bucket.
func (c *Client) RemoveObject(ctx context.Context, bucket, objectName string, opts minio.RemoveObjectOptions) error {
	if err := validateBucketName(bucket); err != nil {
		return err
	}
	return c.client.RemoveObject(ctx, bucket, objectName, opts)
}

// BucketExists checks if the bucket exists.
func (c *Client) BucketExists(ctx context.Context, bucket string) (bool, error) {
	if err := validateBucketName(bucket); err != nil {
		return false, err
	}
	return c.client.BucketExists(ctx, bucket)
}

// CreateBucket creates a new bucket in Minio.
func (c *Client) CreateBucket(ctx context.Context, bucket string, region string) error {
	if err := validateBucketName(bucket); err != nil {
		return err
	}
	return c.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: region})
}

// DeleteBucket removes a bucket from Minio (must be empty).
func (c *Client) DeleteBucket(ctx context.Context, bucket string) error {
	if err := validateBucketName(bucket); err != nil {
		return err
	}
	return c.client.RemoveBucket(ctx, bucket)
}

// HeadObject returns metadata about an object without downloading it.
func (c *Client) HeadObject(ctx context.Context, bucket, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	if err := validateBucketName(bucket); err != nil {
		return minio.ObjectInfo{}, err
	}
	return c.client.StatObject(ctx, bucket, objectName, opts)
}

// validateBucketName ensures bucket name is not empty.
func validateBucketName(bucket string) error {
	if bucket == "" {
		return fmt.Errorf("bucket name cannot be empty")
	}
	return nil
}
