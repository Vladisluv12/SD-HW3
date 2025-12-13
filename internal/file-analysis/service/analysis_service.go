package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"sd_hw3/internal/file-analysis/models"
	"sd_hw3/internal/file-analysis/repository"
	"sd_hw3/pkg/config"
)

type FileStorageClient interface {
	GetFileContent(ctx context.Context, fileID string) ([]byte, error)
	GetFileMetadata(ctx context.Context, fileID string) (map[string]interface{}, error)
	GetSimilarWorks(ctx context.Context, fileID string) ([]string, error)
}

type AnalysisService interface {
	AnalyzeFile(ctx context.Context, req *models.AnalysisRequest) (*models.Report, error)
	GetReport(ctx context.Context, reportID string) (*models.Report, error)
	GetWorkReports(ctx context.Context, workID string) ([]*models.Report, error)
	ListReports(ctx context.Context, params repository.ListReportsParams) ([]*models.Report, int, error)
}

type analysisService struct {
	config            config.Config
	repo              repository.ReportRepository
	fileStorageClient FileStorageClient
	cache             map[string]*models.Report // простой in-memory кэш
}

func NewAnalysisService(cfg config.Config, repo repository.ReportRepository) AnalysisService {
	return &analysisService{
		config: cfg,
		repo:   repo,
		fileStorageClient: &httpFileStorageClient{
			baseURL: cfg.FileStorageURL,
		},
		cache: make(map[string]*models.Report),
	}
}

func (s *analysisService) AnalyzeFile(ctx context.Context, req *models.AnalysisRequest) (*models.Report, error) {
	startTime := time.Now()

	reportID := generateReportID(req.FileID)

	if cached, ok := s.cache[reportID]; ok && s.config.EnableCaching {
		return cached, nil
	}

	report := &models.Report{
		ReportID:     reportID,
		WorkID:       req.WorkID,
		FileID:       req.FileID,
		StudentID:    getStringValue(req.StudentID),
		AssignmentID: getStringValue(req.AssignmentID),
		Status:       "completed",
		CreatedAt:    time.Now(),
	}

	defer func() {
		report.AnalysisDurationMs = int(time.Since(startTime).Milliseconds())
	}()

	fileContent, err := s.fileStorageClient.GetFileContent(ctx, req.FileID)
	if err != nil {
		report.Status = "failed"
		errMsg := fmt.Sprintf("Failed to get file from storage: %v", err)
		report.ErrorMessage = &errMsg
		s.repo.CreateReport(ctx, report)
		return report, fmt.Errorf("failed to get file content: %w", err)
	}

	if int64(len(fileContent)) > s.config.MaxUploadSize {
		report.Status = "failed"
		errMsg := fmt.Sprintf("File too large: %d bytes (max: %d)", len(fileContent), s.config.MaxUploadSize)
		report.ErrorMessage = &errMsg
		s.repo.CreateReport(ctx, report)
		return report, fmt.Errorf("file too large")
	}

	text := string(fileContent)

	report.WordCount = countWords(text)

	report.PlagiarismScore = calculatePlagiarismScore(text)

	report.IsPlagiarism = report.PlagiarismScore > s.config.PlagiarismThreshold

	if report.PlagiarismScore > 50 {
		similarWorks, err := s.fileStorageClient.GetSimilarWorks(ctx, req.FileID)
		if err != nil {
			fmt.Printf("Failed to get similar works: %v\n", err)
		}
		for _, similar := range similarWorks {
			if err := s.repo.AddSimilarWork(ctx, models.MapFileIDToSimilarWorks(req.FileID, similar, reportID)); err != nil {
				fmt.Printf("Failed to add similar work: %v\n", err)
			}
		}
	}

	if err := s.repo.CreateReport(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	if s.config.EnableCaching {
		s.cache[reportID] = report
	}

	return report, nil
}

func (s *analysisService) GetReport(ctx context.Context, reportID string) (*models.Report, error) {
	// Проверяем кэш
	if cached, ok := s.cache[reportID]; ok && s.config.EnableCaching {
		return cached, nil
	}

	// Получаем из БД
	report, err := s.repo.GetReport(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	// Получаем похожие работы
	similarWorks, err := s.repo.GetSimilarWorks(ctx, reportID)
	if err != nil {
		fmt.Printf("Failed to get similar works: %v\n", err)
	}
	report.SimilarWorks = similarWorks
	if s.config.EnableCaching {
		s.cache[reportID] = report
	}

	return report, nil
}

func (s *analysisService) GetWorkReports(ctx context.Context, workID string) ([]*models.Report, error) {
	return s.repo.GetReportsByWorkID(ctx, workID)
}

func (s *analysisService) ListReports(ctx context.Context, params repository.ListReportsParams) ([]*models.Report, int, error) {
	return s.repo.ListReports(ctx, params)
}

// Вспомогательные функции
func generateReportID(fileID string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", fileID, time.Now().UnixNano())))
	return fmt.Sprintf("report-%s", hex.EncodeToString(hash[:8]))
}

func countWords(text string) int {
	// Удаляем лишние пробелы и символы
	cleaned := strings.TrimSpace(text)
	if cleaned == "" {
		return 0
	}

	// Простой подсчет слов
	re := regexp.MustCompile(`\s+`)
	words := re.Split(cleaned, -1)

	// Фильтруем пустые строки
	count := 0
	for _, word := range words {
		if strings.TrimSpace(word) != "" {
			count++
		}
	}
	return count
}

func CompareFiles(file1, file2 string) (bool, error) {
	f1, err := os.Open(file1)
	if err != nil {
		return false, fmt.Errorf("ошибка открытия файла %s: %w", file1, err)
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return false, fmt.Errorf("ошибка открытия файла %s: %w", file2, err)
	}
	defer f2.Close()

	// Сравниваем размеры файлов
	stat1, err := f1.Stat()
	if err != nil {
		return false, err
	}

	stat2, err := f2.Stat()
	if err != nil {
		return false, err
	}

	if stat1.Size() != stat2.Size() {
		return false, nil
	}

	// Используем буферизованное сравнение
	return compareReaders(f1, f2), nil
}

// compareReaders сравнивает два io.Reader
func compareReaders(r1, r2 io.Reader) bool {
	const chunkSize = 64 * 1024 // 64KB

	buf1 := make([]byte, chunkSize)
	buf2 := make([]byte, chunkSize)

	for {
		n1, err1 := io.ReadFull(r1, buf1)
		n2, err2 := io.ReadFull(r2, buf2)

		if n1 != n2 || !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false
		}

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				return true
			}
			if err1 == io.ErrUnexpectedEOF && err2 == io.ErrUnexpectedEOF {
				return true
			}
			return false
		}
	}
}

func getStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// HTTP клиент для file-storage сервиса
type httpFileStorageClient struct {
	baseURL string
}

func (c *httpFileStorageClient) GetFileContent(ctx context.Context, fileID string) ([]byte, error) {
	url := fmt.Sprintf("%s/internal/files/%s/content", c.baseURL, fileID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file storage returned status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (c *httpFileStorageClient) GetFileMetadata(ctx context.Context, fileID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/files/%s/metadata", c.baseURL, fileID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file storage returned status: %d", resp.StatusCode)
	}

	var metadata map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return metadata, nil
}

func (c *httpFileStorageClient) GetSimilarWorks(ctx context.Context, fileID string) ([]string, error) {
	url := fmt.Sprintf("%s/files/%s/exists", c.baseURL, fileID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get similar works: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file storage returned status: %d", resp.StatusCode)
	}
	var result struct {
		Files []string `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode similar works: %w", err)
	}
	return result.Files, nil
}

func calculatePlagiarismScore(text string) float32 {
	// Заглушка: в реальном приложении была бы сложная логика анализа
	return float32(100.00)
}
