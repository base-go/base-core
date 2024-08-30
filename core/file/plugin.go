package file

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

/*

Usage:
func SomeFunction(c *gin.Context) {
    fileHeader, err := c.FormFile("file")
    if err != nil {
        // Handle error
    }


    result, err := file.Upload(fileHeader, file.DefaultConfig)
    if err != nil {
        // Handle error
    }
    // Use result.Filename, result.Size, result.Path as needed

	customConfig := file.UploadConfig{
		AllowedExtensions: []string{".pdf", ".doc", ".docx"},
		MaxFileSize:       20 << 20, // 20 MB
		UploadPath:        "custom_uploads",
	}

	result, err := file.Upload(fileHeader, customConfig)

}
*/

// UploadConfig holds configuration for file uploads
type UploadConfig struct {
	AllowedExtensions []string
	MaxFileSize       int64 // in bytes
	UploadPath        string
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// DefaultConfig provides a default configuration for file uploads
var DefaultConfig = UploadConfig{
	AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"},
	MaxFileSize:       10 << 20, // 10 MB
	UploadPath:        "storage/uploads",
}

// UploadResult contains information about the uploaded file
type UploadResult struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Path     string `json:"path"`
}

// Upload handles file upload with the given configuration
func Upload(file *multipart.FileHeader, config UploadConfig) (*UploadResult, error) {
	// Validate file size
	if file.Size > config.MaxFileSize {
		return nil, fmt.Errorf("file size exceeds the limit of %d bytes", config.MaxFileSize)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !contains(config.AllowedExtensions, ext) {
		return nil, fmt.Errorf("file extension %s is not allowed", ext)
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(config.UploadPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate a unique filename
	filename := generateUniqueFilename(file.Filename)
	dst := filepath.Join(config.UploadPath, filename)

	// Open the source file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Create the destination file
	out, err := os.Create(dst)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	// Copy the uploaded file to the destination file
	_, err = io.Copy(out, src)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	return &UploadResult{
		Filename: filename,
		Size:     file.Size,
		Path:     dst,
	}, nil
}

// UploadHandler is a Gin handler for file uploads
// @Summary Upload a file
// @Description Upload a file
// @Security ApiKeyAuth
// @Tags Core/FileUpload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 200 {object} UploadResult
// @Failure 400 {object} ErrorResponse
// @Router /upload [post]
func UploadHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to get file from form"})
		return
	}

	result, err := Upload(file, DefaultConfig)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "File uploaded successfully",
		"file":    result,
	})
}

// SetupFileRoutes sets up the file-related routes
func SetupFileRoutes(router *gin.RouterGroup) {
	router.POST("/upload", UploadHandler)
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	name := strings.TrimSuffix(originalFilename, ext)
	return fmt.Sprintf("%s_%d%s", name, makeTimestamp(), ext)
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// InitFileModule initializes the file module
func InitFileModule(router *gin.RouterGroup) {
	log.Info("Initializing File module")
	SetupFileRoutes(router)
}
