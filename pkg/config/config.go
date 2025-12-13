package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort          string
	DatabaseURL         string
	UploadDir           string
	MaxUploadSize       int64
	AllowedTypes        []string
	RunMigrations       bool
	MigrationsDir       string
	PlagiarismThreshold float32
	EnableCaching       bool
	FileStorageURL      string
	FileAnalysisURL     string
}

func Load() *Config {
	return &Config{
		ServerPort:          getEnv("PORT", "8081"),
		DatabaseURL:         getEnv("DB_CONNECTION_STRING", "postgres://postgres:password@postgres:5432/file_storage?sslmode=disable"),
		UploadDir:           getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadSize:       parseInt64(getEnv("MAX_UPLOAD_SIZE", "10485760")),
		RunMigrations:       parseBool(getEnv("RUN_MIGRATIONS", "true")),
		MigrationsDir:       getEnv("MIGRATIONS_DIR", "/migrations"),
		PlagiarismThreshold: parseEnvFloat32("PLAGIARISM_THRESHOLD", 70.0),
		EnableCaching:       parseBool(getEnv("ENABLE_CACHING", "true")),
		FileStorageURL:      getEnv("FILE_STORAGE_URL", "http://localhost:8081"),
		FileAnalysisURL:     getEnv("FILE_ANALYSIS_URL", "http://localhost:8082"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt64(value string) int64 {
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return i
	}
	return 10485760 // 10MB
}

func parseBool(value string) bool {
	if i, err := strconv.ParseBool(value); err == nil {
		return i
	}
	return true
}

func parseEnvFloat32(value string, defaultValue float32) float32 {
	if value := os.Getenv(value); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatVal)
		}
	}
	return defaultValue
}
