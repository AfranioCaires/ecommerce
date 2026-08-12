package authenticationtransport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/afraniocaires/ecommerce/internal/authentication/adapter/http/dto"
	authenticationrepository "github.com/afraniocaires/ecommerce/internal/authentication/adapter/repository/memory"
	"github.com/afraniocaires/ecommerce/internal/authentication/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

func TestHandler(t *testing.T) {
	userRepository := authenticationrepository.NewUserRepository()
	passwordHasher := security.NewBcryptPasswordHasher(bcrypt.MinCost)
	accessTokenManager := security.NewJSONWebTokenManager("secret", "ecommerce", time.Hour)
	registerUserUseCase := usecase.NewRegisterUserUseCase(userRepository, passwordHasher, time.Now)
	loginUserUseCase := usecase.NewLoginUserUseCase(userRepository, passwordHasher, accessTokenManager, time.Now)
	getUserUseCase := usecase.NewGetUserUseCase(userRepository)
	listUsersUseCase := usecase.NewListUsersUseCase(userRepository)
	handler := NewHandler(registerUserUseCase, loginUserUseCase, getUserUseCase, listUsersUseCase)
	router := http.NewServeMux()
	router.HandleFunc("POST /register", handler.Register)
	router.HandleFunc("POST /login", handler.Login)
	router.HandleFunc("POST /customers", handler.CreateCustomer)
	router.HandleFunc("GET /customers", handler.ListCustomers)
	router.HandleFunc("GET /customers/{customerID}", handler.GetCustomer)

	t.Run("it should register a customer", func(t *testing.T) {
		responseRecorder := performJSONRequest(router, http.MethodPost, "/register", dto.CredentialsRequest{Email: "customer@example.com", Password: "password"})
		if responseRecorder.Code != http.StatusCreated {
			t.Fatalf("expected created, received %d and %s", responseRecorder.Code, responseRecorder.Body.String())
		}
		var response dto.UserResponse
		if errorValue := json.Unmarshal(responseRecorder.Body.Bytes(), &response); errorValue != nil || response.Email != "customer@example.com" {
			t.Fatalf("unexpected response: %#v, %v", response, errorValue)
		}
	})

	t.Run("it should create list and get challenge customers without exposing passwords", func(t *testing.T) {
		created := performJSONRequest(router, http.MethodPost, "/customers", dto.CreateCustomerRequest{
			Name: "Ada Lovelace", Email: "ada@example.com", PasswordHash: "password",
		})
		if created.Code != http.StatusCreated || strings.Contains(created.Body.String(), "password") {
			t.Fatalf("unexpected create response: %d %s", created.Code, created.Body.String())
		}
		var customer dto.UserResponse
		if err := json.Unmarshal(created.Body.Bytes(), &customer); err != nil || customer.Name != "Ada Lovelace" {
			t.Fatalf("unexpected customer: %#v, %v", customer, err)
		}

		listed := performJSONRequest(router, http.MethodGet, "/customers", nil)
		if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), "ada@example.com") || strings.Contains(listed.Body.String(), "password") {
			t.Fatalf("unexpected list response: %d %s", listed.Code, listed.Body.String())
		}

		found := performJSONRequest(router, http.MethodGet, "/customers/"+customer.ID, nil)
		if found.Code != http.StatusOK || !strings.Contains(found.Body.String(), "Ada Lovelace") {
			t.Fatalf("unexpected get response: %d %s", found.Code, found.Body.String())
		}
	})

	t.Run("it should validate challenge customer credentials", func(t *testing.T) {
		for _, request := range []dto.CreateCustomerRequest{
			{Name: "Missing", Email: "missing@example.com"},
			{Name: "Both", Email: "both@example.com", Password: "one", PasswordHash: "two"},
			{Email: "nameless@example.com", Password: "password"},
		} {
			response := performJSONRequest(router, http.MethodPost, "/customers", request)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected bad request, received %d: %s", response.Code, response.Body.String())
			}
		}
	})

	t.Run("it should reject a duplicate customer", func(t *testing.T) {
		responseRecorder := performJSONRequest(router, http.MethodPost, "/register", dto.CredentialsRequest{Email: "customer@example.com", Password: "password"})
		if responseRecorder.Code != http.StatusConflict {
			t.Fatalf("expected conflict, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should authenticate a customer", func(t *testing.T) {
		responseRecorder := performJSONRequest(router, http.MethodPost, "/login", dto.CredentialsRequest{Email: "customer@example.com", Password: "password"})
		if responseRecorder.Code != http.StatusOK {
			t.Fatalf("expected success, received %d and %s", responseRecorder.Code, responseRecorder.Body.String())
		}
		var response dto.AccessTokenResponse
		if errorValue := json.Unmarshal(responseRecorder.Body.Bytes(), &response); errorValue != nil || response.AccessToken == "" {
			t.Fatalf("unexpected response: %#v, %v", response, errorValue)
		}
	})

	t.Run("it should reject malformed JSON", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("{"))
		request.Header.Set("Content-Type", "application/json")
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should reject invalid registration data", func(t *testing.T) {
		responseRecorder := performJSONRequest(router, http.MethodPost, "/register", dto.CredentialsRequest{Email: " ", Password: "password"})
		if responseRecorder.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should reject malformed login JSON", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("{"))
		request.Header.Set("Content-Type", "application/json")
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)
		if responseRecorder.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request, received %d", responseRecorder.Code)
		}
	})

	t.Run("it should reject invalid login credentials", func(t *testing.T) {
		responseRecorder := performJSONRequest(router, http.MethodPost, "/login", dto.CredentialsRequest{Email: "customer@example.com", Password: "wrong"})
		if responseRecorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected unauthorized, received %d", responseRecorder.Code)
		}
	})
}

func performJSONRequest(router http.Handler, method string, path string, value any) *httptest.ResponseRecorder {
	requestBody, _ := json.Marshal(value)
	request := httptest.NewRequest(method, path, bytes.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	return responseRecorder
}
