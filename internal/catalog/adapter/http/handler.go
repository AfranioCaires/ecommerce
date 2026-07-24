package catalogtransport

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/afraniocaires/ecommerce/internal/catalog/adapter/http/dto"
	"github.com/afraniocaires/ecommerce/internal/catalog/domain"
	"github.com/afraniocaires/ecommerce/internal/catalog/usecase"
	"github.com/afraniocaires/ecommerce/internal/platform/httpresponse"
)

type Handler struct {
	createProductUseCase *usecase.CreateProductUseCase
	getProductUseCase    *usecase.GetProductUseCase
	listProductsUseCase  *usecase.ListProductsUseCase
}

func NewHandler(
	createProductUseCase *usecase.CreateProductUseCase,
	getProductUseCase *usecase.GetProductUseCase,
	listProductsUseCase *usecase.ListProductsUseCase,
) *Handler {
	return &Handler{
		createProductUseCase: createProductUseCase,
		getProductUseCase:    getProductUseCase,
		listProductsUseCase:  listProductsUseCase,
	}
}

func (handler *Handler) Create(responseWriter http.ResponseWriter, request *http.Request) {
	var createRequest dto.CreateProductRequest
	if errorValue := httpresponse.DecodeJSON(responseWriter, request, &createRequest); errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusBadRequest, httpresponse.ErrorResponse{Error: "the JSON request body is invalid."})
		return
	}

	product, errorValue := handler.createProductUseCase.Execute(
		request.Context(),
		usecase.CreateProductInput{
			Name:        createRequest.Name,
			Description: createRequest.Description,
			PriceCents:  createRequest.PriceCents,
		},
	)
	if errorValue != nil {
		statusCode := http.StatusInternalServerError
		message := "an unexpected error occurred."
		if errors.Is(errorValue, domain.ErrEmptyProductName) ||
			errors.Is(errorValue, domain.ErrInvalidProductPrice) {
			statusCode = http.StatusBadRequest
			message = errorValue.Error()
		}
		httpresponse.JSON(responseWriter, statusCode, httpresponse.ErrorResponse{Error: message})
		return
	}

	httpresponse.JSON(responseWriter, http.StatusCreated, toProductResponse(product))
}

func (handler *Handler) GetByID(responseWriter http.ResponseWriter, request *http.Request) {
	product, errorValue := handler.getProductUseCase.Execute(
		request.Context(),
		request.PathValue("productID"),
	)
	if errorValue != nil {
		statusCode := http.StatusInternalServerError
		message := "an unexpected error occurred."
		if errors.Is(errorValue, domain.ErrProductNotFound) {
			statusCode = http.StatusNotFound
			message = errorValue.Error()
		}

		httpresponse.JSON(responseWriter, statusCode, httpresponse.ErrorResponse{Error: message})
		return
	}

	httpresponse.JSON(responseWriter, http.StatusOK, toProductResponse(product))
}

func (handler *Handler) List(responseWriter http.ResponseWriter, request *http.Request) {
	pageNumber, errorValue := integerQueryValue(request, "page", usecase.DefaultPageNumber)
	if errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusBadRequest, httpresponse.ErrorResponse{Error: usecase.ErrInvalidPagination.Error()})
		return
	}

	pageSize, errorValue := integerQueryValue(request, "page_size", usecase.DefaultPageSize)
	if errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusBadRequest, httpresponse.ErrorResponse{Error: usecase.ErrInvalidPagination.Error()})
		return
	}

	pageRequest, errorValue := usecase.NewProductPageRequest(pageNumber, pageSize)
	if errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusBadRequest, httpresponse.ErrorResponse{Error: errorValue.Error()})
		return
	}

	productPage, errorValue := handler.listProductsUseCase.Execute(request.Context(), pageRequest)
	if errorValue != nil {
		httpresponse.JSON(responseWriter, http.StatusInternalServerError, httpresponse.ErrorResponse{Error: "an unexpected error occurred."})
		return
	}

	productResponses := make([]dto.ProductResponse, 0, len(productPage.Products))
	for _, product := range productPage.Products {
		productResponses = append(productResponses, toProductResponse(product))
	}

	httpresponse.JSON(responseWriter, http.StatusOK, dto.ProductPageResponse{
		Products:   productResponses,
		PageNumber: productPage.PageNumber,
		PageSize:   productPage.PageSize,
		TotalItems: productPage.TotalItems,
		TotalPages: productPage.TotalPages,
	})
}

func integerQueryValue(request *http.Request, name string, fallback int) (int, error) {
	value := request.URL.Query().Get(name)
	if value == "" {
		return fallback, nil
	}

	return strconv.Atoi(value)
}

func toProductResponse(product *domain.Product) dto.ProductResponse {
	return dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		PriceCents:  product.PriceCents,
		Status:      string(product.Status),
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	}
}
