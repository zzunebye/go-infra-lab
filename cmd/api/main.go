package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// alias for the package
	linkhttp "go-infra-lab/internal/http"
	"go-infra-lab/internal/link"
	"go-infra-lab/internal/storage"
)

func main() {
	port := envOr("PORT", "8080")
	baseURL := envOr("BASE_URL", "http://localhost:"+port)

	repository := storage.NewMemoryRepository()
	// "go-infra-lab/internal/link" package name as identifier
	service := link.NewService(repository)
	handler := linkhttp.NewHandler(service, baseURL)
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           linkhttp.NewRouter(handler),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// signal.NotifyContext(...)는 OS 신호를 context의 cancel 이벤트로 바꿔주는 함수
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown failed", "error", err)
		}
	}()

	slog.Info("starting URL shortener", "address", server.Addr, "base_url", baseURL)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
