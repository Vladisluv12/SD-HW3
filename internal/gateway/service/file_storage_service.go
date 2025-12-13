package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"sd_hw3/internal/gateway/models"
)

type FileStorageService interface {
	UploadFile(ctx context.Context, studentID, assignmentID string, file *multipart.FileHeader) (*models.WorkSubmissionResponse, error)
	DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, string, int64, error)
	GetFileMetadata(ctx context.Context, fileID string) (*models.FileMetadata, error)
}

type fileStorageServiceImpl struct {
	baseURL string
	client  *http.Client
}

func NewFileStorageService(baseURL string) FileStorageService {
	return &fileStorageServiceImpl{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  &http.Client{},
	}
}

func (s *fileStorageServiceImpl) UploadFile(ctx context.Context, studentID, assignmentID string, file *multipart.FileHeader) (*models.WorkSubmissionResponse, error) {
	url := fmt.Sprintf("%s/files", s.baseURL)

	// Открываем файл
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Создаем multipart запрос
	body := &strings.Builder{}
	writer := multipart.NewWriter(body)

	// Добавляем поля
	writer.WriteField("student_id", studentID)
	writer.WriteField("assignment_id", assignmentID)

	// Добавляем файл
	part, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, src); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	writer.Close()

	// Создаем запрос
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Отправляем запрос
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("file storage returned status: %d", resp.StatusCode)
	}

	// Парсим ответ
	var uploadResp models.WorkSubmissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &uploadResp, nil
}

func (s *fileStorageServiceImpl) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, string, int64, error) {
	url := fmt.Sprintf("%s/files/%s", s.baseURL, fileID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to download file: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, "", 0, fmt.Errorf("file storage returned status: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	contentLength := resp.ContentLength

	return resp.Body, contentType, contentLength, nil
}

func (s *fileStorageServiceImpl) GetFileMetadata(ctx context.Context, fileID string) (*models.FileMetadata, error) {
	url := fmt.Sprintf("%s/files/%s/metadata", s.baseURL, fileID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file storage returned status: %d", resp.StatusCode)
	}

	var metadata models.FileMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return &metadata, nil
}
