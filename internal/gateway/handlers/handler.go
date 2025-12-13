package handlers

import (
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	gateway "sd_hw3/api/generated/gateway"
	"sd_hw3/internal/gateway/models"
	"sd_hw3/internal/gateway/service"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	fileStorageService  service.FileStorageService
	fileAnalysisService service.FileAnalysisService
}

func NewHandler(fileStorageService service.FileStorageService, fileAnalysisService service.FileAnalysisService) *Handler {
	return &Handler{
		fileStorageService:  fileStorageService,
		fileAnalysisService: fileAnalysisService,
	}
}

func (h *Handler) SubmitWork(ctx echo.Context) error {
	// Парсим multipart форму
	form, err := ctx.MultipartForm()
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, gateway.ApiError{
			Error:     (*gateway.ApiErrorError)(stringPtr("VALIDATION_ERROR")),
			Message:   stringPtr("Invalid multipart form data"),
			Timestamp: &[]time.Time{time.Now()}[0],
		})
	}

	// Получаем поля формы
	studentID := getFormValue(form, "student_id")
	assignmentID := getFormValue(form, "assignment_id")

	if studentID == "" || assignmentID == "" {
		return ctx.JSON(http.StatusBadRequest, gateway.ApiError{
			Error:     (*gateway.ApiErrorError)(stringPtr("VALIDATION_ERROR")),
			Message:   stringPtr("student_id and assignment_id are required"),
			Timestamp: &[]time.Time{time.Now()}[0],
		})
	}

	// Получаем файл
	files := form.File["file"]
	if len(files) == 0 {
		return ctx.JSON(http.StatusBadRequest, gateway.ApiError{
			Error:     (*gateway.ApiErrorError)(stringPtr("VALIDATION_ERROR")),
			Message:   stringPtr("No file uploaded"),
			Timestamp: &[]time.Time{time.Now()}[0],
		})
	}

	fileHeader := files[0]

	// 1. Загружаем файл в хранилище
	uploadResp, err := h.fileStorageService.UploadFile(ctx.Request().Context(), studentID, assignmentID, fileHeader)
	if err != nil {
		return ctx.JSON(http.StatusServiceUnavailable, gateway.ApiError{
			Error:     (*gateway.ApiErrorError)(stringPtr("SERVICE_UNAVAILABLE")),
			Message:   stringPtr("File storage service unavailable"),
			Timestamp: &[]time.Time{time.Now()}[0],
		})
	}

	// 2. Запускаем анализ файла
	analysisReq := &models.AnalysisRequest{
		WorkID:       uploadResp.WorkID,
		FileID:       uploadResp.FileID,
		StudentID:    &studentID,
		AssignmentID: &assignmentID,
	}

	_, err = h.fileAnalysisService.AnalyzeFile(ctx.Request().Context(), analysisReq)
	if err != nil {
		// Анализ может завершиться ошибкой, но файл уже загружен
		// Возвращаем успешный ответ с информацией о загрузке
		return ctx.JSON(http.StatusOK, gateway.WorkSubmissionResponse{
			WorkId:      stringPtr(uploadResp.WorkID),
			FileId:      stringPtr(uploadResp.FileID),
			SubmittedAt: &uploadResp.UploadedAt,
		})
	}

	// 3. Возвращаем ответ
	response := gateway.WorkSubmissionResponse{
		WorkId:      stringPtr(uploadResp.WorkID),
		FileId:      stringPtr(uploadResp.FileID),
		SubmittedAt: &uploadResp.UploadedAt,
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h *Handler) GetWorkReports(ctx echo.Context, workId string) error {
	reports, err := h.fileAnalysisService.GetWorkReports(ctx.Request().Context(), workId)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, gateway.ApiError{
			Error:     (*gateway.ApiErrorError)(stringPtr("NOT_FOUND")),
			Message:   stringPtr("Reports not found for this work"),
			Timestamp: &[]time.Time{time.Now()}[0],
		})
	}

	return ctx.JSON(http.StatusOK, reports)
}

func (h *Handler) DownloadFile(ctx echo.Context, fileId string) error {
	// Получаем метаданные файла
	metadata, err := h.fileStorageService.GetFileMetadata(ctx.Request().Context(), fileId)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, gateway.ApiError{
			Error:     (*gateway.ApiErrorError)(stringPtr("NOT_FOUND")),
			Message:   stringPtr("File not found"),
			Timestamp: &[]time.Time{time.Now()}[0],
		})
	}

	// Скачиваем файл
	content, contentType, contentLength, err := h.fileStorageService.DownloadFile(ctx.Request().Context(), fileId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, gateway.ApiError{
			Error:     (*gateway.ApiErrorError)(stringPtr("INTERNAL_ERROR")),
			Message:   stringPtr("Failed to download file"),
			Timestamp: &[]time.Time{time.Now()}[0],
		})
	}
	defer content.Close()

	// Устанавливаем заголовки
	ctx.Response().Header().Set("Content-Type", contentType)
	ctx.Response().Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	ctx.Response().Header().Set("Content-Disposition",
		"attachment; filename=\""+metadata.Filename+"\"")

	// Отправляем файл
	_, err = io.Copy(ctx.Response().Writer, content)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, gateway.ApiError{
			Error:     (*gateway.ApiErrorError)(stringPtr("INTERNAL_ERROR")),
			Message:   stringPtr("Failed to send file"),
			Timestamp: &[]time.Time{time.Now()}[0],
		})
	}

	return nil
}

func (h *Handler) HealthCheck(ctx echo.Context) error {
	// Проверяем здоровье микросервисов
	gatewayStatus := "healthy"
	fileStorageStatus := "unhealthy"
	fileAnalysisStatus := "unhealthy"

	// Проверка file-storage
	if _, err := h.fileStorageService.GetFileMetadata(ctx.Request().Context(), "health-check"); err == nil {
		fileStorageStatus = "healthy"
	}

	// Проверка file-analysis
	if _, err := h.fileAnalysisService.ListReports(ctx.Request().Context(), &models.ListReportsParams{Limit: intPtr(1)}); err == nil {
		fileAnalysisStatus = "healthy"
	}

	response := map[string]string{
		"status":        "OK",
		"gateway":       gatewayStatus,
		"file_storage":  fileStorageStatus,
		"file_analysis": fileAnalysisStatus,
	}

	return ctx.JSON(http.StatusOK, response)
}

// Вспомогательные функции
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func getFormValue(form *multipart.Form, key string) string {
	values := form.Value[key]
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
