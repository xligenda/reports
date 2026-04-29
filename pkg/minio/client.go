package minio

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client wraps the minio.Client with common operations.
type Client struct {
	client *minio.Client
	bucket string
}

// NewMinioClient creates a new Minio client with the given configuration.
func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool, bucket string) (*Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		client: minioClient,
		bucket: bucket,
	}, nil
}

// PutObject uploads an object to the Minio bucket.
func (c *Client) PutObject(ctx context.Context, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return c.client.PutObject(ctx, c.bucket, objectName, reader, objectSize, opts)
}

// GetObject downloads an object from the Minio bucket.
func (c *Client) GetObject(ctx context.Context, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return c.client.GetObject(ctx, c.bucket, objectName, opts)
}

// ListObjects lists objects in the Minio bucket.
func (c *Client) ListObjects(ctx context.Context, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	return c.client.ListObjects(ctx, c.bucket, opts)
}

// RemoveObject deletes an object from the Minio bucket.
func (c *Client) RemoveObject(ctx context.Context, objectName string, opts minio.RemoveObjectOptions) error {
	return c.client.RemoveObject(ctx, c.bucket, objectName, opts)
}

// BucketExists checks if the bucket exists.
func (c *Client) BucketExists(ctx context.Context) (bool, error) {
	return c.client.BucketExists(ctx, c.bucket)
}

// Bucket returns the configured bucket name.
func (c *Client) Bucket() string {
	return c.bucket
}
