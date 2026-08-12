package authenticationtransport

import (
	"errors"
	"net/http"
	"time"

	"github.com/afraniocaires/ecommerce/internal/authentication/adapter/http/dto"
	"github.com/afraniocaires/ecommerce/internal/authentication/domain"
	"github.com/afraniocaires/ecommerce/internal/authentication/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/httpresponse"
)

type Handler struct {
	registerUserUseCase *usecase.RegisterUserUseCase
	loginUserUseCase    *usecase.LoginUserUseCase
	getUserUseCase      *usecase.GetUserUseCase
	listUsersUseCase    *usecase.ListUsersUseCase
}

func NewHandler(
	registerUserUseCase *usecase.RegisterUserUseCase,
	loginUserUseCase *usecase.LoginUserUseCase,
	getUserUseCase *usecase.GetUserUseCase,
	listUsersUseCase *usecase.ListUsersUseCase,
) *Handler {
	return &Handler{
		registerUserUseCase: registerUserUseCase,
		loginUserUseCase:    loginUserUseCase,
		getUserUseCase:      getUserUseCase,
		listUsersUseCase:    listUsersUseCase,
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
func (handler *Handler) Register(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	var credentialsRequest dto.CredentialsRequest
	if errorValue := httpresponse.DecodeJSON(
		responseWriter,
		request,
		&credentialsRequest,
	); errorValue != nil {
		httpresponse.JSON(
			responseWriter,
			http.StatusBadRequest,
			httpresponse.ErrorResponse{Error: "the JSON request body is invalid."},
		)
		return
	}

	user, errorValue := handler.registerUserUseCase.Execute(
		request.Context(),
		usecase.RegisterUserInput{
			Email:    credentialsRequest.Email,
			Password: credentialsRequest.Password,
		},
	)
	if errorValue != nil {
		statusCode := http.StatusInternalServerError
		message := "an unexpected error occurred."
		switch {
		case errors.Is(errorValue, domain.ErrEmailAlreadyUsed):
			statusCode = http.StatusConflict
			message = errorValue.Error()
		case errors.Is(errorValue, domain.ErrEmptyEmail),
			errors.Is(errorValue, domain.ErrEmptyPasswordHash),
			errors.Is(errorValue, domain.ErrInvalidRole):
			statusCode = http.StatusBadRequest
			message = errorValue.Error()
		}

		httpresponse.JSON(
			responseWriter,
			statusCode,
			httpresponse.ErrorResponse{Error: message},
		)
		return
	}

	httpresponse.JSON(responseWriter, http.StatusCreated, toUserResponse(user))
}

func (handler *Handler) CreateCustomer(responseWriter http.ResponseWriter, request *http.Request) {
	var createRequest dto.CreateCustomerRequest
	if errorValue := httpresponse.DecodeJSON(responseWriter, request, &createRequest); errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusBadRequest, httpresponse.ErrorResponse{Error: "the JSON request body is invalid."})
		return
	}

	if createRequest.Name == "" || (createRequest.Password == "") == (createRequest.PasswordHash == "") {
		httpresponse.JSON(responseWriter, http.StatusBadRequest, httpresponse.ErrorResponse{Error: "name and exactly one password field are required."})
		return
	}
	password := createRequest.Password
	if password == "" {
		password = createRequest.PasswordHash
	}
	user, errorValue := handler.registerUserUseCase.Execute(request.Context(), usecase.RegisterUserInput{
		Name: createRequest.Name, Email: createRequest.Email, Password: password,
	})
	if errorValue != nil {
		writeRegistrationError(responseWriter, errorValue)
		return
	}
	httpresponse.JSON(responseWriter, http.StatusCreated, toUserResponse(user))
}

func (handler *Handler) ListCustomers(responseWriter http.ResponseWriter, request *http.Request) {
	users, errorValue := handler.listUsersUseCase.Execute(request.Context())
	if errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusInternalServerError, httpresponse.ErrorResponse{Error: "an unexpected error occurred."})
		return
	}
	responses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, toUserResponse(user))
	}
	httpresponse.JSON(responseWriter, http.StatusOK, responses)
}

func (handler *Handler) GetCustomer(responseWriter http.ResponseWriter, request *http.Request) {
	user, errorValue := handler.getUserUseCase.Execute(request.Context(), request.PathValue("customerID"))
	if errorValue != nil {
		if errors.Is(errorValue, domain.ErrUserNotFound) {
			httpresponse.JSON(responseWriter, http.StatusNotFound, httpresponse.ErrorResponse{Error: errorValue.Error()})
			return
		}
		httpresponse.JSON(responseWriter, http.StatusInternalServerError, httpresponse.ErrorResponse{Error: "an unexpected error occurred."})
		return
	}
	httpresponse.JSON(responseWriter, http.StatusOK, toUserResponse(user))
}

func writeRegistrationError(responseWriter http.ResponseWriter, errorValue error) {
	statusCode := http.StatusInternalServerError
	message := "an unexpected error occurred."
	switch {
	case errors.Is(errorValue, domain.ErrEmailAlreadyUsed):
		statusCode, message = http.StatusConflict, errorValue.Error()
	case errors.Is(errorValue, domain.ErrEmptyUserName), errors.Is(errorValue, domain.ErrEmptyEmail),
		errors.Is(errorValue, domain.ErrEmptyPasswordHash), errors.Is(errorValue, domain.ErrInvalidRole):
		statusCode, message = http.StatusBadRequest, errorValue.Error()
	}
	httpresponse.JSON(responseWriter, statusCode, httpresponse.ErrorResponse{Error: message})
}

func toUserResponse(user *domain.User) dto.UserResponse {
	roleValues := make([]string, len(user.Roles))
	for index, role := range user.Roles {
		roleValues[index] = string(role)
	}
	return dto.UserResponse{ID: user.ID, Name: user.Name, Email: user.Email, Roles: roleValues, CreatedAt: user.CreatedAt.Format(time.RFC3339)}
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
func (handler *Handler) Login(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	var credentialsRequest dto.CredentialsRequest
	if errorValue := httpresponse.DecodeJSON(
		responseWriter,
		request,
		&credentialsRequest,
	); errorValue != nil {
		httpresponse.JSON(
			responseWriter,
			http.StatusBadRequest,
			httpresponse.ErrorResponse{Error: "the JSON request body is invalid."},
		)
		return
	}

	output, errorValue := handler.loginUserUseCase.Execute(
		request.Context(),
		usecase.LoginUserInput{
			Email:    credentialsRequest.Email,
			Password: credentialsRequest.Password,
		},
	)
	if errorValue != nil {
		httpresponse.JSON(
			responseWriter,
			http.StatusUnauthorized,
			httpresponse.ErrorResponse{Error: domain.ErrInvalidCredentials.Error()},
		)
		return
	}

	httpresponse.JSON(
		responseWriter,
		http.StatusOK,
		dto.AccessTokenResponse{AccessToken: output.AccessToken},
	)
}
