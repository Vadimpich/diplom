package storage

import (
	"context"
	"io"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	UseSSL          bool
}

type S3Client struct {
	client *minio.Client
	bucket string
}

func NewS3(ctx context.Context, cfg Config) (*S3Client, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	secure := cfg.UseSSL
	if strings.Contains(endpoint, "://") {
		parsed, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}
		endpoint = parsed.Host
		secure = parsed.Scheme == "https"
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	return &S3Client{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func (c *S3Client) Upload(ctx context.Context, key, contentType string, size int64, body io.Reader) error {
	_, err := c.client.PutObject(ctx, c.bucket, key, body, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (c *S3Client) Delete(ctx context.Context, key string) error {
	return c.client.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{})
}
