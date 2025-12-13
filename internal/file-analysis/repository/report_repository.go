package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"sd_hw3/internal/file-analysis/models"
	"sd_hw3/pkg/db"
)

type ReportRepository interface {
	CreateReport(ctx context.Context, report *models.Report) error
	GetReport(ctx context.Context, reportID string) (*models.Report, error)
	GetReportsByWorkID(ctx context.Context, workID string) ([]*models.Report, error)
	ListReports(ctx context.Context, params ListReportsParams) ([]*models.Report, int, error)
	UpdateReportStatus(ctx context.Context, reportID, status string, errorMsg *string) error
	AddSimilarWork(ctx context.Context, similar *models.SimilarWork) error
	GetSimilarWorks(ctx context.Context, reportID string) ([]models.SimilarWork, error)
	DeleteReport(ctx context.Context, reportID string) error
}

type reportRepository struct {
	db *sql.DB
}

func NewReportRepository() ReportRepository {
	return &reportRepository{db: db.DB}
}

type ListReportsParams struct {
	WorkID       *string
	FileID       *string
	AssignmentID *string
	StudentID    *string
	Limit        int
	Offset       int
}

func (r *reportRepository) CreateReport(ctx context.Context, report *models.Report) error {
	query := `
		INSERT INTO reports (
			report_id, work_id, file_id, student_id, assignment_id,
			plagiarism_score, is_plagiarism, word_count,
			analysis_duration_ms, status, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := db.Exec(ctx, query,
		report.ReportID,
		report.WorkID,
		report.FileID,
		report.StudentID,
		report.AssignmentID,
		report.PlagiarismScore,
		report.IsPlagiarism,
		report.WordCount,
		report.AnalysisDurationMs,
		report.Status,
		report.ErrorMessage,
		report.CreatedAt,
	)

	return err
}

func (r *reportRepository) GetReport(ctx context.Context, reportID string) (*models.Report, error) {
	query := `
		SELECT 
			report_id, work_id, file_id, student_id, assignment_id,
			plagiarism_score, is_plagiarism, word_count,
			analysis_duration_ms, status, error_message, created_at
		FROM reports
		WHERE report_id = $1
	`

	row := db.QueryRow(ctx, query, reportID)
	return r.scanReport(row)
}

func (r *reportRepository) GetReportsByWorkID(ctx context.Context, workID string) ([]*models.Report, error) {
	query := `
		SELECT 
			report_id, work_id, file_id, student_id, assignment_id,
			plagiarism_score, is_plagiarism, word_count,
			analysis_duration_ms, status, error_message, created_at
		FROM reports
		WHERE work_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.Query(ctx, query, workID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*models.Report
	for rows.Next() {
		report, err := r.scanReportFromRows(rows)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	return reports, nil
}

func (r *reportRepository) ListReports(ctx context.Context, params ListReportsParams) ([]*models.Report, int, error) {
	// Строим динамический запрос
	whereClauses := []string{}
	args := []interface{}{}
	argPos := 1

	if params.WorkID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("work_id = $%d", argPos))
		args = append(args, *params.WorkID)
		argPos++
	}
	if params.FileID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("file_id = $%d", argPos))
		args = append(args, *params.FileID)
		argPos++
	}
	if params.AssignmentID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("assignment_id = $%d", argPos))
		args = append(args, *params.AssignmentID)
		argPos++
	}
	if params.StudentID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("student_id = $%d", argPos))
		args = append(args, *params.StudentID)
		argPos++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Получаем общее количество
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM reports %s", whereSQL)
	var total int
	row := db.QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&total); err != nil {
		return nil, 0, err
	}

	// Получаем данные с пагинацией
	args = append(args, params.Limit, params.Offset)
	dataQuery := fmt.Sprintf(`
		SELECT 
			report_id, work_id, file_id, student_id, assignment_id,
			plagiarism_score, is_plagiarism, word_count,
			analysis_duration_ms, status, error_message, created_at
		FROM reports %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argPos, argPos+1)

	rows, err := db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reports []*models.Report
	for rows.Next() {
		report, err := r.scanReportFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		reports = append(reports, report)
	}

	return reports, total, nil
}

func (r *reportRepository) UpdateReportStatus(ctx context.Context, reportID, status string, errorMsg *string) error {
	query := `
		UPDATE reports 
		SET status = $2, error_message = $3 
		WHERE report_id = $1
	`
	_, err := db.Exec(ctx, query, reportID, status, errorMsg)
	return err
}

func (r *reportRepository) AddSimilarWork(ctx context.Context, similar *models.SimilarWork) error {
	query := `
		INSERT INTO similar_works (
			similar_id, report_id, original_work_id, similar_work_id, similarity_percentage
		) VALUES ($1, $2, $3, $4, $5)
	`
	_, err := db.Exec(ctx, query,
		similar.SimilarID,
		similar.ReportID,
		similar.OriginalWorkID,
		similar.SimilarWorkID,
		similar.SimilarityPercentage,
	)
	return err
}

func (r *reportRepository) GetSimilarWorks(ctx context.Context, reportID string) ([]models.SimilarWork, error) {
	query := `
		SELECT similar_id, report_id, original_work_id, similar_work_id, similarity_percentage
		FROM similar_works
		WHERE report_id = $1
		ORDER BY similarity_percentage DESC
	`

	rows, err := db.Query(ctx, query, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var similarWorks []models.SimilarWork
	for rows.Next() {
		var sw models.SimilarWork
		err := rows.Scan(
			&sw.SimilarID,
			&sw.ReportID,
			&sw.OriginalWorkID,
			&sw.SimilarWorkID,
			&sw.SimilarityPercentage,
		)
		if err != nil {
			return nil, err
		}
		similarWorks = append(similarWorks, sw)
	}

	return similarWorks, nil
}

func (r *reportRepository) DeleteReport(ctx context.Context, reportID string) error {
	query := "DELETE FROM reports WHERE report_id = $1"
	_, err := db.Exec(ctx, query, reportID)
	return err
}

func (r *reportRepository) scanReport(row *sql.Row) (*models.Report, error) {
	var report models.Report
	err := row.Scan(
		&report.ReportID,
		&report.WorkID,
		&report.FileID,
		&report.StudentID,
		&report.AssignmentID,
		&report.PlagiarismScore,
		&report.IsPlagiarism,
		&report.WordCount,
		&report.AnalysisDurationMs,
		&report.Status,
		&report.ErrorMessage,
		&report.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportRepository) scanReportFromRows(rows *sql.Rows) (*models.Report, error) {
	var report models.Report
	err := rows.Scan(
		&report.ReportID,
		&report.WorkID,
		&report.FileID,
		&report.StudentID,
		&report.AssignmentID,
		&report.PlagiarismScore,
		&report.IsPlagiarism,
		&report.WordCount,
		&report.AnalysisDurationMs,
		&report.Status,
		&report.ErrorMessage,
		&report.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &report, nil
}
