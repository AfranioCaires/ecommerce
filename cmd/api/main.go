package main

import (
	"log/slog"
	"net/http"
	"os"
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
	orderrepository "github.com/afraniocaires/ecommerce/internal/order/adapter/repository/sqlc"
	orderusecase "github.com/afraniocaires/ecommerce/internal/order/usecase"
	paymentrepository "github.com/afraniocaires/ecommerce/internal/payment/adapter/repository/sqlc"
	paymentusecase "github.com/afraniocaires/ecommerce/internal/payment/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/configuration"
	"github.com/afraniocaires/ecommerce/internal/platform/database"
	databasequeries "github.com/afraniocaires/ecommerce/internal/platform/database/sqlc"
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
	paymentRepository := paymentrepository.NewPaymentRepository(queries)

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

	paymentGateway := paymentusecase.NewSimulatedPaymentGateway()
	paymentService := paymentusecase.NewPaymentService(
		paymentRepository,
		paymentGateway,
		currentTime,
	)

	checkoutUseCase := checkoutusecase.NewCheckoutUseCase(
		productRepository,
		inventoryService,
		orderRepository,
		paymentService,
		transactionManager,
		currentTime,
	)

	getOrderUseCase := orderusecase.NewGetOrderUseCase(orderRepository)
	listUserOrdersUseCase := orderusecase.NewListUserOrdersUseCase(orderRepository)
	listAllOrdersUseCase := orderusecase.NewListAllOrdersUseCase(orderRepository)

	authenticationHandler := authenticationtransport.NewHandler(registerUserUseCase, loginUserUseCase, getUserUseCase, listUsersUseCase)
	productHandler := catalogtransport.NewHandler(createProductUseCase, getProductUseCase, listProductsUseCase)
	inventoryHandler := inventorytransport.NewHandler(inventoryService)
	checkoutHandler := checkouttransport.NewHandler(checkoutUseCase)
	orderHandler := ordertransport.NewHandler(getOrderUseCase, listUserOrdersUseCase, listAllOrdersUseCase)

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

	if errorValue := server.ListenAndServe(); errorValue != nil {
		slog.Error("The HTTP server stopped unexpectedly.", "error", errorValue)
		os.Exit(1)
	}
}
