package minio

import (
	"context"
	"fmt"
	"os"
)

// InitOptions holds configuration options for initializing Minio
type InitOptions struct {
	Endpoint     string
	AccessKey    string
	SecretKey    string
	UseSSL       bool
	BucketPrefix string
	Region       string
	BucketTypes  []BucketType
}

// NewInitOptions creates InitOptions from environment variables
// Supported environment variables:
// - MINIO_ENDPOINT: Minio endpoint address (default: localhost:9000)
// - MINIO_ACCESS_KEY: Access key (default: minioadmin)
// - MINIO_SECRET_KEY: Secret key (default: minioadmin)
// - MINIO_USE_SSL: Whether to use SSL (default: false)
// - MINIO_BUCKET_PREFIX: Prefix for bucket names (default: empty)
// - MINIO_REGION: AWS region (default: us-east-1)
func NewInitOptions() *InitOptions {
	return &InitOptions{
		Endpoint:     getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey:    getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey:    getEnv("MINIO_SECRET_KEY", "minioadmin"),
		UseSSL:       getEnvBool("MINIO_USE_SSL", false),
		BucketPrefix: getEnv("MINIO_BUCKET_PREFIX", ""),
		Region:       getEnv("MINIO_REGION", "us-east-1"),
		BucketTypes:  []BucketType{BucketProof},
	}
}

// InitMinio initializes Minio client and ensures buckets exist
// This function should be called during application startup
func InitMinio(ctx context.Context, opts *InitOptions) (*BucketManager, error) {
	if opts == nil {
		opts = NewInitOptions()
	}

	// Create client
	client, err := NewMinioClient(opts.Endpoint, opts.AccessKey, opts.SecretKey, opts.UseSSL)
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Create bucket manager
	manager := NewBucketManager(client, opts.BucketPrefix, opts.Region)

	// Ensure buckets exist
	if err := manager.EnsureBucketsExist(ctx, opts.BucketTypes...); err != nil {
		return nil, fmt.Errorf("failed to ensure buckets exist: %w", err)
	}

	return manager, nil
}

// InitMinioFromEnv initializes Minio directly from environment variables
func InitMinioFromEnv(ctx context.Context) (*BucketManager, error) {
	opts := NewInitOptions()
	return InitMinio(ctx, opts)
}

// getEnv retrieves an environment variable or returns the default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool retrieves a boolean environment variable or returns the default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	// Parse as boolean
	switch value {
	case "true", "True", "TRUE", "1", "yes", "YES":
		return true
	case "false", "False", "FALSE", "0", "no", "NO":
		return false
	default:
		return defaultValue
	}
}
