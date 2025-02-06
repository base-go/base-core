package storage

import (
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// R2Config holds configuration for Cloudflare R2 storage
type R2Config struct {
	AccessKeyID     string
	AccessKeySecret string
	AccountID       string
	Bucket          string
	BaseURL         string
	CDN             string
}

type r2Provider struct {
	client   *s3.S3
	bucket   string
	endpoint string
	baseURL  string
	cdn      string
}

func NewR2Provider(config R2Config) (Provider, error) {
	// R2 endpoint format: https://<account_id>.r2.cloudflarestorage.com
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", config.AccountID)

	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.AccessKeyID, config.AccessKeySecret, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("auto"), // R2 uses 'auto' as region
		S3ForcePathStyle: aws.Bool(false),    // R2 requires this to be false
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create R2 session: %w", err)
	}

	// Create S3 client with the correct API version
	s3Config := &aws.Config{
		S3ForcePathStyle: aws.Bool(false),
	}
	client := s3.New(sess, s3Config)

	return &r2Provider{
		client:   client,
		bucket:   config.Bucket,
		endpoint: endpoint,
		baseURL:  config.BaseURL,
		cdn:      config.CDN,
	}, nil
}

func (p *r2Provider) Upload(file *multipart.FileHeader, config UploadConfig) (*UploadResult, error) {
	// Open source file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Generate unique filename
	filename := generateUniqueFilename(file.Filename)
	key := fmt.Sprintf("%s/%s", config.UploadPath, filename)

	// Upload to R2
	_, err = p.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(key),
		Body:        src,
		ContentType: aws.String(file.Header.Get("Content-Type")),
		// Note: R2 doesn't support ACL, so we remove the ACL setting
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to R2: %w", err)
	}

	return &UploadResult{
		Filename: filename,
		Path:     key,
		Size:     file.Size,
	}, nil
}

func (p *r2Provider) Delete(path string) error {
	_, err := p.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(path),
	})
	return err
}

func (p *r2Provider) GetURL(path string) string {
	// Always prefer CDN for R2 storage
	if p.cdn != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(p.cdn, "/"), path)
	}
	// Fallback to BaseURL if CDN is not configured
	if p.baseURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(p.baseURL, "/"), path)
	}
	// Last resort: use R2 URL
	return fmt.Sprintf("https://%s/%s/%s", p.endpoint, p.bucket, path)
}
