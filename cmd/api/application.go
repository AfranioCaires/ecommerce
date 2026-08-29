package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	authenticationtransport "github.com/afraniocaires/ecommerce/internal/authentication/adapter/http"
	authenticationrepository "github.com/afraniocaires/ecommerce/internal/authentication/adapter/repository/sqlc"
	authenticationusecase "github.com/afraniocaires/ecommerce/internal/authentication/usecase"
	catalogtransport "github.com/afraniocaires/ecommerce/internal/catalog/adapter/http"
	catalogrepository "github.com/afraniocaires/ecommerce/internal/catalog/adapter/repository/sqlc"
	catalogusecase "github.com/afraniocaires/ecommerce/internal/catalog/usecase"
	checkouttransport "github.com/afraniocaires/ecommerce/internal/checkout/adapter/http"
	checkoutusecase "github.com/afraniocaires/ecommerce/internal/checkout/usecase"
	inventorytransport "github.com/afraniocaires/ecommerce/internal/inventory/adapter/http"
	inventoryrepository "github.com/afraniocaires/ecommerce/internal/inventory/adapter/repository/sqlc"
	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	ordertransport "github.com/afraniocaires/ecommerce/internal/order/adapter/http"
	ordermessaging "github.com/afraniocaires/ecommerce/internal/order/adapter/messaging"
	orderrepository "github.com/afraniocaires/ecommerce/internal/order/adapter/repository/sqlc"
	orderusecase "github.com/afraniocaires/ecommerce/internal/order/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/configuration"
	"github.com/afraniocaires/ecommerce/internal/platform/database"
	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
	"github.com/afraniocaires/ecommerce/internal/platform/inbox"
	platformmessaging "github.com/afraniocaires/ecommerce/internal/platform/messaging"
	"github.com/afraniocaires/ecommerce/internal/platform/outbox"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
	"github.com/afraniocaires/ecommerce/internal/platform/transaction"
)

type application struct {
	server          *http.Server
	publisher       *outbox.Publisher
	consumer        *platformmessaging.Consumer
	handler         *ordermessaging.PaymentResultHandler
	shutdownTimeout time.Duration
	close           func() error
}

func newApplication(applicationConfiguration *configuration.Configuration, logger *slog.Logger) (*application, error) {
	databaseConnection, errorValue := database.NewPostgreSQLConnection(applicationConfiguration.PostgreSQLDataSource)
	if errorValue != nil {
		return nil, fmt.Errorf("open ecommerce database: %w", errorValue)
	}
	closeDatabase := true
	defer func() {
		if closeDatabase {
			_ = databaseConnection.Close()
		}
	}()
	if errorValue := database.ApplyMigrations(databaseConnection, database.MigrationDirectionUp); errorValue != nil {
		return nil, errorValue
	}

	currentTime := time.Now
	queries := databasequeries.New(databaseConnection)
	passwordHasher := security.NewBcryptPasswordHasher(bcrypt.DefaultCost)
	accessTokenManager := security.NewJSONWebTokenManager(applicationConfiguration.JSONWebTokenSecret, applicationConfiguration.JSONWebTokenIssuer, applicationConfiguration.JSONWebTokenLifetime)
	userRepository := authenticationrepository.NewUserRepository(queries)
	productRepository := catalogrepository.NewProductRepository(queries)
	stockRepository := inventoryrepository.NewStockRepository(queries)
	orderRepository := orderrepository.NewOrderRepository(queries)
	sagaRepository := orderrepository.NewSagaRepository(queries)
	outboxRepository := outbox.NewSQLCRepository(queries)
	inboxRepository := inbox.NewSQLCRepository(queries)
	transactionManager := transaction.NewManager(databaseConnection)

	registerUserUseCase := authenticationusecase.NewRegisterUserUseCase(userRepository, passwordHasher, currentTime)
	loginUserUseCase := authenticationusecase.NewLoginUserUseCase(userRepository, passwordHasher, accessTokenManager, currentTime)
	getUserUseCase := authenticationusecase.NewGetUserUseCase(userRepository)
	listUsersUseCase := authenticationusecase.NewListUsersUseCase(userRepository)
	createProductUseCase := catalogusecase.NewCreateProductUseCase(productRepository, currentTime)
	getProductUseCase := catalogusecase.NewGetProductUseCase(productRepository)
	listProductsUseCase := catalogusecase.NewListProductsUseCase(productRepository)
	inventoryService := inventoryusecase.NewInventoryService(stockRepository, currentTime)
	checkoutUseCase := checkoutusecase.NewCheckoutUseCase(productRepository, userRepository, inventoryService, orderRepository, transactionManager, currentTime)
	getOrderUseCase := orderusecase.NewGetOrderUseCase(orderRepository)
	listUserOrdersUseCase := orderusecase.NewListUserOrdersUseCase(orderRepository)
	listAllOrdersUseCase := orderusecase.NewListAllOrdersUseCase(orderRepository)
	cancelOrderUseCase := orderusecase.NewCancelOrderUseCase(orderRepository, inventoryService, transactionManager, currentTime)
	payOrderUseCase := orderusecase.NewPayOrderUseCase(orderRepository, sagaRepository, outboxRepository, transactionManager, currentTime)
	handlePaymentResultUseCase := orderusecase.NewHandlePaymentResultUseCase(orderRepository, sagaRepository, inboxRepository, inventoryService, transactionManager, currentTime)

	authenticationHandler := authenticationtransport.NewHandler(registerUserUseCase, loginUserUseCase, getUserUseCase, listUsersUseCase)
	productHandler := catalogtransport.NewHandler(createProductUseCase, getProductUseCase, listProductsUseCase)
	inventoryHandler := inventorytransport.NewHandler(inventoryService)
	checkoutHandler := checkouttransport.NewHandler(checkoutUseCase)
	orderHandler := ordertransport.NewHandler(getOrderUseCase, listUserOrdersUseCase, listAllOrdersUseCase, cancelOrderUseCase, payOrderUseCase)
	paymentResultHandler := ordermessaging.NewPaymentResultHandler(handlePaymentResultUseCase)

	broker, errorValue := platformmessaging.Dial(platformmessaging.Config{
		URL:             applicationConfiguration.RabbitMQURL,
		CommandExchange: applicationConfiguration.RabbitMQCommandExchange,
		EventExchange:   applicationConfiguration.RabbitMQEventExchange,
		PaymentQueue:    applicationConfiguration.RabbitMQPaymentQueue,
		ResultQueue:     applicationConfiguration.RabbitMQResultQueue,
		RetryLimit:      applicationConfiguration.MessageRetryLimit,
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
	resultConsumer, errorValue := broker.NewConsumer(applicationConfiguration.RabbitMQResultQueue)
	if errorValue != nil {
		return nil, errorValue
	}
	outboxPublisher := outbox.NewPublisher(outboxRepository, messagePublisher, applicationConfiguration.OutboxInterval, applicationConfiguration.OutboxBatchSize, currentTime, logger)
	router := newRouter(authenticationHandler, productHandler, inventoryHandler, checkoutHandler, orderHandler, accessTokenManager)
	server := &http.Server{
		Addr:              ":" + applicationConfiguration.ApplicationPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	closeDatabase, closeBroker, closePublisher = false, false, false
	return &application{
		server: server, publisher: outboxPublisher, consumer: resultConsumer, handler: paymentResultHandler,
		shutdownTimeout: applicationConfiguration.ShutdownTimeout,
		close: func() error {
			return errors.Join(resultConsumer.Close(), messagePublisher.Close(), broker.Close(), databaseConnection.Close())
		},
	}, nil
}

func (application *application) Run(applicationContext context.Context) error {
	runContext, cancel := context.WithCancel(applicationContext)
	defer cancel()
	errorsChannel := make(chan error, 2)
	go func() { errorsChannel <- application.server.ListenAndServe() }()
	go func() {
		errorsChannel <- runWorkers(runContext, application.publisher, application.consumer, application.handler)
	}()

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
