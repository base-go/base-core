package storage

import (
	"fmt"
	"mime/multipart"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// ContaboConfig holds configuration for Contabo storage
type ContaboConfig struct {
	APIKey    string
	APISecret string
	Endpoint  string
	Bucket    string
	BaseURL   string
}

type contaboProvider struct {
	client   *s3.S3
	bucket   string
	endpoint string
	baseURL  string
}

func NewContaboProvider(config ContaboConfig) (Provider, error) {
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = "eu2.contabostorage.com"
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.APIKey, config.APISecret, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("eu-central-1"),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &contaboProvider{
		client:   s3.New(sess),
		bucket:   config.Bucket,
		endpoint: endpoint,
		baseURL:  config.BaseURL,
	}, nil
}

func (p *contaboProvider) Upload(file *multipart.FileHeader, config UploadConfig) (*UploadResult, error) {
	// Open source file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Generate unique filename
	filename := generateUniqueFilename(file.Filename)
	key := fmt.Sprintf("%s/%s", config.UploadPath, filename)

	// Upload to Contabo
	_, err = p.client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
		Body:   src,
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to Contabo: %w", err)
	}

	return &UploadResult{
		Filename: filename,
		Path:     key,
		Size:     file.Size,
	}, nil
}

func (p *contaboProvider) Delete(path string) error {
	_, err := p.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(path),
	})
	return err
}

func (p *contaboProvider) GetURL(path string) string {
	return fmt.Sprintf("https://%s/%s/%s", p.endpoint, p.bucket, path)
}
