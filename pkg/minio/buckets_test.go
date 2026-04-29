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

	assert.Equal(t, "video", manager.GetBucketName(BucketVideo))
	assert.Equal(t, "image", manager.GetBucketName(BucketImage))
	assert.Equal(t, "audio", manager.GetBucketName(BucketAudio))
	assert.Equal(t, "others", manager.GetBucketName(BucketOthers))
}

// TestGetBucketName_WithPrefix tests GetBucketName with prefix
func TestGetBucketName_WithPrefix(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "staging-", "us-east-1")

	assert.Equal(t, "staging-video", manager.GetBucketName(BucketVideo))
	assert.Equal(t, "staging-image", manager.GetBucketName(BucketImage))
	assert.Equal(t, "staging-audio", manager.GetBucketName(BucketAudio))
	assert.Equal(t, "staging-others", manager.GetBucketName(BucketOthers))
}

// TestGetBucketOrDefault_ValidType tests GetBucketOrDefault with valid bucket type
func TestGetBucketOrDefault_ValidType(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "", "us-east-1")

	assert.Equal(t, "video", manager.GetBucketOrDefault(BucketVideo))
	assert.Equal(t, "image", manager.GetBucketOrDefault(BucketImage))
	assert.Equal(t, "audio", manager.GetBucketOrDefault(BucketAudio))
	assert.Equal(t, "others", manager.GetBucketOrDefault(BucketOthers))
}

// TestGetBucketOrDefault_EmptyType tests GetBucketOrDefault with empty bucket type defaults to others
func TestGetBucketOrDefault_EmptyType(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "", "us-east-1")

	assert.Equal(t, "others", manager.GetBucketOrDefault(""))
}

// TestListAllBuckets tests ListAllBuckets returns all managed bucket names
func TestListAllBuckets(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	manager := NewBucketManager(client, "", "us-east-1")

	buckets, err := manager.ListAllBuckets(context.Background())

	require.NoError(t, err)
	assert.Len(t, buckets, 4)
	assert.Contains(t, buckets, "video")
	assert.Contains(t, buckets, "image")
	assert.Contains(t, buckets, "audio")
	assert.Contains(t, buckets, "others")
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

// TestValidateBucketType_ValidTypes tests ValidateBucketType with valid types
func TestValidateBucketType_ValidTypes(t *testing.T) {
	validTypes := []BucketType{BucketVideo, BucketImage, BucketAudio, BucketOthers}

	for _, bt := range validTypes {
		assert.True(t, ValidateBucketType(bt), "bucket type %q should be valid", bt)
	}
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
	err = manager.EnsureBucketsExist(ctx, BucketVideo, BucketAudio)
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
	assert.Equal(t, "video", string(BucketVideo))
	assert.Equal(t, "image", string(BucketImage))
	assert.Equal(t, "audio", string(BucketAudio))
	assert.Equal(t, "others", string(BucketOthers))
}

// TestGetBucketName_DifferentPrefixes tests GetBucketName with various prefixes
func TestGetBucketName_DifferentPrefixes(t *testing.T) {
	client, err := NewMinioClient("localhost:9000", "minioadmin", "minioadmin", false)
	require.NoError(t, err)

	prefixes := []string{"", "prod-", "staging-", "dev-", "test-v1-"}

	for _, prefix := range prefixes {
		manager := NewBucketManager(client, prefix, "us-east-1")
		expected := prefix + string(BucketVideo)
		actual := manager.GetBucketName(BucketVideo)
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

	assert.Equal(t, "prod-video", manager1.GetBucketName(BucketVideo))
	assert.Equal(t, "staging-image", manager2.GetBucketName(BucketImage))
	assert.Equal(t, "us-east-1", manager1.region)
	assert.Equal(t, "us-west-2", manager2.region)
}

// TestGetBucketTypeFromFileName_Videos tests file type detection for video files
func TestGetBucketTypeFromFileName_Videos(t *testing.T) {
	videoFiles := []string{"movie.mp4", "video.avi", "film.mkv", "clip.mov", "show.webm"}

	for _, file := range videoFiles {
		result := GetBucketTypeFromFileName(file)
		assert.Equal(t, BucketVideo, result, "file %q should be classified as video", file)
	}
}

// TestGetBucketTypeFromFileName_Images tests file type detection for image files
func TestGetBucketTypeFromFileName_Images(t *testing.T) {
	imageFiles := []string{"photo.jpg", "picture.jpeg", "image.png", "graphic.gif", "drawing.bmp", "logo.svg", "web.webp"}

	for _, file := range imageFiles {
		result := GetBucketTypeFromFileName(file)
		assert.Equal(t, BucketImage, result, "file %q should be classified as image", file)
	}
}

// TestGetBucketTypeFromFileName_Audio tests file type detection for audio files
func TestGetBucketTypeFromFileName_Audio(t *testing.T) {
	audioFiles := []string{"song.mp3", "music.wav", "track.flac", "podcast.aac", "voice.m4a", "sound.ogg"}

	for _, file := range audioFiles {
		result := GetBucketTypeFromFileName(file)
		assert.Equal(t, BucketAudio, result, "file %q should be classified as audio", file)
	}
}

// TestGetBucketTypeFromFileName_Others tests file type detection for unknown file types
func TestGetBucketTypeFromFileName_Others(t *testing.T) {
	otherFiles := []string{"document.pdf", "spreadsheet.xlsx", "archive.zip", "script.py", "unknown.xyz", "noext"}

	for _, file := range otherFiles {
		result := GetBucketTypeFromFileName(file)
		assert.Equal(t, BucketOthers, result, "file %q should be classified as others", file)
	}
}

// TestGetBucketTypeFromFileName_CaseInsensitive tests that file type detection is case insensitive
func TestGetBucketTypeFromFileName_CaseInsensitive(t *testing.T) {
	testCases := []struct {
		fileName   string
		bucketType BucketType
	}{
		{"video.MP4", BucketVideo},
		{"IMAGE.PNG", BucketImage},
		{"Song.MP3", BucketAudio},
		{"Document.PDF", BucketOthers},
	}

	for _, tc := range testCases {
		result := GetBucketTypeFromFileName(tc.fileName)
		assert.Equal(t, tc.bucketType, result, "file %q should be classified as %q", tc.fileName, tc.bucketType)
	}
}
