package repository

import (
	"context"
	"database/sql"
	"fmt"

	"sd_hw3/internal/file-storage/models"
	"sd_hw3/pkg/db"
)

type FileRepository interface {
	CreateFile(ctx context.Context, file *models.File) error
	GetFileByID(ctx context.Context, fileID string) (*models.File, error)
	GetFilesByWorkID(ctx context.Context, workID string) ([]*models.File, error)
	GetFileMetadata(ctx context.Context, fileID string) (*models.File, error)
	DeleteFile(ctx context.Context, fileID string) error
	UpdateFile(ctx context.Context, file *models.File) error
	GetFilesByChecksum(ctx context.Context, checksum string, excludedFile string) ([]*models.File, error)
}

type fileRepository struct {
	db *sql.DB
}

func NewFileRepository() FileRepository {
	return &fileRepository{db: db.DB}
}

func (r *fileRepository) CreateFile(ctx context.Context, file *models.File) error {
	query := `
		INSERT INTO files (
			file_id, work_id, filename, original_filename, 
			content_type, size_bytes, storage_path,
			checksum_md5, checksum_sha256, uploaded_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := db.Exec(ctx, query,
		file.FileID,
		file.WorkID,
		file.Filename,
		file.OriginalFilename,
		file.ContentType,
		file.SizeBytes,
		file.StoragePath,
		file.ChecksumMD5,
		file.ChecksumSHA256,
		file.UploadedAt,
	)

	return err
}

func (r *fileRepository) GetFileByID(ctx context.Context, fileID string) (*models.File, error) {
	query := `
		SELECT 
			file_id, work_id, filename, original_filename,
			content_type, size_bytes, storage_path,
			checksum_md5, checksum_sha256, uploaded_at
		FROM files
		WHERE file_id = $1
	`

	row := db.QueryRow(ctx, query, fileID)

	return r.scanFile(row)
}

func (r *fileRepository) GetFileMetadata(ctx context.Context, fileID string) (*models.File, error) {
	// Та же логика, что и GetFileByID, но можно добавить оптимизации
	return r.GetFileByID(ctx, fileID)
}

func (r *fileRepository) GetFilesByWorkID(ctx context.Context, workID string) ([]*models.File, error) {
	query := `
		SELECT 
			file_id, work_id, filename, original_filename,
			content_type, size_bytes, storage_path,
			checksum_md5, checksum_sha256, uploaded_at
		FROM files
		WHERE work_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := db.Query(ctx, query, workID)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	var files []*models.File
	for rows.Next() {
		file, err := r.scanFileFromRows(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return files, nil
}

func (r *fileRepository) DeleteFile(ctx context.Context, fileID string) error {
	query := "DELETE FROM files WHERE file_id = $1"

	_, err := db.Exec(ctx, query, fileID)
	return err
}

func (r *fileRepository) UpdateFile(ctx context.Context, file *models.File) error {
	query := `
		UPDATE files SET
			filename = $2,
			original_filename = $3,
			content_type = $4,
			size_bytes = $5,
			storage_path = $6,
			checksum_md5 = $7,
			checksum_sha256 = $8
		WHERE file_id = $1
	`

	_, err := db.Exec(ctx, query,
		file.FileID,
		file.Filename,
		file.OriginalFilename,
		file.ContentType,
		file.SizeBytes,
		file.StoragePath,
		file.ChecksumMD5,
		file.ChecksumSHA256,
	)

	return err
}

func (r *fileRepository) GetFilesByChecksum(ctx context.Context, checksum string, excludeFileID string) ([]*models.File, error) {
	query := `
		SELECT 
			file_id, work_id, filename, original_filename,
			content_type, size_bytes, storage_path,
			checksum_md5, checksum_sha256, uploaded_at
		FROM files
		WHERE checksum_md5 = $1 AND file_id != $2
		LIMIT 1
	`

	rows, err := db.Query(ctx, query, checksum, excludeFileID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	files := []*models.File{}
	for rows.Next() {
		file, err := r.scanFileFromRows(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func (r *fileRepository) scanFile(row *sql.Row) (*models.File, error) {
	var file models.File

	err := row.Scan(
		&file.FileID,
		&file.WorkID,
		&file.Filename,
		&file.OriginalFilename,
		&file.ContentType,
		&file.SizeBytes,
		&file.StoragePath,
		&file.ChecksumMD5,
		&file.ChecksumSHA256,
		&file.UploadedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("file not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan file: %w", err)
	}

	return &file, nil
}

func (r *fileRepository) scanFileFromRows(rows *sql.Rows) (*models.File, error) {
	var file models.File

	err := rows.Scan(
		&file.FileID,
		&file.WorkID,
		&file.Filename,
		&file.OriginalFilename,
		&file.ContentType,
		&file.SizeBytes,
		&file.StoragePath,
		&file.ChecksumMD5,
		&file.ChecksumSHA256,
		&file.UploadedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan file from rows: %w", err)
	}

	return &file, nil
}
