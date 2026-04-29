package storage

import (
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// PutObjectInput holds parameters for a single-part PutObject call.
type PutObjectInput struct {
	Bucket      string
	Key         string
	Body        io.Reader
	ContentType string
	Metadata    map[string]string
	ACL         types.ObjectCannedACL
	// ServerSideEncryption enables AES256 or aws:kms encryption.
	ServerSideEncryption types.ServerSideEncryption
}

// PutObjectOutput contains the result of a PutObject call.
type PutObjectOutput struct {
	ETag      string
	VersionID string
}

// GetObjectInput holds parameters for a GetObject call.
type GetObjectInput struct {
	Bucket    string
	Key       string
	VersionID string
	// Range follows HTTP range syntax e.g. "bytes=0-1023"
	Range string
}

// GetObjectOutput wraps the object body and its metadata.
// Callers must close Body when done.
type GetObjectOutput struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
	ETag          string
	LastModified  time.Time
	Metadata      map[string]string
	VersionID     string
}

// DeleteObjectInput holds parameters for a DeleteObject call.
type DeleteObjectInput struct {
	Bucket    string
	Key       string
	VersionID string
}

// DeleteError records a key that failed during a batch delete.
type DeleteError struct {
	Key     string
	Code    string
	Message string
}

// HeadObjectInput holds parameters for a HeadObject call.
type HeadObjectInput struct {
	Bucket    string
	Key       string
	VersionID string
}

// HeadObjectOutput contains metadata for an object without its body.
type HeadObjectOutput struct {
	ContentType   string
	ContentLength int64
	ETag          string
	LastModified  time.Time
	Metadata      map[string]string
	VersionID     string
	StorageClass  types.StorageClass
}

// CopyObjectInput holds parameters for a server-side copy.
type CopyObjectInput struct {
	SourceBucket string
	SourceKey    string
	DestBucket   string
	DestKey      string
	Metadata     map[string]string
	// MetadataDirective is COPY (default) or REPLACE.
	MetadataDirective types.MetadataDirective
	ACL               types.ObjectCannedACL
}

// UploadInput wraps manager.UploadInput for multipart-aware uploads.
type UploadInput struct {
	Bucket      string
	Key         string
	Body        io.Reader
	ContentType string
	Metadata    map[string]string
	ACL         types.ObjectCannedACL
	// PartSize overrides the default 5 MiB multipart part size.
	PartSize int64
	// Concurrency overrides the default upload concurrency (5).
	Concurrency int
}

// UploadOutput contains the result of a (possibly multipart) upload.
type UploadOutput struct {
	Location  string
	VersionID string
	ETag      string
}

// DownloadInput wraps manager.DownloadInput.
type DownloadInput struct {
	Bucket      string
	Key         string
	VersionID   string
	WriterAt    io.WriterAt
	PartSize    int64 // default 5 MiB
	Concurrency int   // default 5
}

// ListObjectsInput controls pagination and filtering for list calls.
type ListObjectsInput struct {
	Bucket     string
	Prefix     string
	Delimiter  string
	MaxKeys    int32
	StartAfter string
}

// ListObjectsOutput contains a page of object results.
type ListObjectsOutput struct {
	Objects               []ObjectInfo
	CommonPrefixes        []string
	IsTruncated           bool
	NextContinuationToken string
}

// ObjectInfo is a lightweight representation of an S3 object.
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
	StorageClass types.StorageClass
}

// PresignInput holds parameters for generating presigned URLs.
type PresignInput struct {
	Bucket  string
	Key     string
	Expires time.Duration
	// ContentType is used only for presigned PUT requests.
	ContentType string
}

// CreateBucketInput holds parameters for bucket creation.
type CreateBucketInput struct {
	Bucket string
	// ACL e.g. types.BucketCannedACLPrivate
	ACL types.BucketCannedACL
	// Region overrides the client's default region for the new bucket.
	Region string
}

// BucketInfo is a lightweight representation of an S3 bucket.
type BucketInfo struct {
	Name         string
	CreationDate time.Time
}
