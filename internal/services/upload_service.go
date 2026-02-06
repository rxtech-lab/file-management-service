package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// UploadService handles file uploads to S3-compatible storage
type UploadService interface {
	UploadFile(ctx context.Context, userID string, filename string, content []byte, contentType string) (string, error)
	GetPresignedUploadURL(ctx context.Context, userID string, filename string, contentType string) (string, string, error)
	GetPresignedDownloadURL(ctx context.Context, key string) (string, error)
	DeleteFile(ctx context.Context, key string) error
	DownloadFile(ctx context.Context, key string) ([]byte, string, error)
	UploadToExistingKey(ctx context.Context, key string, content []byte, contentType string) error
}

// S3Config holds S3 configuration
type S3Config struct {
	Endpoint        string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	UsePathStyle    bool // For MinIO and other S3-compatible services
}

type uploadService struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

// NewUploadService creates a new UploadService with S3 configuration
func NewUploadService(cfg S3Config) (UploadService, error) {
	// Create custom resolver for endpoint
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if cfg.Endpoint != "" {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				HostnameImmutable: cfg.UsePathStyle,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with path-style addressing if needed
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	})

	presignClient := s3.NewPresignClient(client)

	return &uploadService{
		client:        client,
		presignClient: presignClient,
		bucket:        cfg.Bucket,
	}, nil
}

// UploadFile uploads a file to S3 and returns the object key
func (s *uploadService) UploadFile(ctx context.Context, userID string, filename string, content []byte, contentType string) (string, error) {
	// Generate unique key with user ID prefix
	ext := filepath.Ext(filename)
	key := fmt.Sprintf("files/%s/%s%s", userID, uuid.New().String(), ext)

	// Upload to S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return key, nil
}

// GetPresignedUploadURL generates a presigned URL for direct upload
// Returns the presigned URL and the object key
func (s *uploadService) GetPresignedUploadURL(ctx context.Context, userID string, filename string, contentType string) (string, string, error) {
	// Generate unique key with user ID prefix
	ext := filepath.Ext(filename)
	key := fmt.Sprintf("files/%s/%s%s", userID, uuid.New().String(), ext)

	// Generate presigned PUT URL
	presignResult, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignResult.URL, key, nil
}

// GetPresignedDownloadURL generates a presigned URL for downloading a file
func (s *uploadService) GetPresignedDownloadURL(ctx context.Context, key string) (string, error) {
	presignResult, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 1 * time.Hour
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return presignResult.URL, nil
}

// DownloadFile downloads file content from S3
func (s *uploadService) DownloadFile(ctx context.Context, key string) ([]byte, string, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to download file: %w", err)
	}
	defer result.Body.Close()

	content, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file content: %w", err)
	}

	contentType := ""
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	return content, contentType, nil
}

// UploadToExistingKey overwrites content at an existing S3 key
func (s *uploadService) UploadToExistingKey(ctx context.Context, key string, content []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}

// DeleteFile deletes a file from S3
func (s *uploadService) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// MockUploadService is a mock implementation for testing
type MockUploadService struct {
	files map[string][]byte
}

// NewMockUploadService creates a mock upload service for testing
func NewMockUploadService() UploadService {
	return &MockUploadService{
		files: make(map[string][]byte),
	}
}

func (m *MockUploadService) UploadFile(ctx context.Context, userID string, filename string, content []byte, contentType string) (string, error) {
	ext := filepath.Ext(filename)
	key := fmt.Sprintf("files/%s/%s%s", userID, uuid.New().String(), ext)
	m.files[key] = content
	return key, nil
}

func (m *MockUploadService) GetPresignedUploadURL(ctx context.Context, userID string, filename string, contentType string) (string, string, error) {
	ext := filepath.Ext(filename)
	key := fmt.Sprintf("files/%s/%s%s", userID, uuid.New().String(), ext)
	// Return a mock presigned URL
	return fmt.Sprintf("https://mock-s3.example.com/%s?presigned=true", key), key, nil
}

func (m *MockUploadService) GetPresignedDownloadURL(ctx context.Context, key string) (string, error) {
	return fmt.Sprintf("https://mock-s3.example.com/%s?download=true", key), nil
}

func (m *MockUploadService) DeleteFile(ctx context.Context, key string) error {
	delete(m.files, key)
	return nil
}

func (m *MockUploadService) DownloadFile(ctx context.Context, key string) ([]byte, string, error) {
	content, ok := m.files[key]
	if !ok {
		return nil, "", fmt.Errorf("file not found: %s", key)
	}
	return content, "application/octet-stream", nil
}

func (m *MockUploadService) UploadToExistingKey(ctx context.Context, key string, content []byte, contentType string) error {
	m.files[key] = content
	return nil
}
