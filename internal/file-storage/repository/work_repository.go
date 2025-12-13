package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sd_hw3/internal/file-storage/models"
	"sd_hw3/pkg/db"
)

type WorkRepository interface {
	CreateWork(ctx context.Context, work *models.Work) error
	GetWorkByID(ctx context.Context, workID string) (*models.Work, error)
	GetOrCreateWork(ctx context.Context, studentID, assignmentID string) (*models.Work, error)
	DeleteWork(ctx context.Context, workID string) error
}

type workRepository struct {
	db *sql.DB
}

func NewWorkRepository() WorkRepository {
	return &workRepository{db: db.DB}
}

func (r *workRepository) CreateWork(ctx context.Context, work *models.Work) error {
	query := `
		INSERT INTO works (work_id, student_id, assignment_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := db.Exec(ctx, query,
		work.WorkID,
		work.StudentID,
		work.AssignmentID,
		work.CreatedAt,
		work.UpdatedAt,
	)

	return err
}

func (r *workRepository) GetWorkByID(ctx context.Context, workID string) (*models.Work, error) {
	query := `
		SELECT work_id, student_id, assignment_id, created_at, updated_at
		FROM works
		WHERE work_id = $1
	`

	row := db.QueryRow(ctx, query, workID)

	var work models.Work
	err := row.Scan(
		&work.WorkID,
		&work.StudentID,
		&work.AssignmentID,
		&work.CreatedAt,
		&work.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("work not found: %s", workID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get work: %w", err)
	}

	return &work, nil
}

func (r *workRepository) GetOrCreateWork(ctx context.Context, studentID, assignmentID string) (*models.Work, error) {
	// Сначала пытаемся найти существующую работу
	query := `
		SELECT work_id, student_id, assignment_id, created_at, updated_at
		FROM works
		WHERE student_id = $1 AND assignment_id = $2
	`

	row := db.QueryRow(ctx, query, studentID, assignmentID)

	var work models.Work
	err := row.Scan(
		&work.WorkID,
		&work.StudentID,
		&work.AssignmentID,
		&work.CreatedAt,
		&work.UpdatedAt,
	)

	if err == nil {
		// Работа найдена
		return &work, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		// Другая ошибка
		return nil, fmt.Errorf("failed to get work: %w", err)
	}

	// Работа не найдена, создаем новую
	work = models.Work{
		WorkID:       generateWorkID(studentID, assignmentID),
		StudentID:    studentID,
		AssignmentID: assignmentID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = r.CreateWork(ctx, &work)
	if err != nil {
		return nil, fmt.Errorf("failed to create work: %w", err)
	}

	return &work, nil
}

func (r *workRepository) DeleteWork(ctx context.Context, workID string) error {
	query := "DELETE FROM works WHERE work_id = $1"

	_, err := db.Exec(ctx, query, workID)
	return err
}

func generateWorkID(studentID, assignmentID string) string {
	return fmt.Sprintf("work-%s-%s-%d", studentID, assignmentID, time.Now().UnixNano())
}
