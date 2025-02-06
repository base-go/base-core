package storage

import (
	"fmt"
	"mime/multipart"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Config holds configuration for S3 storage
type S3Config struct {
	APIKey          string
	APISecret       string
	AccessKeyID     string
	AccessKeySecret string
	AccountID       string
	Endpoint        string
	Bucket          string
	BaseURL         string
	Region          string
}

type s3Provider struct {
	client   *s3.S3
	bucket   string
	endpoint string
	baseURL  string
}

func NewS3Provider(config S3Config) (Provider, error) {
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = "s3.amazonaws.com"
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.AccessKeyID, config.AccessKeySecret, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(config.Region),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &s3Provider{
		client:   s3.New(sess),
		bucket:   config.Bucket,
		endpoint: endpoint,
		baseURL:  config.BaseURL,
	}, nil
}

func (p *s3Provider) Upload(file *multipart.FileHeader, config UploadConfig) (*UploadResult, error) {
	// Open source file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Generate unique filename
	filename := generateUniqueFilename(file.Filename)
	key := fmt.Sprintf("%s/%s", config.UploadPath, filename)

	// Upload to S3
	_, err = p.client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
		Body:   src,
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	return &UploadResult{
		Filename: filename,
		Path:     key,
		Size:     file.Size,
	}, nil
}

func (p *s3Provider) Delete(path string) error {
	_, err := p.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(path),
	})
	return err
}

func (p *s3Provider) GetURL(path string) string {
	return fmt.Sprintf("https://%s/%s/%s", p.endpoint, p.bucket, path)
}
