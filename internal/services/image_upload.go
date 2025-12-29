package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/develoFavour/e-voting-system-be/internal/config"
)

type ImageUploadService struct {
	cld *cloudinary.Cloudinary
}

func NewImageUploadService(cfg *config.Config) *ImageUploadService {
	cld, err := cloudinary.NewFromParams(
		cfg.CloudinaryCloudName,
		cfg.CloudinaryAPIKey,
		cfg.CloudinaryAPISecret,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Cloudinary: %v", err))
	}

	return &ImageUploadService{
		cld: cld,
	}
}

// UploadImage uploads an image to Cloudinary and returns the secure URL
func (s *ImageUploadService) UploadImage(file *multipart.FileHeader, folder string) (string, error) {
	if file == nil {
		return "", fmt.Errorf("file is nil")
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	// Validate file type (only allow images)
	if !isValidImageType(file.Header.Get("Content-Type")) {
		return "", fmt.Errorf("invalid file type: %s", file.Header.Get("Content-Type"))
	}

	// Generate a public ID for the image
	publicID := fmt.Sprintf("%s/%s", folder, generatePublicID(file.Filename))

	// Upload to Cloudinary using the file reader directly
	ctx := context.Background()
	uploadResult, err := s.cld.Upload.Upload(ctx, src, uploader.UploadParams{
		PublicID:       publicID,
		Folder:         folder,
		ResourceType:   "image",
		Transformation: "c_fill,w_300,h_300,q_auto,f_auto", // Auto-crop to 300x300 with optimization
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to Cloudinary: %v", err)
	}

	return uploadResult.SecureURL, nil
}

// isValidImageType checks if the content type is a valid image type
func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

// generatePublicID creates a clean public ID from filename
func generatePublicID(filename string) string {
	// Remove file extension and clean up the name
	name := strings.TrimSuffix(filename, getFileExtension(filename))

	// Replace spaces and special characters with underscores
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	// Remove any remaining non-alphanumeric characters except underscores
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// getFileExtension returns the file extension including the dot
func getFileExtension(filename string) string {
	if lastDot := strings.LastIndex(filename, "."); lastDot != -1 {
		return filename[lastDot:]
	}
	return ""
}
