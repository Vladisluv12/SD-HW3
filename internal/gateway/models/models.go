package models

import (
	"time"
)

// Модели для работы с микросервисами

type WorkSubmissionResponse struct {
	WorkID      string    `json:"work_id"`
	FileID      string    `json:"file_id"`
	Filename    string    `json:"filename,omitempty"`
	SizeBytes   int64     `json:"size_bytes,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at,omitempty"`
	StoragePath string    `json:"storage_path,omitempty"`
}

type FileMetadata struct {
	FileID       string    `json:"file_id"`
	WorkID       string    `json:"work_id"`
	StudentID    string    `json:"student_id"`
	AssignmentID string    `json:"assignment_id"`
	Filename     string    `json:"filename"`
	ContentType  string    `json:"content_type"`
	SizeBytes    int64     `json:"size_bytes"`
	UploadedAt   time.Time `json:"uploaded_at"`
	Checksum     *string   `json:"checksum,omitempty"`
}

type AnalysisRequest struct {
	WorkID       string  `json:"work_id"`
	FileID       string  `json:"file_id"`
	StudentID    *string `json:"student_id,omitempty"`
	AssignmentID *string `json:"assignment_id,omitempty"`
}

type Report struct {
	ReportID           string        `json:"report_id"`
	WorkID             string        `json:"work_id"`
	FileID             string        `json:"file_id"`
	StudentID          *string       `json:"student_id,omitempty"`
	AssignmentID       *string       `json:"assignment_id,omitempty"`
	PlagiarismScore    *float32      `json:"plagiarism_score,omitempty"`
	IsPlagiarism       *bool         `json:"is_plagiarism,omitempty"`
	SimilarWorks       []SimilarWork `json:"similar_works,omitempty"`
	WordCount          *int          `json:"word_count,omitempty"`
	AnalysisDurationMs *int          `json:"analysis_duration_ms,omitempty"`
	Status             string        `json:"status"`
	ErrorMessage       *string       `json:"error_message,omitempty"`
	CreatedAt          time.Time     `json:"created_at"`
}

type SimilarWork struct {
	WorkID               string   `json:"work_id"`
	StudentID            *string  `json:"student_id,omitempty"`
	SimilarityPercentage *float32 `json:"similarity_percentage,omitempty"`
}

type ListReportsParams struct {
	WorkID       *string `json:"work_id,omitempty"`
	FileID       *string `json:"file_id,omitempty"`
	AssignmentID *string `json:"assignment_id,omitempty"`
	StudentID    *string `json:"student_id,omitempty"`
	Limit        *int    `json:"limit,omitempty"`
	Offset       *int    `json:"offset,omitempty"`
}

type ReportListResponse struct {
	Reports []*Report `json:"reports"`
	Total   int       `json:"total"`
	Limit   int       `json:"limit"`
	Offset  int       `json:"offset"`
}
