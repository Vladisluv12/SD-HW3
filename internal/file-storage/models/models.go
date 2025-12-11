package models

import "time"

type Work struct {
	WorkID       string    `db:"work_id" json:"work_id"`
	StudentID    string    `db:"student_id" json:"student_id"`
	AssignmentID string    `db:"assignment_id" json:"assignment_id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type File struct {
	FileID           string    `db:"file_id" json:"file_id"`
	WorkID           string    `db:"work_id" json:"work_id"`
	Filename         string    `db:"filename" json:"filename"`
	OriginalFilename string    `db:"original_filename" json:"original_filename"`
	ContentType      *string   `db:"content_type" json:"content_type,omitempty"`
	SizeBytes        int64     `db:"size_bytes" json:"size_bytes"`
	StoragePath      string    `db:"storage_path" json:"storage_path"`
	ChecksumMD5      *string   `db:"checksum_md5" json:"checksum_md5,omitempty"`
	ChecksumSHA256   *string   `db:"checksum_sha256" json:"checksum_sha256,omitempty"`
	UploadedAt       time.Time `db:"uploaded_at" json:"uploaded_at"`
}
