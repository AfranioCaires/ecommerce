package authenticationtransport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/afraniocaires/ecommerce/internal/authentication/adapter/http/dto"
	authenticationrepository "github.com/afraniocaires/ecommerce/internal/authentication/adapter/repository/memory"
	"github.com/afraniocaires/ecommerce/internal/authentication/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/security"
)

func TestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepository := authenticationrepository.NewUserRepository()
	passwordHasher := security.NewBcryptPasswordHasher(bcrypt.MinCost)
	accessTokenManager := security.NewJSONWebTokenManager("secret", "ecommerce", time.Hour)
	registerUserUseCase := usecase.NewRegisterUserUseCase(userRepository, passwordHasher, time.Now)
	loginUserUseCase := usecase.NewLoginUserUseCase(userRepository, passwordHasher, accessTokenManager, time.Now)
	handler := NewHandler(registerUserUseCase, loginUserUseCase)
	router := gin.New()
	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)

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
}

func performJSONRequest(router http.Handler, method string, path string, value any) *httptest.ResponseRecorder {
	requestBody, _ := json.Marshal(value)
	request := httptest.NewRequest(method, path, bytes.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	return responseRecorder
}
