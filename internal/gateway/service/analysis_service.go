package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"sd_hw3/internal/gateway/models"
)

type FileAnalysisService interface {
	AnalyzeFile(ctx context.Context, req *models.AnalysisRequest) (*models.Report, error)
	GetReport(ctx context.Context, reportID string) (*models.Report, error)
	GetWorkReports(ctx context.Context, workID string) ([]*models.Report, error)
	ListReports(ctx context.Context, params *models.ListReportsParams) (*models.ReportListResponse, error)
}

type fileAnalysisServiceImpl struct {
	baseURL string
	client  *http.Client
}

func NewFileAnalysisService(baseURL string) FileAnalysisService {
	return &fileAnalysisServiceImpl{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (s *fileAnalysisServiceImpl) AnalyzeFile(ctx context.Context, req *models.AnalysisRequest) (*models.Report, error) {
	url := fmt.Sprintf("%s/analyze", s.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file analysis returned status: %d", resp.StatusCode)
	}

	var report models.Report
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &report, nil
}

func (s *fileAnalysisServiceImpl) GetReport(ctx context.Context, reportID string) (*models.Report, error) {
	url := fmt.Sprintf("%s/reports/%s", s.baseURL, reportID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file analysis returned status: %d", resp.StatusCode)
	}

	var report models.Report
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return nil, fmt.Errorf("failed to decode report: %w", err)
	}

	return &report, nil
}

func (s *fileAnalysisServiceImpl) GetWorkReports(ctx context.Context, workID string) ([]*models.Report, error) {
	url := fmt.Sprintf("%s/reports/work/%s", s.baseURL, workID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get work reports: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file analysis returned status: %d", resp.StatusCode)
	}

	var reports []*models.Report
	if err := json.NewDecoder(resp.Body).Decode(&reports); err != nil {
		return nil, fmt.Errorf("failed to decode reports: %w", err)
	}

	return reports, nil
}

func (s *fileAnalysisServiceImpl) ListReports(ctx context.Context, params *models.ListReportsParams) (*models.ReportListResponse, error) {
	url := fmt.Sprintf("%s/reports", s.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Добавляем query параметры
	q := req.URL.Query()
	if params.WorkID != nil {
		q.Add("work_id", *params.WorkID)
	}
	if params.FileID != nil {
		q.Add("file_id", *params.FileID)
	}
	if params.AssignmentID != nil {
		q.Add("assignment_id", *params.AssignmentID)
	}
	if params.StudentID != nil {
		q.Add("student_id", *params.StudentID)
	}
	if params.Limit != nil {
		q.Add("limit", fmt.Sprintf("%d", *params.Limit))
	}
	if params.Offset != nil {
		q.Add("offset", fmt.Sprintf("%d", *params.Offset))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file analysis returned status: %d", resp.StatusCode)
	}

	var response models.ReportListResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}
