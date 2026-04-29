package minio

import (
	"context"
	"fmt"
)

// BucketType represents the type of storage bucket
type BucketType string

const (
	// BucketReports stores generated reports and documents
	BucketReports BucketType = "reports"
	// BucketUploads stores user-uploaded files
	BucketUploads BucketType = "uploads"
	// BucketCache stores temporary cache data
	BucketCache BucketType = "cache"
)

// BucketManager manages multiple buckets with type-based organization
type BucketManager struct {
	client *Client
	prefix string // optional prefix for bucket names (e.g., "prod-", "staging-")
	region string
}

// NewBucketManager creates a new bucket manager
func NewBucketManager(client *Client, prefix, region string) *BucketManager {
	return &BucketManager{
		client: client,
		prefix: prefix,
		region: region,
	}
}

// GetBucketName returns the full bucket name for a given type
func (m *BucketManager) GetBucketName(bucketType BucketType) string {
	return m.prefix + string(bucketType)
}

// EnsureBucketsExist creates necessary buckets if they don't exist
func (m *BucketManager) EnsureBucketsExist(ctx context.Context, bucketTypes ...BucketType) error {
	if len(bucketTypes) == 0 {
		// Default buckets
		bucketTypes = []BucketType{BucketReports, BucketUploads, BucketCache}
	}

	for _, bType := range bucketTypes {
		bucketName := m.GetBucketName(bType)

		// Check if bucket exists
		exists, err := m.client.BucketExists(ctx, bucketName)
		if err != nil {
			return fmt.Errorf("failed to check bucket %q existence: %w", bucketName, err)
		}

		// Create if doesn't exist
		if !exists {
			if err := m.client.CreateBucket(ctx, bucketName, m.region); err != nil {
				return fmt.Errorf("failed to create bucket %q: %w", bucketName, err)
			}
		}
	}

	return nil
}

// GetBucketOrDefault returns the bucket name for a type, or a default bucket if the type is invalid
func (m *BucketManager) GetBucketOrDefault(bucketType BucketType) string {
	if bucketType == "" {
		return m.GetBucketName(BucketUploads) // default
	}
	return m.GetBucketName(bucketType)
}

// ListAllBuckets lists all managed buckets (returns only our managed buckets)
func (m *BucketManager) ListAllBuckets(ctx context.Context) ([]string, error) {
	buckets := []string{
		m.GetBucketName(BucketReports),
		m.GetBucketName(BucketUploads),
		m.GetBucketName(BucketCache),
	}
	return buckets, nil
}

// ValidateBucketType checks if a bucket type is valid
func ValidateBucketType(bt BucketType) bool {
	switch bt {
	case BucketReports, BucketUploads, BucketCache:
		return true
	default:
		return false
	}
}
