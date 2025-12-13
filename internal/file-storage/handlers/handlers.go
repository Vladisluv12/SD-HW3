package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"

	// "time"

	filestorage "sd_hw3/api/generated/file-storage"
	"sd_hw3/internal/file-storage/models"
	"sd_hw3/internal/file-storage/service"

	"github.com/labstack/echo/v4"
)

// StorageService интерфейс для работы с хранилищем файлов
type FileStorageService interface {
	UploadFile(studentID, assignmentID string, fileData []byte, filename, contentType string, size int64) (*models.File, *models.Work, error)
	GetFile(fileID string) (*models.File, error)
	GetFileMetadata(fileID string) (*models.File, error)
	GetFileContent(fileID string) ([]byte, error)
	CheckFileExists(fileID string) (bool, error)
}

// Handler реализует ServerInterface из сгенерированного кода
type Handler struct {
	service service.StorageService
}

// NewHandler создает новый обработчик
func NewHandler(service *service.StorageService) *Handler {
	return &Handler{
		service: *service,
	}
}

// UploadFile загружает файл
func (h *Handler) UploadFile(ctx echo.Context) error {
	fmt.Println("goyda!!!")
	// Парсим multipart форму
	form, err := ctx.MultipartForm()
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, filestorage.ApiError{
			Error:   stringPtr("INVALID_FORM"),
			Message: stringPtr("Invalid multipart form data"),
		})
	}

	// Получаем поля формы
	studentID := getFormValue(form, "student_id")
	assignmentID := getFormValue(form, "assignment_id")

	if studentID == "" || assignmentID == "" {
		return ctx.JSON(http.StatusBadRequest, filestorage.ApiError{
			Error:   stringPtr("MISSING_REQUIRED_FIELDS"),
			Message: stringPtr("student_id and assignment_id are required"),
		})
	}

	// Получаем файл
	files := form.File["file"]
	if len(files) == 0 {
		return ctx.JSON(http.StatusBadRequest, filestorage.ApiError{
			Error:   stringPtr("NO_FILE"),
			Message: stringPtr("No file uploaded"),
		})
	}

	fileHeader := files[0]
	if fileHeader.Size == 0 {
		return ctx.JSON(http.StatusBadRequest, filestorage.ApiError{
			Error:   stringPtr("EMPTY_FILE"),
			Message: stringPtr("File is empty"),
		})
	}

	// Читаем файл
	src, err := fileHeader.Open()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, filestorage.ApiError{
			Error:   stringPtr("FILE_READ_ERROR"),
			Message: stringPtr("Failed to read uploaded file"),
		})
	}
	defer src.Close()

	// Здесь должна быть реализация чтения файла
	// Для примера создаем пустой массив байт
	fileData := make([]byte, fileHeader.Size)

	// Вызываем сервис
	fileModel, workModel, err := h.service.UploadFile(
		ctx.Request().Context(),
		studentID,
		assignmentID,
		fileData,
		fileHeader.Filename,
		fileHeader.Header.Get("Content-Type"),
		fileHeader.Size,
	)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, filestorage.ApiError{
			Error:   stringPtr("UPLOAD_ERROR"),
			Message: stringPtr(fmt.Sprintf("Failed to upload file: %v", err)),
		})
	}

	// Конвертируем в DTO
	response := MapFileToUploadResponse(fileModel, workModel)

	return ctx.JSON(http.StatusCreated, response)
}

// GetFile скачивает файл
func (h *Handler) GetFile(ctx echo.Context, fileId string) error {
	file, err := h.service.GetFile(ctx.Request().Context(), fileId)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, filestorage.ApiError{
			Error:   stringPtr("FILE_NOT_FOUND"),
			Message: stringPtr(fmt.Sprintf("File with id %s not found", fileId)),
		})
	}

	content, err := h.service.GetFileContent(ctx.Request().Context(), fileId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, filestorage.ApiError{
			Error:   stringPtr("FILE_READ_ERROR"),
			Message: stringPtr("Failed to read file content"),
		})
	}

	// Устанавливаем заголовки
	ctx.Response().Header().Set("Content-Type", *file.ContentType)
	ctx.Response().Header().Set("Content-Length", strconv.FormatInt(file.SizeBytes, 10))
	ctx.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"", file.OriginalFilename))

	return ctx.Blob(http.StatusOK, *file.ContentType, content)
}

// GetFileMetadata получает метаданные файла
func (h *Handler) GetFileMetadata(ctx echo.Context, fileId string) error {
	file, err := h.service.GetFileMetadata(ctx.Request().Context(), fileId)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, filestorage.ApiError{
			Error:   stringPtr("FILE_NOT_FOUND"),
			Message: stringPtr(fmt.Sprintf("File with id %s not found", fileId)),
		})
	}

	metadata := MapFileMetaToMetadata(file)
	return ctx.JSON(http.StatusOK, metadata)
}

// GetFileContentInternal получает содержимое файла для внутреннего использования
func (h *Handler) GetFileContentInternal(ctx echo.Context, fileId string) error {
	file, err := h.service.GetFile(ctx.Request().Context(), fileId)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, filestorage.ApiError{
			Error:   stringPtr("FILE_NOT_FOUND"),
			Message: stringPtr(fmt.Sprintf("File with id %s not found", fileId)),
		})
	}

	content, err := h.service.GetFileContent(ctx.Request().Context(), fileId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, filestorage.ApiError{
			Error:   stringPtr("FILE_READ_ERROR"),
			Message: stringPtr("Failed to read file content"),
		})
	}

	ctx.Response().Header().Set("Content-Type", *file.ContentType)
	return ctx.Blob(http.StatusOK, *file.ContentType, content)
}

func (h *Handler) CheckFileExists(ctx echo.Context, fileId string) error {
	files, err := h.service.CheckFileExists(ctx.Request().Context(), fileId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, filestorage.ApiError{
			Error:   stringPtr("CHECK_FILE_ERROR"),
			Message: stringPtr("Failed to check file existence"),
		})
	}
	return ctx.JSON(http.StatusOK, map[string][]string{"files": files})
}

// Вспомогательные функции
func getFormValue(form *multipart.Form, key string) string {
	values := form.Value[key]
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func stringPtr(s string) *string {
	return &s
}
