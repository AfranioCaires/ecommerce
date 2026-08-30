package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	paymentconfiguration "github.com/afraniocaires/ecommerce/internal/payment/platform/configuration"
)

func main() {
	os.Exit(run())
}

func run() int {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "payment-service")
	slog.SetDefault(logger)

	configuration, errorValue := paymentconfiguration.Load()
	if errorValue != nil {
		logger.Error("Payment service configuration could not be loaded.", "operation", "application.start", "result", "failed", "error", errorValue)
		return 1
	}
	application, errorValue := newApplication(configuration, logger)
	if errorValue != nil {
		logger.Error("Payment service could not be initialized.", "operation", "application.start", "result", "failed", "error", errorValue)
		return 1
	}
	defer func() {
		if errorValue := application.Close(); errorValue != nil {
			logger.Warn("Payment service resources did not close cleanly.", "operation", "application.stop", "result", "failed", "error", errorValue)
		}
	}()

	applicationContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	logger.Info("Payment service is running.", "operation", "application.start", "result", "success", "address", application.server.Addr)
	if errorValue := application.Run(applicationContext); errorValue != nil {
		logger.Error("Payment service stopped unexpectedly.", "operation", "application.stop", "result", "failed", "error", errorValue)
		return 1
	}
	logger.Info("Payment service stopped.", "operation", "application.stop", "result", "success")
	return 0
}
