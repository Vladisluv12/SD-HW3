package handlers

import (
	filestorage "sd_hw3/api/generated/file-storage"
	"sd_hw3/internal/file-storage/models"
	"sd_hw3/internal/file-storage/service"
)

// MapFileToUploadResponse конвертирует модели File и Work в FileUploadResponse
func MapFileToUploadResponse(file *models.File, work *models.Work) filestorage.FileUploadResponse {
	return filestorage.FileUploadResponse{
		FileId:      file.FileID,
		WorkId:      work.WorkID,
		Filename:    stringPtr(file.OriginalFilename),
		SizeBytes:   intPtr(int(file.SizeBytes)),
		UploadedAt:  &file.UploadedAt,
		StoragePath: &file.StoragePath,
	}
}

// MapFileToMetadata конвертирует модель File в FileMetadata
func MapFileMetaToMetadata(filemeta *service.FileMetadata) filestorage.FileMetadata {
	// Получаем work для получения student_id и assignment_id
	// В реальной реализации здесь нужно получить связанную работу

	return filestorage.FileMetadata{
		FileId:       stringPtr(filemeta.FileID),
		WorkId:       stringPtr(filemeta.WorkID),
		StudentId:    stringPtr(filemeta.StudentID),
		AssignmentId: stringPtr(filemeta.AssignmentID),
		Filename:     stringPtr(filemeta.Filename),
		ContentType:  filemeta.ContentType,
		SizeBytes:    int64Ptr(filemeta.SizeBytes),
		UploadedAt:   &filemeta.UploadedAt,
		Checksum:     filemeta.ChecksumMD5, // Используем MD5
	}
}

// MapFileToApiError создает ApiError из ошибки
func MapFileToApiError(err error, code, message string) filestorage.ApiError {
	details := err.Error()
	return filestorage.ApiError{
		Error:   stringPtr(code),
		Message: stringPtr(message),
		Details: &details,
	}
}

func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}
