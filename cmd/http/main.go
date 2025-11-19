package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"amartha-test/config"
	httpHandler "amartha-test/internal/transport/http"
	"amartha-test/internal/transport/repository"
	"amartha-test/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.MySQL.User,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.DBName,
	)

	repo, err := repository.NewMySQLLoanRepository(dsn)
	if err != nil {
		slog.Error("Failed to initialize repository", "error", err)
		os.Exit(1)
	}

	service := usecase.NewBillingService(repo)

	handler := httpHandler.NewLoanHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("/loans/create", handler.CreateLoan)
	mux.HandleFunc("/loans/status", handler.GetLoanDetails)
	mux.HandleFunc("/loans/pay", handler.MakePayment)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := ":8080"
	httpServer := &http.Server{Addr: addr, Handler: mux}

	go func() {
		slog.Info("Starting HTTP server", "address", addr)
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("HTTP server crashed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server exiting")
}
