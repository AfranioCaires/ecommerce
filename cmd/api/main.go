package main

import (
	"log/slog"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	authenticationtransport "github.com/afraniocaires/ecommerce/internal/authentication/adapter/http"
	"github.com/afraniocaires/ecommerce/internal/authentication/adapter/repository/gorm"
	authenticationusecase "github.com/afraniocaires/ecommerce/internal/authentication/usecase"
	catalogtransport "github.com/afraniocaires/ecommerce/internal/catalog/adapter/http"
	"github.com/afraniocaires/ecommerce/internal/catalog/adapter/repository/gorm"
	catalogusecase "github.com/afraniocaires/ecommerce/internal/catalog/usecase"
	checkouttransport "github.com/afraniocaires/ecommerce/internal/checkout/adapter/http"
	checkoutusecase "github.com/afraniocaires/ecommerce/internal/checkout/usecase"
	inventorytransport "github.com/afraniocaires/ecommerce/internal/inventory/adapter/http"
	"github.com/afraniocaires/ecommerce/internal/inventory/adapter/repository/gorm"
	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	ordertransport "github.com/afraniocaires/ecommerce/internal/order/adapter/http"
	"github.com/afraniocaires/ecommerce/internal/order/adapter/repository/gorm"
	orderusecase "github.com/afraniocaires/ecommerce/internal/order/usecase"
	"github.com/afraniocaires/ecommerce/internal/payment/adapter/repository/gorm"
	paymentusecase "github.com/afraniocaires/ecommerce/internal/payment/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/configuration"
	"github.com/afraniocaires/ecommerce/internal/platform/database"
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

	errorValue = databaseConnection.AutoMigrate(
		&authenticationrepository.UserModel{},
		&catalogrepository.ProductModel{},
		&inventoryrepository.StockModel{},
		&orderrepository.OrderModel{},
		&orderrepository.OrderItemModel{},
		&paymentrepository.PaymentModel{},
	)
	if errorValue != nil {
		slog.Error("The database schema could not be migrated.", "error", errorValue)
		os.Exit(1)
	}

	currentTime := time.Now

	passwordHasher := security.NewBcryptPasswordHasher(bcrypt.DefaultCost)
	accessTokenManager := security.NewJSONWebTokenManager(
		applicationConfiguration.JSONWebTokenSecret,
		applicationConfiguration.JSONWebTokenIssuer,
		applicationConfiguration.JSONWebTokenLifetime,
	)

	userRepository := authenticationrepository.NewUserRepository(databaseConnection)
	productRepository := catalogrepository.NewProductRepository(databaseConnection)
	stockRepository := inventoryrepository.NewStockRepository(databaseConnection)
	orderRepository := orderrepository.NewOrderRepository(databaseConnection)
	paymentRepository := paymentrepository.NewPaymentRepository(databaseConnection)

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

	authenticationHandler := authenticationtransport.NewHandler(registerUserUseCase, loginUserUseCase)
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

	if errorValue := router.Run(":" + applicationConfiguration.ApplicationPort); errorValue != nil {
		slog.Error("The HTTP server stopped unexpectedly.", "error", errorValue)
		os.Exit(1)
	}
}
