package models

import (
	"sd_hw3/internal/file-storage" // adjust path to match your module name
)

// DB model to DTO mapper for File
func FileToFileMetadata(file *File) *filestorage.FileMetadata {
	if file == nil {
		return nil
	}
	return &filestorage.FileMetadata{
		FileId:       &file.FileID,
		WorkId:       &file.WorkID,
		Filename:     &file.Filename,
		AssignmentId: nil, // Не хранится в File
		StudentId:    nil, // Не хранится в File
		ContentType:  file.ContentType,
		SizeBytes:    &int(file.SizeBytes),
		UploadedAt:   &file.UploadedAt,
		Checksum:     file.ChecksumSHA256,
	}
}
