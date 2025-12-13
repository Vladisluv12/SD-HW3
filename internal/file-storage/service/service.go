package service

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sd_hw3/internal/file-storage/models"
	"sd_hw3/internal/file-storage/repository"
	"sd_hw3/pkg/config"
)

// StorageService реализация интерфейса работы с хранилищем
type StorageService struct {
	config         config.Config
	workRepo       repository.WorkRepository
	fileRepo       repository.FileRepository
	storageBaseDir string
}

type FileMetadata struct {
	FileID       string
	WorkID       string
	Filename     string
	ContentType  *string
	SizeBytes    int64
	StudentID    string
	AssignmentID string
	UploadedAt   time.Time
	ChecksumMD5  *string
}

// NewStorageService создает новый сервис
func NewStorageService(config config.Config) (*StorageService, error) {
	// Создаем директорию для хранения если не существует
	if err := os.MkdirAll(config.UploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &StorageService{
		config:         config,
		workRepo:       repository.NewWorkRepository(),
		fileRepo:       repository.NewFileRepository(),
		storageBaseDir: config.UploadDir,
	}, nil
}

// UploadFile загружает файл
func (s *StorageService) UploadFile(ctx context.Context, studentID, assignmentID string, fileData []byte, filename, contentType string, size int64) (*models.File, *models.Work, error) {
	// Получаем или создаем работу
	work, err := s.workRepo.GetOrCreateWork(ctx, studentID, assignmentID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get or create work: %w", err)
	}

	// Генерируем уникальный ID файла
	fileID := generateFileID()

	// Вычисляем контрольные суммы
	md5Hash := calculateMD5(fileData)
	sha256Hash := calculateSHA256(fileData)

	// Создаем путь для хранения (используем поддиректории по первым 2 символам ID)
	storageDir := filepath.Join(s.storageBaseDir, fileID[:2])
	storagePath := filepath.Join(storageDir, fileID)

	// Создаем поддиректорию если нужно
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Сохраняем файл на диск
	if err := os.WriteFile(storagePath, fileData, 0644); err != nil {
		return nil, nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Создаем модель файла
	file := &models.File{
		FileID:           fileID,
		WorkID:           work.WorkID,
		Filename:         fileID, // Внутреннее имя файла
		OriginalFilename: filename,
		ContentType:      &contentType,
		SizeBytes:        size,
		StoragePath:      storagePath,
		ChecksumMD5:      &md5Hash,
		ChecksumSHA256:   &sha256Hash,
		UploadedAt:       time.Now(),
	}

	// Сохраняем в БД
	if err := s.fileRepo.CreateFile(ctx, file); err != nil {
		// Удаляем файл если не удалось сохранить в БД
		os.Remove(storagePath)
		return nil, nil, fmt.Errorf("failed to save file metadata to database: %w", err)
	}

	return file, work, nil
}

// GetFile получает информацию о файле
func (s *StorageService) GetFile(ctx context.Context, fileID string) (*models.File, error) {
	return s.fileRepo.GetFileByID(ctx, fileID)
}

// GetFileMetadata получает метаданные файла
func (s *StorageService) GetFileMetadata(ctx context.Context, fileID string) (*FileMetadata, error) {
	file, err := s.fileRepo.GetFileMetadata(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}
	work, err := s.workRepo.GetWorkByID(ctx, file.WorkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get work for file: %w", err)
	}
	return &FileMetadata{
		FileID:       file.FileID,
		WorkID:       file.WorkID,
		Filename:     file.OriginalFilename,
		ContentType:  file.ContentType,
		SizeBytes:    file.SizeBytes,
		StudentID:    work.StudentID,
		AssignmentID: work.AssignmentID,
		UploadedAt:   file.UploadedAt,
		ChecksumMD5:  file.ChecksumMD5,
	}, nil
}

// GetFileContent получает содержимое файла
func (s *StorageService) GetFileContent(ctx context.Context, fileID string) ([]byte, error) {
	file, err := s.fileRepo.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	// Проверяем существование файла
	if _, err := os.Stat(file.StoragePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found on disk: %s", file.StoragePath)
	}

	return os.ReadFile(file.StoragePath)
}

// GetFilesByWorkID получает все файлы работы
func (s *StorageService) GetFilesByWorkID(ctx context.Context, workID string) ([]*models.File, error) {
	return s.fileRepo.GetFilesByWorkID(ctx, workID)
}

// DeleteFile удаляет файл
func (s *StorageService) DeleteFile(ctx context.Context, fileID string) error {
	// Получаем информацию о файле
	file, err := s.fileRepo.GetFileByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Удаляем из БД
	if err := s.fileRepo.DeleteFile(ctx, fileID); err != nil {
		return fmt.Errorf("failed to delete file from database: %w", err)
	}

	// Удаляем файл с диска
	if err := os.Remove(file.StoragePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file from disk: %w", err)
	}

	return nil
}

func (s *StorageService) CheckFileExists(ctx context.Context, fileID string) ([]string, error) {
	file, err := s.fileRepo.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	if file.ChecksumMD5 == nil {
		return nil, nil
	}
	files, err := s.fileRepo.GetFilesByChecksum(ctx, *file.ChecksumMD5, fileID)
	if err != nil {
		return nil, nil
	}
	file_ids := []string{}
	for _, f := range files {
		file_ids = append(file_ids, f.FileID)
	}
	return file_ids, nil
}

// Вспомогательные функции
func generateFileID() string {
	return fmt.Sprintf("file-%d-%s",
		time.Now().UnixNano(),
		hex.EncodeToString(md5.New().Sum([]byte(fmt.Sprint(time.Now().UnixNano()))))[:8])
}

func calculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func calculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
