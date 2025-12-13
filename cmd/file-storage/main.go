package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	filestorage "sd_hw3/api/generated/file-storage"
	handler "sd_hw3/internal/file-storage/handlers"
	"sd_hw3/internal/file-storage/service"
	"sd_hw3/pkg/config"
	"sd_hw3/pkg/db"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Загрузка конфигурации
	cfg := loadConfig()

	// Подключение к БД
	log.Println("Connecting to database...")
	if err := db.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	// Выполнение миграций
	if cfg.RunMigrations {
		log.Println("Running database migrations...")
		if err := db.Migrate(cfg.MigrationsDir); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Migrations completed successfully")
	}

	// Инициализация сервисов
	storageService, err := service.NewStorageService(config.Config{
		UploadDir:     cfg.UploadDir,
		MaxUploadSize: cfg.MaxUploadSize,
		AllowedTypes:  cfg.AllowedTypes,
	})
	if err != nil {
		log.Fatalf("Failed to create storage service: %v", err)
	}

	// Инициализация Echo
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${status} ${method} ${uri} ${latency_human}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Лимит размера файла
	e.Use(middleware.BodyLimit(fmt.Sprintf("%dM", cfg.MaxUploadSize/(1024*1024))))

	// Создание обработчика
	fileHandler := handler.NewHandler(storageService)

	// Регистрация обработчиков
	filestorage.RegisterHandlers(e, fileHandler)

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		// Проверка соединения с БД
		if err := db.DB.Ping(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status":  "not working",
				"service": "file-storage",
				"error":   "database connection failed",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"status":  "OK",
			"service": "file-storage",
			"version": "1.0.0",
		})
	})

	// Запуск сервера
	port := cfg.ServerPort
	serverAddr := fmt.Sprintf(":%s", port)

	// Graceful shutdown
	go func() {
		log.Printf("Starting server on %s", serverAddr)
		log.Printf("Storage path: %s", cfg.UploadDir)
		log.Printf("Max file size: %d bytes", cfg.MaxUploadSize)

		if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("shutting down the server: %v", err)
		}
	}()

	// Ожидание сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Shutting down server...")
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server exited properly")
}

func loadConfig() config.Config {
	return config.Config{
		ServerPort:    getEnv("PORT", "8081"),
		DatabaseURL:   getEnv("DB_CONNECTION_STRING", "postgres://postgres:postgres@localhost:5432/filestorage?sslmode=disable"),
		UploadDir:     getEnv("STORAGE_PATH", "./uploads"),
		MaxUploadSize: parseEnvInt64("MAX_FILE_SIZE", 10*1024*1024), // 10MB
		AllowedTypes:  parseEnvStringSlice("ALLOWED_TYPES", []string{}),
		RunMigrations: parseEnvBool("RUN_MIGRATIONS", true),
		MigrationsDir: getEnv("MIGRATIONS_DIR", "./migrations"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		var result int64
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

func parseEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

func parseEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Простая реализация - разделение запятыми
		var result []string
		// Здесь должна быть логика парсинга массива строк
		return result
	}
	return defaultValue
}
