package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	authenticationtransport "github.com/afraniocaires/ecommerce/internal/authentication/adapter/http"
	authenticationrepository "github.com/afraniocaires/ecommerce/internal/authentication/adapter/repository/memory"
	authenticationdomain "github.com/afraniocaires/ecommerce/internal/authentication/domain"
	authenticationusecase "github.com/afraniocaires/ecommerce/internal/authentication/usecase"
	catalogtransport "github.com/afraniocaires/ecommerce/internal/catalog/adapter/http"
	catalogrepository "github.com/afraniocaires/ecommerce/internal/catalog/adapter/repository/memory"
	catalogusecase "github.com/afraniocaires/ecommerce/internal/catalog/usecase"
	checkouttransport "github.com/afraniocaires/ecommerce/internal/checkout/adapter/http"
	checkoutusecase "github.com/afraniocaires/ecommerce/internal/checkout/usecase"
	inventorytransport "github.com/afraniocaires/ecommerce/internal/inventory/adapter/http"
	inventorydomain "github.com/afraniocaires/ecommerce/internal/inventory/domain"
	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	ordertransport "github.com/afraniocaires/ecommerce/internal/order/adapter/http"
	orderdomain "github.com/afraniocaires/ecommerce/internal/order/domain"
	orderusecase "github.com/afraniocaires/ecommerce/internal/order/usecase"
	paymentdomain "github.com/afraniocaires/ecommerce/internal/payment/domain"
	paymentusecase "github.com/afraniocaires/ecommerce/internal/payment/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

type applicationStockRepository struct {
	mutex  sync.Mutex
	stocks map[string]*inventorydomain.Stock
}

func (repository *applicationStockRepository) Save(applicationContext context.Context, stock *inventorydomain.Stock) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()
	repository.stocks[stock.ProductID] = stock
	return nil
}

func (repository *applicationStockRepository) FindByProductID(applicationContext context.Context, productID string) (*inventorydomain.Stock, error) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()
	stock, available := repository.stocks[productID]
	if !available {
		return nil, inventorydomain.ErrStockNotFound
	}
	return stock, nil
}

func (repository *applicationStockRepository) FindByProductIDForUpdate(applicationContext context.Context, productID string) (*inventorydomain.Stock, error) {
	return repository.FindByProductID(applicationContext, productID)
}

type applicationOrderRepository struct {
	orders map[string]*orderdomain.Order
}

func (repository *applicationOrderRepository) Save(applicationContext context.Context, order *orderdomain.Order) error {
	repository.orders[order.ID] = order
	return nil
}
func (repository *applicationOrderRepository) UpdateStatus(applicationContext context.Context, order *orderdomain.Order) error {
	repository.orders[order.ID] = order
	return nil
}
func (repository *applicationOrderRepository) FindByID(applicationContext context.Context, orderID string) (*orderdomain.Order, error) {
	order, available := repository.orders[orderID]
	if !available {
		return nil, orderdomain.ErrOrderNotFound
	}
	return order, nil
}
func (repository *applicationOrderRepository) FindByUserID(applicationContext context.Context, userID string, pageRequest orderusecase.OrderPageRequest) ([]*orderdomain.Order, error) {
	orders := make([]*orderdomain.Order, 0)
	for _, order := range repository.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return orders, nil
}
func (repository *applicationOrderRepository) FindAll(applicationContext context.Context, pageRequest orderusecase.OrderPageRequest) ([]*orderdomain.Order, error) {
	orders := make([]*orderdomain.Order, 0, len(repository.orders))
	for _, order := range repository.orders {
		orders = append(orders, order)
	}
	return orders, nil
}

type applicationPaymentRepository struct{}

func (repository applicationPaymentRepository) Save(applicationContext context.Context, payment *paymentdomain.Payment) error {
	return nil
}

type applicationTransactionManager struct{}

func (manager applicationTransactionManager) Execute(applicationContext context.Context, operation func(transactionContext context.Context) error) error {
	return operation(applicationContext)
}

func TestApplicationFlow(t *testing.T) {
	currentTime := time.Now
	userRepository := authenticationrepository.NewUserRepository()
	productRepository := catalogrepository.NewProductRepository()
	stockRepository := &applicationStockRepository{stocks: make(map[string]*inventorydomain.Stock)}
	orderRepository := &applicationOrderRepository{orders: make(map[string]*orderdomain.Order)}
	passwordHasher := security.NewBcryptPasswordHasher(bcrypt.MinCost)
	accessTokenManager := security.NewJSONWebTokenManager("secret", "ecommerce", time.Hour)
	registerUserUseCase := authenticationusecase.NewRegisterUserUseCase(userRepository, passwordHasher, currentTime)
	loginUserUseCase := authenticationusecase.NewLoginUserUseCase(userRepository, passwordHasher, accessTokenManager, currentTime)
	createProductUseCase := catalogusecase.NewCreateProductUseCase(productRepository, currentTime)
	getProductUseCase := catalogusecase.NewGetProductUseCase(productRepository)
	listProductsUseCase := catalogusecase.NewListProductsUseCase(productRepository)
	inventoryService := inventoryusecase.NewInventoryService(stockRepository, currentTime)
	paymentService := paymentusecase.NewPaymentService(applicationPaymentRepository{}, paymentusecase.NewSimulatedPaymentGateway(), currentTime)
	checkoutUseCase := checkoutusecase.NewCheckoutUseCase(productRepository, inventoryService, orderRepository, paymentService, applicationTransactionManager{}, currentTime)
	authenticationHandler := authenticationtransport.NewHandler(registerUserUseCase, loginUserUseCase)
	productHandler := catalogtransport.NewHandler(createProductUseCase, getProductUseCase, listProductsUseCase)
	inventoryHandler := inventorytransport.NewHandler(inventoryService)
	checkoutHandler := checkouttransport.NewHandler(checkoutUseCase)
	orderHandler := ordertransport.NewHandler(orderusecase.NewGetOrderUseCase(orderRepository), orderusecase.NewListUserOrdersUseCase(orderRepository), orderusecase.NewListAllOrdersUseCase(orderRepository))
	router := newRouter(authenticationHandler, productHandler, inventoryHandler, checkoutHandler, orderHandler, accessTokenManager)

	administratorToken, _ := accessTokenManager.Generate("administrator-1", []authenticationdomain.Role{authenticationdomain.RoleAdministrator}, time.Now())
	registerResponse := applicationJSONRequest(router, http.MethodPost, "/api/authentication/register", "", map[string]any{"email": "customer@example.com", "password": "password"})
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("registration failed: %d, %s", registerResponse.Code, registerResponse.Body.String())
	}
	loginResponse := applicationJSONRequest(router, http.MethodPost, "/api/authentication/login", "", map[string]any{"email": "customer@example.com", "password": "password"})
	var loginBody struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(loginResponse.Body.Bytes(), &loginBody)
	productResponse := applicationJSONRequest(router, http.MethodPost, "/api/products", administratorToken, map[string]any{"name": "Keyboard", "description": "Mechanical", "price_cents": 10000})
	var productBody struct {
		ID string `json:"id"`
	}
	json.Unmarshal(productResponse.Body.Bytes(), &productBody)
	stockResponse := applicationJSONRequest(router, http.MethodPut, "/api/inventory/"+productBody.ID, administratorToken, map[string]any{"quantity": 5})
	if stockResponse.Code != http.StatusOK {
		t.Fatalf("inventory failed: %d, %s", stockResponse.Code, stockResponse.Body.String())
	}
	checkoutResponse := applicationJSONRequest(router, http.MethodPost, "/api/orders", loginBody.AccessToken, map[string]any{"items": []map[string]any{{"product_id": productBody.ID, "quantity": 2}}})
	if checkoutResponse.Code != http.StatusCreated {
		t.Fatalf("checkout failed: %d, %s", checkoutResponse.Code, checkoutResponse.Body.String())
	}
	var checkoutBody struct {
		OrderID string `json:"order_id"`
	}
	json.Unmarshal(checkoutResponse.Body.Bytes(), &checkoutBody)
	request := httptest.NewRequest(http.MethodGet, "/api/orders/"+checkoutBody.OrderID, nil)
	request.Header.Set("Authorization", "Bearer "+loginBody.AccessToken)
	orderResponse := httptest.NewRecorder()
	router.ServeHTTP(orderResponse, request)
	if orderResponse.Code != http.StatusOK {
		t.Fatalf("order retrieval failed: %d, %s", orderResponse.Code, orderResponse.Body.String())
	}
}

func applicationJSONRequest(router http.Handler, method string, path string, accessToken string, value any) *httptest.ResponseRecorder {
	requestBody, _ := json.Marshal(value)
	request := httptest.NewRequest(method, path, bytes.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	return responseRecorder
}
