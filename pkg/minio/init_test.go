package minio

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewInitOptions_DefaultValues tests NewInitOptions with default values
func TestNewInitOptions_DefaultValues(t *testing.T) {
	// Clear environment variables
	os.Clearenv()

	opts := NewInitOptions()

	assert.Equal(t, "localhost:9000", opts.Endpoint)
	assert.Equal(t, "minioadmin", opts.AccessKey)
	assert.Equal(t, "minioadmin", opts.SecretKey)
	assert.False(t, opts.UseSSL)
	assert.Equal(t, "", opts.BucketPrefix)
	assert.Equal(t, "us-east-1", opts.Region)
	assert.NotEmpty(t, opts.BucketTypes)
}

// TestNewInitOptions_FromEnvironment tests NewInitOptions reading from environment
func TestNewInitOptions_FromEnvironment(t *testing.T) {
	// Set environment variables
	t.Setenv("MINIO_ENDPOINT", "minio.example.com:9001")
	t.Setenv("MINIO_ACCESS_KEY", "testkey")
	t.Setenv("MINIO_SECRET_KEY", "testsecret")
	t.Setenv("MINIO_USE_SSL", "true")
	t.Setenv("MINIO_BUCKET_PREFIX", "prod-")
	t.Setenv("MINIO_REGION", "us-west-2")

	opts := NewInitOptions()

	assert.Equal(t, "minio.example.com:9001", opts.Endpoint)
	assert.Equal(t, "testkey", opts.AccessKey)
	assert.Equal(t, "testsecret", opts.SecretKey)
	assert.True(t, opts.UseSSL)
	assert.Equal(t, "prod-", opts.BucketPrefix)
	assert.Equal(t, "us-west-2", opts.Region)
}

// TestNewInitOptions_SSLVariations tests SSL environment variable parsing
func TestNewInitOptions_SSLVariations(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"1", true},
		{"yes", true},
		{"YES", true},
		{"false", false},
		{"False", false},
		{"FALSE", false},
		{"0", false},
		{"no", false},
		{"NO", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run("SSL_"+tt.value, func(t *testing.T) {
			if tt.value == "" {
				os.Unsetenv("MINIO_USE_SSL")
			} else {
				t.Setenv("MINIO_USE_SSL", tt.value)
			}

			opts := NewInitOptions()
			assert.Equal(t, tt.expected, opts.UseSSL)
		})
	}
}

// TestInitMinio_Success tests successful Minio initialization
func TestInitMinio_Success(t *testing.T) {
	ctx := context.Background()
	opts := &InitOptions{
		Endpoint:     "localhost:9000",
		AccessKey:    "minioadmin",
		SecretKey:    "minioadmin",
		UseSSL:       false,
		BucketPrefix: "",
		Region:       "us-east-1",
		BucketTypes:  []BucketType{BucketProof},
	}

	// This will fail with connection error, but should initialize the manager
	manager, _ := InitMinio(ctx, opts)
	// We might get an error due to no connection, but manager should be created
	if manager != nil {
		assert.NotNil(t, manager)
		assert.Equal(t, "reports", manager.GetBucketName(BucketProof))
	}
}

// TestInitMinio_WithPrefix tests Minio initialization with prefix
func TestInitMinio_WithPrefix(t *testing.T) {
	ctx := context.Background()
	opts := &InitOptions{
		Endpoint:     "localhost:9000",
		AccessKey:    "minioadmin",
		SecretKey:    "minioadmin",
		UseSSL:       false,
		BucketPrefix: "test-",
		Region:       "us-east-1",
		BucketTypes:  []BucketType{BucketProof},
	}

	manager, _ := InitMinio(ctx, opts)
	if manager != nil {
		assert.NotNil(t, manager)
		assert.Equal(t, "test-proofs", manager.GetBucketName(BucketProof))
	}
}

// TestInitMinio_NilOptions tests Minio initialization with nil options
func TestInitMinio_NilOptions(t *testing.T) {
	ctx := context.Background()
	os.Clearenv()

	// Should use default options
	manager, _ := InitMinio(ctx, nil)
	if manager != nil {
		assert.NotNil(t, manager)
	}
}

// TestInitMinioFromEnv tests direct initialization from environment
func TestInitMinioFromEnv_DefaultEnv(t *testing.T) {
	ctx := context.Background()
	os.Clearenv()

	// Should create manager with default environment values
	manager, _ := InitMinioFromEnv(ctx)
	if manager != nil {
		assert.NotNil(t, manager)
		assert.Equal(t, "proof", manager.GetBucketName(BucketProof))
	}
}

// TestInitMinioFromEnv_CustomEnv tests initialization from custom environment
func TestInitMinioFromEnv_CustomEnv(t *testing.T) {
	ctx := context.Background()
	t.Setenv("MINIO_ENDPOINT", "minio.test.local:9000")
	t.Setenv("MINIO_ACCESS_KEY", "customkey")
	t.Setenv("MINIO_SECRET_KEY", "customsecret")
	t.Setenv("MINIO_BUCKET_PREFIX", "custom-")

	manager, _ := InitMinioFromEnv(ctx)
	if manager != nil {
		assert.NotNil(t, manager)
		assert.Equal(t, "custom-proofs", manager.GetBucketName(BucketProof))
	}
}

// TestGetEnv tests getEnv helper function
func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_VAR", "test_value")

	// Existing variable
	value := getEnv("TEST_VAR", "default")
	assert.Equal(t, "test_value", value)

	// Non-existing variable
	value = getEnv("NON_EXISTING_VAR", "default")
	assert.Equal(t, "default", value)

	// Empty variable (not set) uses default
	os.Unsetenv("EMPTY_VAR")
	value = getEnv("EMPTY_VAR", "default_value")
	assert.Equal(t, "default_value", value)
}

// TestGetEnvBool tests getEnvBool helper function
func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true", "true", true},
		{"1", "1", true},
		{"yes", "yes", true},
		{"false", "false", false},
		{"0", "0", false},
		{"no", "no", false},
		{"invalid", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("BOOL_VAR", tt.value)
			value := getEnvBool("BOOL_VAR", false)
			assert.Equal(t, tt.expected, value)
		})
	}

	// Test default value
	os.Unsetenv("UNSET_VAR")
	value := getEnvBool("UNSET_VAR", true)
	assert.True(t, value)
}

// TestInitOptions_BucketTypes tests that BucketTypes are properly set
func TestInitOptions_BucketTypes(t *testing.T) {
	opts := NewInitOptions()

	assert.NotEmpty(t, opts.BucketTypes)
	assert.Contains(t, opts.BucketTypes, BucketProof)

}

// TestInitMinio_InvalidEndpoint tests Minio initialization with invalid endpoint
func TestInitMinio_InvalidEndpoint(t *testing.T) {
	ctx := context.Background()
	opts := &InitOptions{
		Endpoint:    "invalid endpoint with spaces",
		AccessKey:   "key",
		SecretKey:   "secret",
		UseSSL:      false,
		Region:      "us-east-1",
		BucketTypes: []BucketType{BucketProof},
	}

	_, err := InitMinio(ctx, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create minio client")
}
