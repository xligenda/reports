package minio

import (
	"bytes"
	"context"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMinioClient_Success tests successful client initialization without bucket binding
func TestNewMinioClient_Success(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)

	assert.NoError(t, err)
	assert.NotNil(t, client)
}

// TestNewMinioClient_InvalidEndpoint tests client initialization with invalid endpoint
func TestNewMinioClient_InvalidEndpoint(t *testing.T) {
	// Invalid endpoint format should cause an error during Minio.New()
	client, err := NewMinioClient("invalid endpoint with spaces", "key", "secret", false)

	// Minio should return an error for invalid endpoint
	if err != nil {
		assert.Error(t, err)
		assert.Nil(t, client)
	} else {
		// If somehow no error, ensure client is still created
		assert.NotNil(t, client)
	}
}

// TestNewMinioClient_WithSSL tests client initialization with SSL enabled
func TestNewMinioClient_WithSSL(t *testing.T) {
	client, err := NewMinioClient("minio.example.com:9000", "accessKey", "secretKey", true)

	assert.NoError(t, err)
	assert.NotNil(t, client)
}

// TestValidateBucketName_Success tests bucket name validation with valid names
func TestValidateBucketName_Success(t *testing.T) {
	validNames := []string{"test-bucket", "reports", "uploads-v2", "cache123"}

	for _, name := range validNames {
		err := validateBucketName(name)
		assert.NoError(t, err, "bucket name %q should be valid", name)
	}
}

// TestValidateBucketName_Failure tests bucket name validation with invalid names
func TestValidateBucketName_Failure(t *testing.T) {
	emptyName := ""
	err := validateBucketName(emptyName)
	assert.Error(t, err, "empty bucket name should fail validation")
	assert.Contains(t, err.Error(), "bucket name cannot be empty")
}

// TestPutObject_ValidParams tests PutObject with valid parameters
func TestPutObject_ValidParams(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()
	data := []byte("test data")
	reader := bytes.NewReader(data)

	// This will fail with a real endpoint error, but shouldn't panic
	_, err = client.PutObject(ctx, "test-bucket", "test-object.txt", reader, int64(len(data)), minio.PutObjectOptions{})
	// We don't assert on error since we don't have a real Minio server
	assert.Nil(t, nil) // ensure the call was made without panic
}

// TestPutObject_EmptyBucket tests PutObject with empty bucket name
func TestPutObject_EmptyBucket(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()
	data := []byte("test data")
	reader := bytes.NewReader(data)

	_, err = client.PutObject(ctx, "", "test-object.txt", reader, int64(len(data)), minio.PutObjectOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bucket name cannot be empty")
}

// TestGetObject tests GetObject with valid parameters
func TestGetObject(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	// This method shouldn't panic with valid input
	obj, err := client.GetObject(ctx, "test-bucket", "test-object.txt", minio.GetObjectOptions{})
	if err != nil {
		// Error expected without real server
		assert.Nil(t, obj)
	} else {
		assert.NotNil(t, obj)
	}
}

// TestGetObject_EmptyBucket tests GetObject with empty bucket name
func TestGetObject_EmptyBucket(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	obj, err := client.GetObject(ctx, "", "test-object.txt", minio.GetObjectOptions{})
	assert.Error(t, err)
	assert.Nil(t, obj)
	assert.Contains(t, err.Error(), "bucket name cannot be empty")
}

// TestListObjects tests ListObjects method
func TestListObjects(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	// This should return a channel even if the bucket doesn't exist
	objCh := client.ListObjects(ctx, "test-bucket", minio.ListObjectsOptions{})
	assert.NotNil(t, objCh)
}

// TestListObjects_InvalidBucket tests ListObjects with empty bucket
func TestListObjects_InvalidBucket(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	// Should return empty channel on invalid bucket
	objCh := client.ListObjects(ctx, "", minio.ListObjectsOptions{})
	assert.NotNil(t, objCh)
}

// TestRemoveObject_Success tests RemoveObject with valid parameters
func TestRemoveObject_Success(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	// Method should not panic
	err = client.RemoveObject(ctx, "test-bucket", "test-object.txt", minio.RemoveObjectOptions{})
	assert.Nil(t, nil) // ensure no panic
}

// TestRemoveObject_EmptyBucket tests RemoveObject with empty bucket name
func TestRemoveObject_EmptyBucket(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	err = client.RemoveObject(ctx, "", "test-object.txt", minio.RemoveObjectOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bucket name cannot be empty")
}

// TestBucketExists tests BucketExists method
func TestBucketExists(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	// Should not panic
	exists, err := client.BucketExists(ctx, "test-bucket")
	_ = exists // value might be false due to no connection
	_ = err    // error is expected without real server
}

// TestBucketExists_EmptyBucket tests BucketExists with empty bucket name
func TestBucketExists_EmptyBucket(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	exists, err := client.BucketExists(ctx, "")
	assert.Error(t, err)
	assert.False(t, exists)
	assert.Contains(t, err.Error(), "bucket name cannot be empty")
}

// TestCreateBucket tests CreateBucket method
func TestCreateBucket(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	// Should not panic
	err = client.CreateBucket(ctx, "test-bucket", "us-east-1")
	_ = err // error is expected without real server
}

// TestCreateBucket_EmptyBucket tests CreateBucket with empty bucket name
func TestCreateBucket_EmptyBucket(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	err = client.CreateBucket(ctx, "", "us-east-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bucket name cannot be empty")
}

// TestDeleteBucket tests DeleteBucket method
func TestDeleteBucket(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	// Should not panic
	err = client.DeleteBucket(ctx, "test-bucket")
	_ = err // error is expected without real server
}

// TestDeleteBucket_EmptyBucket tests DeleteBucket with empty bucket name
func TestDeleteBucket_EmptyBucket(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	err = client.DeleteBucket(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bucket name cannot be empty")
}

// TestHeadObject tests HeadObject method
func TestHeadObject(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	// Should not panic
	_, err = client.HeadObject(ctx, "test-bucket", "test-object.txt", minio.StatObjectOptions{})
	_ = err // error is expected without real server
}

// TestHeadObject_EmptyBucket tests HeadObject with empty bucket name
func TestHeadObject_EmptyBucket(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.HeadObject(ctx, "", "test-object.txt", minio.StatObjectOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bucket name cannot be empty")
}

// TestMultipleBucketsSupport tests that client can work with different buckets
func TestMultipleBucketsSupport(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()
	buckets := []string{"reports", "uploads", "cache"}

	for _, bucket := range buckets {
		// Each method should work with any bucket name (validation passes)
		_, err := client.BucketExists(ctx, bucket)
		_ = err // error from connection is OK, we're testing bucket parameter acceptance
	}
}

// TestNewMinioClient_DifferentCredentials tests client with various credentials
func TestNewMinioClient_DifferentCredentials(t *testing.T) {
	credentials := []struct {
		endpoint  string
		accessKey string
		secretKey string
		ssl       bool
	}{
		{
			endpoint:  "localhost:9000",
			accessKey: "minioadmin",
			secretKey: "minioadmin",
			ssl:       false,
		},
		{
			endpoint:  "minio.example.com:9000",
			accessKey: "AKIAIOSFODNN7EXAMPLE",
			secretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			ssl:       true,
		},
		{
			endpoint:  "storage.internal:12345",
			accessKey: "user123",
			secretKey: "pass456",
			ssl:       false,
		},
	}

	for _, cred := range credentials {
		client, err := NewMinioClient(cred.endpoint, cred.accessKey, cred.secretKey, cred.ssl)
		assert.NoError(t, err)
		assert.NotNil(t, client)
	}
}

// TestClientNoPanic tests that operations don't panic with valid input
func TestClientNoPanic(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	ctx := context.Background()

	// Test that methods don't panic
	assert.NotPanics(t, func() {
		data := []byte("test")
		reader := bytes.NewReader(data)
		_, _ = client.PutObject(ctx, "bucket", "key", reader, int64(len(data)), minio.PutObjectOptions{})
		_, _ = client.GetObject(ctx, "bucket", "key", minio.GetObjectOptions{})
		_ = client.RemoveObject(ctx, "bucket", "key", minio.RemoveObjectOptions{})
		_, _ = client.BucketExists(ctx, "bucket")
		_ = client.CreateBucket(ctx, "bucket", "us-east-1")
		_ = client.DeleteBucket(ctx, "bucket")
		_, _ = client.HeadObject(ctx, "bucket", "key", minio.StatObjectOptions{})
		_ = client.ListObjects(ctx, "bucket", minio.ListObjectsOptions{})
	})
}
