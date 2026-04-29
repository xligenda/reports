package storage

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// if AccessKeyID is empty, the default AWS credential chain is used instead
// env vars → shared credentials file → IAM role
func New(ctx context.Context, cfg Config) (*S3Client, error) {
	var optFns []func(*config.LoadOptions) error

	optFns = append(optFns, config.WithRegion(cfg.Region))

	if cfg.AccessKeyID != "" {
		optFns = append(optFns, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, &Error{Op: "New", Err: err}
	}

	s3c := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.ForcePathStyle
		if cfg.Endpoint != "" {
			o.EndpointResolverV2 = &staticResolver{Endpoint: cfg.Endpoint}
		}
	})

	return &S3Client{
		s3:     s3c,
		psign:  s3.NewPresignClient(s3c),
		tm:     transfermanager.New(s3c),
		region: cfg.Region,
	}, nil
}

func NewFromAWSConfig(awsCfg aws.Config, forcePathStyle bool) *S3Client {
	s3c := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = forcePathStyle
	})

	return &S3Client{
		s3:     s3c,
		psign:  s3.NewPresignClient(s3c),
		tm:     transfermanager.New(s3c),
		region: awsCfg.Region,
	}
}
