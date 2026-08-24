package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	applicationConfiguration, errorValue := configuration.Load()
	if errorValue != nil {
		slog.Error("The application configuration could not be loaded.", "error", errorValue)
		os.Exit(1)
	}

	databaseConnection, errorValue := database.NewPostgreSQLConnection(
		applicationConfiguration.PostgreSQLDataSource,
	)
	if errorValue != nil {
		slog.Error("The PostgreSQL connection could not be created.", "error", errorValue)
		os.Exit(1)
	}
	defer databaseConnection.Close()

	errorValue = database.ApplyMigrations(
		databaseConnection,
		database.MigrationDirectionUp,
	)
	if errorValue != nil {
		slog.Error("The database schema could not be migrated.", "error", errorValue)
		os.Exit(1)
	}

	currentTime := time.Now
	queries := databasequeries.New(databaseConnection)

	passwordHasher := security.NewBcryptPasswordHasher(bcrypt.DefaultCost)
	accessTokenManager := security.NewJSONWebTokenManager(
		applicationConfiguration.JSONWebTokenSecret,
		applicationConfiguration.JSONWebTokenIssuer,
		applicationConfiguration.JSONWebTokenLifetime,
	)

	userRepository := authenticationrepository.NewUserRepository(queries)
	productRepository := catalogrepository.NewProductRepository(queries)
	stockRepository := inventoryrepository.NewStockRepository(queries)
	orderRepository := orderrepository.NewOrderRepository(queries)
	sagaRepository := orderrepository.NewSagaRepository(queries)
	outboxRepository := outbox.NewSQLCRepository(queries)
	inboxRepository := inbox.NewSQLCRepository(queries)

	transactionManager := transaction.NewManager(databaseConnection)

	registerUserUseCase := authenticationusecase.NewRegisterUserUseCase(
		userRepository,
		passwordHasher,
		currentTime,
	)
	loginUserUseCase := authenticationusecase.NewLoginUserUseCase(
		userRepository,
		passwordHasher,
		accessTokenManager,
		currentTime,
	)
	getUserUseCase := authenticationusecase.NewGetUserUseCase(userRepository)
	listUsersUseCase := authenticationusecase.NewListUsersUseCase(userRepository)

	createProductUseCase := catalogusecase.NewCreateProductUseCase(
		productRepository,
		currentTime,
	)
	getProductUseCase := catalogusecase.NewGetProductUseCase(productRepository)
	listProductsUseCase := catalogusecase.NewListProductsUseCase(
		productRepository,
	)

	inventoryService := inventoryusecase.NewInventoryService(stockRepository, currentTime)

	checkoutUseCase := checkoutusecase.NewCheckoutUseCase(
		productRepository,
		userRepository,
		inventoryService,
		orderRepository,
		transactionManager,
		currentTime,
	)

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
		URL: applicationConfiguration.RabbitMQURL, CommandExchange: applicationConfiguration.RabbitMQCommandExchange,
		EventExchange: applicationConfiguration.RabbitMQEventExchange, PaymentQueue: applicationConfiguration.RabbitMQPaymentQueue,
		ResultQueue: applicationConfiguration.RabbitMQResultQueue, RetryLimit: applicationConfiguration.MessageRetryLimit,
	})
	if errorValue != nil {
		slog.Error("RabbitMQ could not be initialized.", "error", errorValue)
		os.Exit(1)
	}
	defer broker.Close()
	messagePublisher, errorValue := broker.NewPublisher()
	if errorValue != nil {
		slog.Error("RabbitMQ publisher could not be initialized.", "error", errorValue)
		os.Exit(1)
	}
	defer messagePublisher.Close()
	resultConsumer, errorValue := broker.NewConsumer(applicationConfiguration.RabbitMQResultQueue)
	if errorValue != nil {
		slog.Error("RabbitMQ consumer could not be initialized.", "error", errorValue)
		os.Exit(1)
	}
	defer resultConsumer.Close()
	outboxPublisher := outbox.NewPublisher(outboxRepository, messagePublisher, applicationConfiguration.OutboxInterval, applicationConfiguration.OutboxBatchSize, currentTime, slog.Default())

	router := newRouter(
		authenticationHandler,
		productHandler,
		inventoryHandler,
		checkoutHandler,
		orderHandler,
		accessTokenManager,
	)

	slog.Info("The HTTP server is running.", "address", ":"+applicationConfiguration.ApplicationPort)

	server := &http.Server{
		Addr:              ":" + applicationConfiguration.ApplicationPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	applicationContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		if errorValue := runWorkers(applicationContext, outboxPublisher, resultConsumer, paymentResultHandler); errorValue != nil {
			slog.Error("A background worker stopped.", "error", errorValue)
			stop()
		}
	}()
	go func() {
		<-applicationContext.Done()
		shutdownContext, cancel := context.WithTimeout(context.Background(), applicationConfiguration.ShutdownTimeout)
		defer cancel()
		_ = server.Shutdown(shutdownContext)
	}()

	if errorValue := server.ListenAndServe(); errorValue != nil && !errors.Is(errorValue, http.ErrServerClosed) {
		slog.Error("The HTTP server stopped unexpectedly.", "error", errorValue)
		os.Exit(1)
	}
}
