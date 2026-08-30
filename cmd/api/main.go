package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/afraniocaires/ecommerce/internal/platform/configuration"
)

// @title Mini E-commerce API
// @version 1.0
// @description Modular e-commerce API for authentication, catalog, inventory, checkout, and orders.
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the access token using the Bearer scheme: Bearer {token}
func main() {
	os.Exit(run())
}

func run() int {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "ecommerce-api")
	slog.SetDefault(logger)

	applicationConfiguration, errorValue := configuration.Load()
	if errorValue != nil {
		logger.Error("The application configuration could not be loaded.", "operation", "application.start", "result", "failed", "error", errorValue)
		return 1
	}
	application, errorValue := newApplication(applicationConfiguration, logger)
	if errorValue != nil {
		logger.Error("The application could not be initialized.", "operation", "application.start", "result", "failed", "error", errorValue)
		return 1
	}
	defer func() {
		if errorValue := application.Close(); errorValue != nil {
			logger.Warn("Application resources did not close cleanly.", "operation", "application.stop", "result", "failed", "error", errorValue)
		}
	}()

	applicationContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	logger.Info("The HTTP server is running.", "operation", "application.start", "result", "success", "address", application.server.Addr)
	if errorValue := application.Run(applicationContext); errorValue != nil {
		logger.Error("The application stopped unexpectedly.", "operation", "application.stop", "result", "failed", "error", errorValue)
		return 1
	}
	logger.Info("The application stopped.", "operation", "application.stop", "result", "success")
	return 0
}
