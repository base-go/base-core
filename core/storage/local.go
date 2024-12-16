package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// LocalConfig holds configuration for local storage
type LocalConfig struct {
	BasePath string
	BaseURL  string
}

type localProvider struct {
	basePath string
	baseURL  string
}

func NewLocalProvider(config LocalConfig) (Provider, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(config.BasePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &localProvider{
		basePath: config.BasePath,
		baseURL:  config.BaseURL,
	}, nil
}

func (p *localProvider) Upload(file *multipart.FileHeader, config UploadConfig) (*UploadResult, error) {
	// Create upload directory
	uploadPath := filepath.Join(p.basePath, config.UploadPath)
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	filename := generateUniqueFilename(file.Filename)
	dst := filepath.Join(uploadPath, filename)

	// Open source file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Create destination file
	out, err := os.Create(dst)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	// Copy file
	if _, err = io.Copy(out, src); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	relativePath := filepath.Join(config.UploadPath, filename)

	return &UploadResult{
		Filename: filename,
		Path:     relativePath,
		Size:     file.Size,
	}, nil
}

func (p *localProvider) Delete(path string) error {
	fullPath := filepath.Join(p.basePath, path)
	return os.Remove(fullPath)
}

func (p *localProvider) GetURL(path string) string {
	return fmt.Sprintf("%s/%s", p.baseURL, path)
}
