package models

import (
	"time"
)

type Report struct {
	ReportID           string        `db:"report_id" json:"report_id"`
	WorkID             string        `db:"work_id" json:"work_id"`
	FileID             string        `db:"file_id" json:"file_id"`
	StudentID          string        `db:"student_id" json:"student_id"`
	AssignmentID       string        `db:"assignment_id" json:"assignment_id"`
	PlagiarismScore    float32       `db:"plagiarism_score" json:"plagiarism_score"`
	IsPlagiarism       bool          `db:"is_plagiarism" json:"is_plagiarism"`
	WordCount          int           `db:"word_count" json:"word_count"`
	AnalysisDurationMs int           `db:"analysis_duration_ms" json:"analysis_duration_ms"`
	Status             string        `db:"status" json:"status"`
	ErrorMessage       *string       `db:"error_message" json:"error_message,omitempty"`
	CreatedAt          time.Time     `db:"created_at" json:"created_at"`
	SimilarWorks       []SimilarWork `json:"similar_works,omitempty"`
}

type SimilarWork struct {
	SimilarID            string  `db:"similar_id" json:"similar_id"`
	ReportID             string  `db:"report_id" json:"report_id"`
	OriginalWorkID       string  `db:"original_work_id" json:"original_work_id"`
	SimilarWorkID        string  `db:"similar_work_id" json:"similar_work_id"`
	SimilarityPercentage float32 `db:"similarity_percentage" json:"similarity_percentage"`
}

type AnalysisRequest struct {
	WorkID       string  `json:"work_id"`
	FileID       string  `json:"file_id"`
	StudentID    *string `json:"student_id,omitempty"`
	AssignmentID *string `json:"assignment_id,omitempty"`
}
