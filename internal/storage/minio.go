package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"dixitme/internal/logger"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient wraps MinIO client with our custom methods
type MinIOClient struct {
	client     *minio.Client
	bucketName string
}

// Config holds MinIO configuration
type MinIOConfig struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	BucketName      string `json:"bucket_name"`
	UseSSL          bool   `json:"use_ssl"`
	Region          string `json:"region"`
}

var minioClient *MinIOClient

// Initialize sets up MinIO client
func Initialize(cfg MinIOConfig) error {
	log := logger.GetLogger()

	// Initialize MinIO client
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return fmt.Errorf("failed to create MinIO client: %w", err)
	}

	minioClient = &MinIOClient{
		client:     client,
		bucketName: cfg.BucketName,
	}

	// Check if bucket exists, create if not
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{
			Region: cfg.Region,
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Info("MinIO bucket created", "bucket", cfg.BucketName)
	}

	// Set bucket policy to public read for card images
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"AWS": "*"},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/cards/*"]
			}
		]
	}`, cfg.BucketName)

	err = client.SetBucketPolicy(ctx, cfg.BucketName, policy)
	if err != nil {
		log.Warn("Failed to set bucket policy", "error", err)
	}

	log.Info("MinIO client initialized", "endpoint", cfg.Endpoint, "bucket", cfg.BucketName)
	return nil
}

// GetClient returns the MinIO client instance
func GetClient() *MinIOClient {
	return minioClient
}

// UploadCardImage uploads a card image to MinIO
func (mc *MinIOClient) UploadCardImage(cardID int, file multipart.File, header *multipart.FileHeader) (string, error) {
	ctx := context.Background()
	log := logger.GetLogger()

	// Generate object name
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg" // Default extension
	}
	objectName := fmt.Sprintf("cards/%d%s", cardID, ext)

	// Get file size
	fileSize := header.Size

	// Set content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	// Upload file
	info, err := mc.client.PutObject(ctx, mc.bucketName, objectName, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"card-id":     fmt.Sprintf("%d", cardID),
			"uploaded-at": time.Now().Format(time.RFC3339),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate public URL
	url := mc.GetCardImageURL(cardID, ext)

	log.Info("Card image uploaded",
		"card_id", cardID,
		"object_name", objectName,
		"size", info.Size,
		"url", url)

	return url, nil
}

// GetCardImageURL returns the public URL for a card image
func (mc *MinIOClient) GetCardImageURL(cardID int, extension string) string {
	if extension == "" {
		extension = ".jpg"
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	objectName := fmt.Sprintf("cards/%d%s", cardID, extension)

	// For public buckets, we can construct the URL directly
	protocol := "http"
	// Note: We'll assume http for simplicity. In production, check the endpoint or use presigned URLs

	return fmt.Sprintf("%s://%s/%s/%s", protocol, mc.client.EndpointURL().Host, mc.bucketName, objectName)
}

// DeleteCardImage removes a card image from MinIO
func (mc *MinIOClient) DeleteCardImage(cardID int, extension string) error {
	ctx := context.Background()
	log := logger.GetLogger()

	if extension == "" {
		extension = ".jpg"
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	objectName := fmt.Sprintf("cards/%d%s", cardID, extension)

	err := mc.client.RemoveObject(ctx, mc.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	log.Info("Card image deleted", "card_id", cardID, "object_name", objectName)
	return nil
}

// GetCardImage downloads a card image from MinIO
func (mc *MinIOClient) GetCardImage(cardID int, extension string) (io.ReadCloser, error) {
	ctx := context.Background()

	if extension == "" {
		extension = ".jpg"
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	objectName := fmt.Sprintf("cards/%d%s", cardID, extension)

	object, err := mc.client.GetObject(ctx, mc.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return object, nil
}

// ListCardImages lists all card images in the bucket
func (mc *MinIOClient) ListCardImages() ([]string, error) {
	ctx := context.Background()

	objectCh := mc.client.ListObjects(ctx, mc.bucketName, minio.ListObjectsOptions{
		Prefix:    "cards/",
		Recursive: true,
	})

	var images []string
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}
		images = append(images, object.Key)
	}

	return images, nil
}

// GeneratePresignedUploadURL generates a presigned URL for direct upload
func (mc *MinIOClient) GeneratePresignedUploadURL(cardID int, extension string, expiry time.Duration) (string, error) {
	ctx := context.Background()

	if extension == "" {
		extension = ".jpg"
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	objectName := fmt.Sprintf("cards/%d%s", cardID, extension)

	url, err := mc.client.PresignedPutObject(ctx, mc.bucketName, objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// GeneratePresignedDownloadURL generates a presigned URL for download
func (mc *MinIOClient) GeneratePresignedDownloadURL(cardID int, extension string, expiry time.Duration) (string, error) {
	ctx := context.Background()

	if extension == "" {
		extension = ".jpg"
	}
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	objectName := fmt.Sprintf("cards/%d%s", cardID, extension)

	url, err := mc.client.PresignedGetObject(ctx, mc.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}
