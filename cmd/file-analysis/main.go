package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	fileanalysis "sd_hw3/api/generated/file-analysis"
	"sd_hw3/internal/file-analysis/handlers"
	"sd_hw3/internal/file-analysis/repository"
	"sd_hw3/internal/file-analysis/service"
	"sd_hw3/pkg/config"
	"sd_hw3/pkg/db"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	// "github.com/labstack/gommon/log"
)

func main() {
	cfg := config.Load()

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

	repo := repository.NewReportRepository()
	svc := service.NewAnalysisService(*cfg, repo)
	h := handlers.NewHandler(svc)

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

	// Routes
	fileanalysis.RegisterHandlers(e, h)

	e.GET("/health", func(c echo.Context) error {
		// Проверка соединения с БД
		if err := db.DB.Ping(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{
				"status":  "not working",
				"service": "file-analysis",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"status":  "OK",
			"service": "file-analysis",
			"version": "1.0.0",
		})
	})

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: e,
	}

	go func() {
		if err := e.StartServer(srv); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalf("shutting down the server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Errorf("error during shutdown: %v", err)
	}
}
