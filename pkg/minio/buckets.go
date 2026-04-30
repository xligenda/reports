package minio

import (
	"context"
	"fmt"
	"strings"
)

// BucketType represents the type of storage bucket based on file type
type BucketType string

const (
	// BucketVideo stores video files
	BucketVideo BucketType = "video"
	// BucketImage stores image files
	BucketImage BucketType = "image"
	// BucketAudio stores audio files
	BucketAudio BucketType = "audio"
	// BucketOthers stores other file types
	BucketOthers BucketType = "others"
)

// FileTypeMap defines file extensions to bucket type mappings
var FileTypeMap = map[string]BucketType{
	// Video formats
	".mp4":  BucketVideo,
	".avi":  BucketVideo,
	".mkv":  BucketVideo,
	".mov":  BucketVideo,
	".flv":  BucketVideo,
	".wmv":  BucketVideo,
	".webm": BucketVideo,
	".3gp":  BucketVideo,
	".m4v":  BucketVideo,

	// Image formats
	".jpg":  BucketImage,
	".jpeg": BucketImage,
	".png":  BucketImage,
	".gif":  BucketImage,
	".bmp":  BucketImage,
	".svg":  BucketImage,
	".webp": BucketImage,
	".tiff": BucketImage,
	".ico":  BucketImage,

	// Audio formats
	".mp3":  BucketAudio,
	".wav":  BucketAudio,
	".flac": BucketAudio,
	".aac":  BucketAudio,
	".m4a":  BucketAudio,
	".aiff": BucketAudio,
	".ogg":  BucketAudio,
	".wma":  BucketAudio,
}

func BucketTypeFromFileName(fileName string) BucketType {
	ext := strings.ToLower(strings.TrimSpace(fileName))

	// Find the extension
	dotIndex := strings.LastIndex(ext, ".")
	if dotIndex == -1 {
		return BucketOthers
	}

	fileExt := ext[dotIndex:]

	if bucketType, exists := FileTypeMap[fileExt]; exists {
		return bucketType
	}

	return BucketOthers
}

// BucketManager manages multiple buckets with type-based organization
type BucketManager struct {
	client *Client
	prefix string
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
		bucketTypes = []BucketType{BucketVideo, BucketImage, BucketAudio, BucketOthers}
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
		return m.GetBucketName(BucketOthers) // default
	}
	return m.GetBucketName(bucketType)
}

// ListAllBuckets lists all managed buckets (returns only our managed buckets)
func (m *BucketManager) ListAllBuckets(ctx context.Context) ([]string, error) {
	buckets := []string{
		m.GetBucketName(BucketVideo),
		m.GetBucketName(BucketImage),
		m.GetBucketName(BucketAudio),
		m.GetBucketName(BucketOthers),
	}
	return buckets, nil
}

// ValidateBucketType checks if a bucket type is valid
func ValidateBucketType(bt BucketType) bool {
	switch bt {
	case BucketVideo, BucketImage, BucketAudio, BucketOthers:
		return true
	default:
		return false
	}
}
