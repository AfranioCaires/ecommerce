package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	paymentmessaging "github.com/afraniocaires/ecommerce/internal/payment/adapter/messaging"
	paymentrepository "github.com/afraniocaires/ecommerce/internal/payment/adapter/repository/sqlc"
	paymentconfiguration "github.com/afraniocaires/ecommerce/internal/payment/platform/configuration"
	paymentdatabase "github.com/afraniocaires/ecommerce/internal/payment/platform/database"
	paymentqueries "github.com/afraniocaires/ecommerce/internal/payment/platform/database/sqlc"
	paymenttransaction "github.com/afraniocaires/ecommerce/internal/payment/platform/transaction"
	paymentusecase "github.com/afraniocaires/ecommerce/internal/payment/usecase"
	platformdatabase "github.com/afraniocaires/ecommerce/internal/platform/database"
	platformmessaging "github.com/afraniocaires/ecommerce/internal/platform/messaging"
	"github.com/afraniocaires/ecommerce/internal/platform/outbox"
)

type application struct {
	server          *http.Server
	publisher       *outbox.Publisher
	consumer        *platformmessaging.Consumer
	handler         *paymentmessaging.PaymentRequestHandler
	shutdownTimeout time.Duration
	close           func() error
}

func newApplication(configuration *paymentconfiguration.Configuration, logger *slog.Logger) (*application, error) {
	databaseConnection, errorValue := platformdatabase.NewPostgreSQLConnection(configuration.PostgreSQLDataSource)
	if errorValue != nil {
		return nil, fmt.Errorf("open payment database: %w", errorValue)
	}
	closeDatabase := true
	defer func() {
		if closeDatabase {
			_ = databaseConnection.Close()
		}
	}()
	if errorValue := paymentdatabase.ApplyMigrations(databaseConnection); errorValue != nil {
		return nil, errorValue
	}

	broker, errorValue := platformmessaging.Dial(platformmessaging.Config{
		URL:             configuration.RabbitMQURL,
		CommandExchange: configuration.RabbitMQCommandExchange,
		EventExchange:   configuration.RabbitMQEventExchange,
		PaymentQueue:    configuration.RabbitMQPaymentQueue,
		ResultQueue:     configuration.RabbitMQResultQueue,
		RetryLimit:      configuration.MessageRetryLimit,
	})
	if errorValue != nil {
		return nil, errorValue
	}
	closeBroker := true
	defer func() {
		if closeBroker {
			_ = broker.Close()
		}
	}()

	messagePublisher, errorValue := broker.NewPublisher()
	if errorValue != nil {
		return nil, errorValue
	}
	closePublisher := true
	defer func() {
		if closePublisher {
			_ = messagePublisher.Close()
		}
	}()
	paymentConsumer, errorValue := broker.NewConsumer(configuration.RabbitMQPaymentQueue)
	if errorValue != nil {
		return nil, errorValue
	}

	queries := paymentqueries.New(databaseConnection)
	paymentRepository := paymentrepository.NewPaymentRepository(queries)
	messageRepository := paymentrepository.NewMessageRepository(queries)
	transactionManager := paymenttransaction.NewManager(databaseConnection)
	gateway := paymentusecase.NewSimulatedPaymentGateway()
	service := paymentusecase.NewMessagePaymentService(paymentRepository, gateway, messageRepository, messageRepository, transactionManager, time.Now)
	handler := paymentmessaging.NewPaymentRequestHandler(service, logger)
	outboxPublisher := outbox.NewPublisher(messageRepository, messagePublisher, configuration.OutboxInterval, configuration.OutboxBatchSize, time.Now, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	server := &http.Server{
		Addr:              ":" + configuration.ApplicationPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	closeDatabase, closeBroker, closePublisher = false, false, false
	return &application{
		server: server, publisher: outboxPublisher, consumer: paymentConsumer, handler: handler,
		shutdownTimeout: configuration.ShutdownTimeout,
		close: func() error {
			return errors.Join(paymentConsumer.Close(), messagePublisher.Close(), broker.Close(), databaseConnection.Close())
		},
	}, nil
}

func healthHandler(responseWriter http.ResponseWriter, _ *http.Request) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)
	_, _ = responseWriter.Write([]byte(`{"status":"ok","service":"payment-service"}`))
}

func (application *application) Run(applicationContext context.Context) error {
	runContext, cancel := context.WithCancel(applicationContext)
	defer cancel()
	errorsChannel := make(chan error, 3)
	go func() { errorsChannel <- application.server.ListenAndServe() }()
	go func() { errorsChannel <- application.publisher.Run(runContext) }()
	go func() { errorsChannel <- application.consumer.Run(runContext, application.handler.Handle) }()

	var runError error
	select {
	case <-applicationContext.Done():
	case runError = <-errorsChannel:
		if errors.Is(runError, http.ErrServerClosed) || errors.Is(runError, context.Canceled) {
			runError = nil
		}
	}
	cancel()
	shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), application.shutdownTimeout)
	defer shutdownCancel()
	if shutdownError := application.server.Shutdown(shutdownContext); runError == nil && shutdownError != nil {
		runError = shutdownError
	}
	return runError
}

func (application *application) Close() error {
	if application.close == nil {
		return nil
	}
	return application.close()
}
