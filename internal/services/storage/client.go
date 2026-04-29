package storage

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	transport "github.com/aws/smithy-go/endpoints"
)

// Client is the interface that defines all supported S3 operations.
type Client interface {
	// Object operations
	PutObject(ctx context.Context, input PutObjectInput) (*PutObjectOutput, error)
	GetObject(ctx context.Context, input GetObjectInput) (*GetObjectOutput, error)
	DeleteObject(ctx context.Context, input DeleteObjectInput) error
	DeleteObjects(ctx context.Context, bucket string, keys []string) ([]DeleteError, error)
	HeadObject(ctx context.Context, input HeadObjectInput) (*HeadObjectOutput, error)
	CopyObject(ctx context.Context, input CopyObjectInput) error

	// Upload / Download (multipart-aware)
	Upload(ctx context.Context, input UploadInput) (*UploadOutput, error)
	Download(ctx context.Context, input DownloadInput) (int64, error)

	// Listing
	ListObjects(ctx context.Context, input ListObjectsInput) (*ListObjectsOutput, error)
	ListObjectsAll(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error)

	// Presigned URLs
	PresignGetObject(ctx context.Context, input PresignInput) (string, error)
	PresignPutObject(ctx context.Context, input PresignInput) (string, error)

	// Bucket operations
	CreateBucket(ctx context.Context, input CreateBucketInput) error
	DeleteBucket(ctx context.Context, bucket string) error
	BucketExists(ctx context.Context, bucket string) (bool, error)
	ListBuckets(ctx context.Context) ([]BucketInfo, error)
}

type S3Client struct {
	s3     *s3.Client
	psign  *s3.PresignClient
	tm     *transfermanager.Client
	region string
}

type Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string // optional
	Endpoint        string // optional
	ForcePathStyle  bool   // required for many custom endpoints
}

type staticResolver struct {
	Endpoint string
}

func (r *staticResolver) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (transport.Endpoint, error) {
	// If no custom endpoint is provided, we fall back to the default SDK resolver.
	if r.Endpoint == "" {
		return s3.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
	}

	u, err := url.Parse(r.Endpoint)
	if err != nil {
		return transport.Endpoint{}, fmt.Errorf("storage: invalid endpoint URL: %w", err)
	}

	return transport.Endpoint{
		URI: *u,
	}, nil
}
