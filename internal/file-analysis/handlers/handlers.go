package handlers

import (
	"fmt"
	"net/http"

	fileanalysis "sd_hw3/api/generated/file-analysis"
	"sd_hw3/internal/file-analysis/models"
	"sd_hw3/internal/file-analysis/repository"
	"sd_hw3/internal/file-analysis/service"

	"github.com/labstack/echo/v4"
)

// Handler реализует ServerInterface из сгенерированного кода
type Handler struct {
	service service.AnalysisService
}

// NewHandler создает новый обработчик
func NewHandler(svc service.AnalysisService) *Handler {
	return &Handler{
		service: svc,
	}
}

// AnalyzeFile анализирует файл и создает отчет о плагиате
func (h *Handler) AnalyzeFile(ctx echo.Context) error {
	var req models.AnalysisRequest

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, fileanalysis.ApiError{
			Error:   (*fileanalysis.ApiErrorError)(stringPtr("INVALID_REQUEST")),
			Message: stringPtr("Invalid request body"),
		})
	}

	if req.WorkID == "" || req.FileID == "" {
		return ctx.JSON(http.StatusBadRequest, fileanalysis.ApiError{
			Error:   (*fileanalysis.ApiErrorError)(stringPtr("MISSING_REQUIRED_FIELDS")),
			Message: stringPtr("work_id and file_id are required"),
		})
	}

	report, err := h.service.AnalyzeFile(ctx.Request().Context(), &req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, fileanalysis.ApiError{
			Error:   (*fileanalysis.ApiErrorError)(stringPtr("ANALYSIS_ERROR")),
			Message: stringPtr(fmt.Sprintf("Failed to analyze file: %v", err)),
		})
	}

	return ctx.JSON(http.StatusCreated, mapReportToResponse(report))
}

// GetReport получает отчет по ID
func (h *Handler) GetReport(ctx echo.Context, reportId string) error {
	report, err := h.service.GetReport(ctx.Request().Context(), reportId)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, fileanalysis.ApiError{
			Error:   (*fileanalysis.ApiErrorError)(stringPtr("REPORT_NOT_FOUND")),
			Message: stringPtr(fmt.Sprintf("Report with id %s not found", reportId)),
		})
	}

	return ctx.JSON(http.StatusOK, mapReportToResponse(report))
}

// GetReportsByWorkID получает отчеты по ID работы
func (h *Handler) GetWorkReports(ctx echo.Context, workId string) error {
	reports, err := h.service.GetWorkReports(ctx.Request().Context(), workId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, fileanalysis.ApiError{
			Error:   (*fileanalysis.ApiErrorError)(stringPtr("FETCH_ERROR")),
			Message: stringPtr(fmt.Sprintf("Failed to fetch reports: %v", err)),
		})
	}

	var response []map[string]interface{}
	for _, report := range reports {
		response = append(response, mapReportToResponse(report))
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"reports": response,
		"count":   len(response),
	})
}

// ListReports список отчетов с пагинацией
func (h *Handler) ListReports(ctx echo.Context, params fileanalysis.ListReportsParams) error {
	page := 1
	pageSize := 10

	listParams := repository.ListReportsParams{
		Offset:       *params.Offset,
		Limit:        *params.Limit,
		StudentID:    params.StudentId,
		AssignmentID: params.AssignmentId,
		WorkID:       params.WorkId,
		FileID:       params.FileId,
	}

	reports, total, err := h.service.ListReports(ctx.Request().Context(), listParams)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, fileanalysis.ApiError{
			Error:   (*fileanalysis.ApiErrorError)(stringPtr("FETCH_ERROR")),
			Message: stringPtr(fmt.Sprintf("Failed to fetch reports: %v", err)),
		})
	}

	var response []map[string]interface{}
	for _, report := range reports {
		response = append(response, mapReportToResponse(report))
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"reports": response,
		"total":   total,
		"limit":   (total + pageSize - 1) / pageSize,
		"offset":  page,
	})
}

// Вспомогательные функции
func stringPtr(s string) *string {
	return &s
}

func mapReportToResponse(report *models.Report) map[string]interface{} {
	response := map[string]interface{}{
		"report_id":            report.ReportID,
		"work_id":              report.WorkID,
		"file_id":              report.FileID,
		"student_id":           report.StudentID,
		"assignment_id":        report.AssignmentID,
		"plagiarism_score":     report.PlagiarismScore,
		"is_plagiarism":        report.IsPlagiarism,
		"word_count":           report.WordCount,
		"analysis_duration_ms": report.AnalysisDurationMs,
		"status":               report.Status,
		"created_at":           report.CreatedAt,
	}

	if report.ErrorMessage != nil {
		response["error_message"] = *report.ErrorMessage
	}

	if len(report.SimilarWorks) > 0 {
		var similarWorks []map[string]interface{}
		for _, sw := range report.SimilarWorks {
			similarWorks = append(similarWorks, map[string]interface{}{
				"similar_id":            sw.SimilarID,
				"report_id":             sw.ReportID,
				"original_work_id":      sw.OriginalWorkID,
				"similar_work_id":       sw.SimilarWorkID,
				"similarity_percentage": sw.SimilarityPercentage,
			})
		}
		response["similar_works"] = similarWorks
	}

	return response
}
