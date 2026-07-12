package authenticationtransport

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/afraniocaires/ecommerce/internal/authentication/adapter/http/dto"
	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
	"github.com/afraniocaires/ecommerce/internal/authentication/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/httpresponse"
)

type Handler struct {
	registerUserUseCase *usecase.RegisterUserUseCase
	loginUserUseCase    *usecase.LoginUserUseCase
}

func NewHandler(
	registerUserUseCase *usecase.RegisterUserUseCase,
	loginUserUseCase *usecase.LoginUserUseCase,
) *Handler {
	return &Handler{
		registerUserUseCase: registerUserUseCase,
		loginUserUseCase:    loginUserUseCase,
	}
}

// Register godoc
// @Summary Register a customer
// @Description Creates a customer account.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.CredentialsRequest true "Customer credentials"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} httpresponse.ErrorResponse
// @Failure 409 {object} httpresponse.ErrorResponse
// @Router /api/authentication/register [post]
func (handler *Handler) Register(context *gin.Context) {
	var request dto.CredentialsRequest

	if errorValue := context.ShouldBindJSON(&request); errorValue != nil {
		context.JSON(http.StatusBadRequest, httpresponse.ErrorResponse{Error: "the JSON request body is invalid."})
		return
	}

	user, errorValue := handler.registerUserUseCase.Execute(
		context.Request.Context(),
		usecase.RegisterUserInput{
			Email:    request.Email,
			Password: request.Password,
		},
	)
	if errorValue != nil {
		statusCode := http.StatusBadRequest
		if errors.Is(errorValue, domain.ErrEmailAlreadyUsed) {
			statusCode = http.StatusConflict
		}

		context.JSON(statusCode, httpresponse.ErrorResponse{Error: errorValue.Error()})
		return
	}

	roleValues := make([]string, len(user.Roles))
	for index, role := range user.Roles {
		roleValues[index] = string(role)
	}

	context.JSON(
		http.StatusCreated,
		dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Roles:     roleValues,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
	)
}

// Login godoc
// @Summary Authenticate a customer
// @Description Returns a JWT access token for valid credentials.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.CredentialsRequest true "Customer credentials"
// @Success 200 {object} dto.AccessTokenResponse
// @Failure 400 {object} httpresponse.ErrorResponse
// @Failure 401 {object} httpresponse.ErrorResponse
// @Router /api/authentication/login [post]
func (handler *Handler) Login(context *gin.Context) {
	var request dto.CredentialsRequest

	if errorValue := context.ShouldBindJSON(&request); errorValue != nil {
		context.JSON(http.StatusBadRequest, httpresponse.ErrorResponse{Error: "the JSON request body is invalid."})
		return
	}

	output, errorValue := handler.loginUserUseCase.Execute(
		context.Request.Context(),
		usecase.LoginUserInput{
			Email:    request.Email,
			Password: request.Password,
		},
	)
	if errorValue != nil {
		context.JSON(http.StatusUnauthorized, httpresponse.ErrorResponse{Error: domain.ErrInvalidCredentials.Error()})
		return
	}

	context.JSON(http.StatusOK, dto.AccessTokenResponse{AccessToken: output.AccessToken})
}
