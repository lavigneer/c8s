/*
Copyright 2025 C8S Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package s3 provides S3 implementation of the StorageClient interface
package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/org/c8s/pkg/storage"
)

// Client implements StorageClient interface using AWS S3
type Client struct {
	s3Client   *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucket     string
}

// NewClient creates a new S3 storage client
func NewClient(config *storage.Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	awsConfig := &aws.Config{
		Region:      aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
	}

	// For S3-compatible services like MinIO
	if config.Endpoint != "" {
		awsConfig.Endpoint = aws.String(config.Endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(config.UsePathStyle)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	s3Client := s3.New(sess)

	return &Client{
		s3Client:   s3Client,
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
		bucket:     config.Bucket,
	}, nil
}

// UploadLog uploads log content to S3
func (c *Client) UploadLog(ctx context.Context, key string, content io.Reader) error {
	_, err := c.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        content,
		ContentType: aws.String("text/plain"),
	})
	if err != nil {
		return fmt.Errorf("%w: %v", storage.ErrUploadFailed, err)
	}
	return nil
}

// DownloadLog downloads log content from S3
func (c *Client) DownloadLog(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := c.s3Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", storage.ErrDownloadFailed, err)
	}
	return result.Body, nil
}

// UploadArtifact uploads an artifact file to S3
func (c *Client) UploadArtifact(ctx context.Context, key string, content io.Reader) error {
	_, err := c.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        content,
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		return fmt.Errorf("%w: %v", storage.ErrUploadFailed, err)
	}
	return nil
}

// DownloadArtifact downloads an artifact file from S3
func (c *Client) DownloadArtifact(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := c.s3Client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", storage.ErrDownloadFailed, err)
	}
	return result.Body, nil
}

// GenerateSignedURL generates a pre-signed URL for downloading a file
func (c *Client) GenerateSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	req, _ := c.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(expiry)
	if err != nil {
		return "", fmt.Errorf("%w: %v", storage.ErrSignedURLFailed, err)
	}

	return url, nil
}

// ListObjects lists objects with the given prefix
func (c *Client) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	var keys []string

	err := c.s3Client.ListObjectsV2PagesWithContext(ctx,
		&s3.ListObjectsV2Input{
			Bucket: aws.String(c.bucket),
			Prefix: aws.String(prefix),
		},
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, obj := range page.Contents {
				keys = append(keys, *obj.Key)
			}
			return true
		},
	)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", storage.ErrListFailed, err)
	}

	return keys, nil
}

// DeleteObject deletes an object from S3
func (c *Client) DeleteObject(ctx context.Context, key string) error {
	_, err := c.s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("%w: %v", storage.ErrDeleteFailed, err)
	}
	return nil
}

// ObjectExists checks if an object exists in S3
func (c *Client) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := c.s3Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if it's a "not found" error
		if _, ok := err.(interface{ StatusCode() int }); ok {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
