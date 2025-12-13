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

	"sd_hw3/api/generated/gateway"
	"sd_hw3/internal/gateway/handlers"
	"sd_hw3/internal/gateway/service"
	"sd_hw3/pkg/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Создание клиентов для микросервисов
	fileStorageService := service.NewFileStorageService(cfg.FileStorageURL)
	fileAnalysisService := service.NewFileAnalysisService(cfg.FileAnalysisURL)

	// Создание обработчика
	handler := handlers.NewHandler(fileStorageService, fileAnalysisService)

	// Создание Echo роутера
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Регистрация маршрутов
	gateway.RegisterHandlers(e, handler)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "OK",
			"service": "api gateway",
			"version": "1.0.0",
		})
	})

	// Запуск сервера
	port := cfg.ServerPort
	serverAddr := fmt.Sprintf(":%s", port)
	if port == "" {
		port = "8080"
	}

	log.Printf("Gateway starting on port %s", port)
	if err := e.Start(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

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
