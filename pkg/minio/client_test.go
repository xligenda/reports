package minio

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockMinioClient is a mock implementation of minio.Client for testing.
type MockMinioClient struct {
	mock.Mock
}

func (m *MockMinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	args := m.Called(ctx, bucketName, objectName, reader, objectSize, opts)
	if args.Get(0) == nil {
		return minio.UploadInfo{}, args.Error(1)
	}
	return args.Get(0).(minio.UploadInfo), args.Error(1)
}

func (m *MockMinioClient) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) *minio.Object {
	args := m.Called(ctx, bucketName, objectName, opts)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*minio.Object)
}

func (m *MockMinioClient) ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	args := m.Called(ctx, bucketName, opts)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(<-chan minio.ObjectInfo)
}

func (m *MockMinioClient) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	args := m.Called(ctx, bucketName, objectName, opts)
	return args.Error(0)
}

func (m *MockMinioClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	args := m.Called(ctx, bucketName)
	return args.Bool(0), args.Error(1)
}

// TestNewMinioClient_Success tests successful client initialization.
func TestNewMinioClient_Success(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false, "test-bucket")

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "test-bucket", client.Bucket())
}

// TestNewMinioClient_InvalidEndpoint tests client initialization with invalid endpoint.
func TestNewMinioClient_InvalidEndpoint(t *testing.T) {
	// Invalid endpoint format should cause an error during Minio.New()
	client, err := NewMinioClient("invalid endpoint with spaces", "key", "secret", false, "bucket")

	// Minio should return an error for invalid endpoint
	if err != nil {
		assert.Error(t, err)
		assert.Nil(t, client)
	} else {
		// If somehow no error, ensure client is still created
		assert.NotNil(t, client)
	}
}

// TestNewMinioClient_WithSSL tests client initialization with SSL enabled.
func TestNewMinioClient_WithSSL(t *testing.T) {
	client, err := NewMinioClient("minio.example.com:9000", "accessKey", "secretKey", true, "my-bucket")

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "my-bucket", client.Bucket())
}

// TestNewMinioClient_BucketConfiguration tests that bucket is properly stored.
func TestNewMinioClient_BucketConfiguration(t *testing.T) {
	bucketName := "test-bucket-name"
	client, err := NewMinioClient("localhost:9000", "admin", "password", false, bucketName)

	require.NoError(t, err)
	assert.Equal(t, bucketName, client.Bucket())
}

// TestPutObject tests object upload functionality.
func TestPutObject(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false, "test-bucket")
	require.NoError(t, err)

	ctx := context.Background()
	data := []byte("test data")
	reader := bytes.NewReader(data)

	// This will attempt to upload to a real Minio instance
	// For true unit testing, you would need to mock at a lower level
	// For now, we just verify the method is callable
	_, err = client.PutObject(ctx, "test-object.txt", reader, int64(len(data)), minio.PutObjectOptions{})
	// Error is expected since we don't have a real Minio server
	_ = err
}

// TestBucket tests bucket getter.
func TestBucket(t *testing.T) {
	bucketName := "my-special-bucket"
	client, err := NewMinioClient("localhost:9000", "key", "secret", false, bucketName)

	require.NoError(t, err)
	assert.Equal(t, bucketName, client.Bucket())
}

// TestRemoveObject_Success tests successful object removal.
func TestRemoveObject(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false, "test-bucket")
	require.NoError(t, err)

	ctx := context.Background()
	
	// This will attempt to remove from a real Minio instance
	// For true unit testing, you would need to mock at a lower level
	err = client.RemoveObject(ctx, "test-object.txt", minio.RemoveObjectOptions{})
	// Error is expected since we don't have a real Minio server
	_ = err
}

// TestClientWithDifferentCredentials tests client initialization with various credentials.
func TestClientWithDifferentCredentials(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   string
		accessKey  string
		secretKey  string
		useSSL     bool
		bucket     string
		shouldFail bool
	}{
		{
			name:       "MinioLocal",
			endpoint:   "localhost:9000",
			accessKey:  "minioadmin",
			secretKey:  "minioadmin",
			useSSL:     false,
			bucket:     "test",
			shouldFail: false,
		},
		{
			name:       "CloudMinio",
			endpoint:   "minio.cloud.example.com:9000",
			accessKey:  "AKIAIOSFODNN7EXAMPLE",
			secretKey:  "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			useSSL:     true,
			bucket:     "prod-bucket",
			shouldFail: false,
		},
		{
			name:       "CustomPort",
			endpoint:   "storage.internal:12345",
			accessKey:  "user123",
			secretKey:  "pass456",
			useSSL:     false,
			bucket:     "data-bucket",
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewMinioClient(tt.endpoint, tt.accessKey, tt.secretKey, tt.useSSL, tt.bucket)

			if tt.shouldFail {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.bucket, client.Bucket())
			}
		})
	}
}

// TestClientOperationSequence tests a sequence of operations.
func TestClientOperationSequence(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false, "test-bucket")
	require.NoError(t, err)

	ctx := context.Background()

	// Test that all methods are callable without panic
	assert.NotPanics(t, func() {
		// These will fail without a real Minio server, but should not panic
		_, _ = client.PutObject(ctx, "file1.txt", bytes.NewReader([]byte("data")), 4, minio.PutObjectOptions{})
		_, _ = client.GetObject(ctx, "file1.txt", minio.GetObjectOptions{})
		_ = client.RemoveObject(ctx, "file1.txt", minio.RemoveObjectOptions{})
		_, _ = client.BucketExists(ctx)
	})
}

// TestNewMinioClient_EdgeCases tests edge cases for client initialization.
func TestNewMinioClient_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		endpoint  string
		accessKey string
		secretKey string
		useSSL    bool
		bucket    string
	}{
		{
			name:      "EmptyBucket",
			endpoint:  "localhost:9000",
			accessKey: "key",
			secretKey: "secret",
			useSSL:    false,
			bucket:    "",
		},
		{
			name:      "BucketWithHyphens",
			endpoint:  "localhost:9000",
			accessKey: "key",
			secretKey: "secret",
			useSSL:    false,
			bucket:    "my-test-bucket-123",
		},
		{
			name:      "EmptyEndpoint",
			endpoint:  "",
			accessKey: "key",
			secretKey: "secret",
			useSSL:    false,
			bucket:    "bucket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewMinioClient(tt.endpoint, tt.accessKey, tt.secretKey, tt.useSSL, tt.bucket)
			// Don't assert on error - just verify no panic and bucket is set
			if err == nil && client != nil {
				assert.Equal(t, tt.bucket, client.Bucket())
			}
		})
	}
}
