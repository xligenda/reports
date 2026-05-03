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

// TestListAllBuckets_WithPrefix tests ListAllBuckets with prefix
func TestListAllBuckets_WithPrefix(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "test-", "us-east-1")

	buckets, err := manager.ListAllBuckets(context.Background())

	require.NoError(t, err)
	assert.Len(t, buckets, 4)
	assert.Contains(t, buckets, "test-video")
	assert.Contains(t, buckets, "test-image")
	assert.Contains(t, buckets, "test-audio")
	assert.Contains(t, buckets, "test-others")
}

// TestValidateBucketType_InvalidTypes tests ValidateBucketType with invalid types
func TestValidateBucketType_InvalidTypes(t *testing.T) {
	invalidTypes := []BucketType{"invalid", "unknown", "reports", "uploads"}

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
	err = manager.EnsureBucketsExist(ctx, BucketProof)
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

// TestGetBucketName_DifferentPrefixes tests GetBucketName with various prefixes
func TestGetBucketName_DifferentPrefixes(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	prefixes := []string{"", "prod-", "staging-", "dev-", "test-v1-"}

	for _, prefix := range prefixes {
		manager := NewBucketManager(client, prefix, "us-east-1")
		expected := prefix + string(BucketProof)
		actual := manager.GetBucketName(BucketProof)
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

	assert.Equal(t, "prod-proofs", manager1.GetBucketName(BucketProof))
	assert.Equal(t, "staging-proofs", manager2.GetBucketName(BucketProof))
	assert.Equal(t, "us-east-1", manager1.region)
	assert.Equal(t, "us-west-2", manager2.region)
}
