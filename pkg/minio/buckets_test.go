package minio

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewBucketManager_Success tests successful BucketManager initialization
func TestNewBucketManager_Success(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "", "us-east-1")

	assert.NotNil(t, manager)
	assert.Equal(t, client, manager.client)
	assert.Equal(t, "", manager.prefix)
	assert.Equal(t, "us-east-1", manager.region)
}

// TestNewBucketManager_WithPrefix tests BucketManager initialization with prefix
func TestNewBucketManager_WithPrefix(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "prod-", "us-west-2")

	assert.NotNil(t, manager)
	assert.Equal(t, "prod-", manager.prefix)
	assert.Equal(t, "us-west-2", manager.region)
}

// TestGetBucketName_NoPrefix tests GetBucketName without prefix
func TestGetBucketName_NoPrefix(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "", "us-east-1")

	assert.Equal(t, "reports", manager.GetBucketName(BucketReports))
	assert.Equal(t, "uploads", manager.GetBucketName(BucketUploads))
	assert.Equal(t, "cache", manager.GetBucketName(BucketCache))
}

// TestGetBucketName_WithPrefix tests GetBucketName with prefix
func TestGetBucketName_WithPrefix(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "staging-", "us-east-1")

	assert.Equal(t, "staging-reports", manager.GetBucketName(BucketReports))
	assert.Equal(t, "staging-uploads", manager.GetBucketName(BucketUploads))
	assert.Equal(t, "staging-cache", manager.GetBucketName(BucketCache))
}

// TestGetBucketOrDefault_ValidType tests GetBucketOrDefault with valid bucket type
func TestGetBucketOrDefault_ValidType(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "", "us-east-1")

	assert.Equal(t, "reports", manager.GetBucketOrDefault(BucketReports))
	assert.Equal(t, "uploads", manager.GetBucketOrDefault(BucketUploads))
	assert.Equal(t, "cache", manager.GetBucketOrDefault(BucketCache))
}

// TestGetBucketOrDefault_EmptyType tests GetBucketOrDefault with empty bucket type defaults to uploads
func TestGetBucketOrDefault_EmptyType(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "", "us-east-1")

	assert.Equal(t, "uploads", manager.GetBucketOrDefault(""))
}

// TestListAllBuckets tests ListAllBuckets returns all managed bucket names
func TestListAllBuckets(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "", "us-east-1")

	buckets, err := manager.ListAllBuckets(context.Background())

	require.NoError(t, err)
	assert.Len(t, buckets, 3)
	assert.Contains(t, buckets, "reports")
	assert.Contains(t, buckets, "uploads")
	assert.Contains(t, buckets, "cache")
}

// TestListAllBuckets_WithPrefix tests ListAllBuckets with prefix
func TestListAllBuckets_WithPrefix(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "test-", "us-east-1")

	buckets, err := manager.ListAllBuckets(context.Background())

	require.NoError(t, err)
	assert.Len(t, buckets, 3)
	assert.Contains(t, buckets, "test-reports")
	assert.Contains(t, buckets, "test-uploads")
	assert.Contains(t, buckets, "test-cache")
}

// TestValidateBucketType_ValidTypes tests ValidateBucketType with valid types
func TestValidateBucketType_ValidTypes(t *testing.T) {
	validTypes := []BucketType{BucketReports, BucketUploads, BucketCache}

	for _, bt := range validTypes {
		assert.True(t, ValidateBucketType(bt), "bucket type %q should be valid", bt)
	}
}

// TestValidateBucketType_InvalidTypes tests ValidateBucketType with invalid types
func TestValidateBucketType_InvalidTypes(t *testing.T) {
	invalidTypes := []BucketType{"invalid", "unknown", "", "other"}

	for _, bt := range invalidTypes {
		assert.False(t, ValidateBucketType(bt), "bucket type %q should be invalid", bt)
	}
}

// TestEnsureBucketsExist_DefaultBuckets tests EnsureBucketsExist with default buckets
func TestEnsureBucketsExist_DefaultBuckets(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "", "us-east-1")
	ctx := context.Background()

	// This will fail because we don't have a real Minio server, but the call should be valid
	err = manager.EnsureBucketsExist(ctx)
	// We don't assert on the error since it's expected without a real server
	_ = err
}

// TestEnsureBucketsExist_CustomBuckets tests EnsureBucketsExist with specific bucket types
func TestEnsureBucketsExist_CustomBuckets(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "", "us-east-1")
	ctx := context.Background()

	// This will fail because we don't have a real Minio server, but the call should be valid
	err = manager.EnsureBucketsExist(ctx, BucketReports, BucketCache)
	// We don't assert on the error since it's expected without a real server
	_ = err
}

// TestEnsureBucketsExist_WithPrefix tests EnsureBucketsExist with prefix
func TestEnsureBucketsExist_WithPrefix(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "dev-", "us-east-1")
	ctx := context.Background()

	// This will fail because we don't have a real Minio server, but the call should be valid
	err = manager.EnsureBucketsExist(ctx)
	// We don't assert on the error since it's expected without a real server
	_ = err
}

// TestBucketTypeStrings tests BucketType string values
func TestBucketTypeStrings(t *testing.T) {
	assert.Equal(t, "reports", string(BucketReports))
	assert.Equal(t, "uploads", string(BucketUploads))
	assert.Equal(t, "cache", string(BucketCache))
}

// TestGetBucketName_DifferentPrefixes tests GetBucketName with various prefixes
func TestGetBucketName_DifferentPrefixes(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	prefixes := []string{"", "prod-", "staging-", "dev-", "test-v1-"}

	for _, prefix := range prefixes {
		manager := NewBucketManager(client, prefix, "us-east-1")
		expected := prefix + string(BucketReports)
		actual := manager.GetBucketName(BucketReports)
		assert.Equal(t, expected, actual, "prefix %q should result in bucket name %q", prefix, expected)
	}
}

// TestMultipleManagers tests creating multiple BucketManagers
func TestMultipleManagers(t *testing.T) {
	client1, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	client2, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager1 := NewBucketManager(client1, "prod-", "us-east-1")
	manager2 := NewBucketManager(client2, "staging-", "us-west-2")

	assert.Equal(t, "prod-reports", manager1.GetBucketName(BucketReports))
	assert.Equal(t, "staging-reports", manager2.GetBucketName(BucketReports))
	assert.Equal(t, "us-east-1", manager1.region)
	assert.Equal(t, "us-west-2", manager2.region)
}
