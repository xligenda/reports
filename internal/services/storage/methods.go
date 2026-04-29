package storage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
)

func (c *S3Client) Upload(ctx context.Context, bucket, key string, body io.Reader) error {
	_, err := c.tm.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil {
		return &Error{Op: "Upload", Err: err}
	}
	return nil
}
