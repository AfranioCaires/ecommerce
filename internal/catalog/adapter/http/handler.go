package catalogtransport

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

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

func (handler *Handler) Create(context *gin.Context) {
	var request dto.CreateProductRequest

	if errorValue := context.ShouldBindJSON(&request); errorValue != nil {
		context.JSON(http.StatusBadRequest, httpresponse.ErrorResponse{Error: "the JSON request body is invalid."})
		return
	}

	product, errorValue := handler.createProductUseCase.Execute(
		context.Request.Context(),
		usecase.CreateProductInput{
			Name:        request.Name,
			Description: request.Description,
			PriceCents:  request.PriceCents,
		},
	)
	if errorValue != nil {
		context.JSON(http.StatusBadRequest, httpresponse.ErrorResponse{Error: errorValue.Error()})
		return
	}

	context.JSON(http.StatusCreated, toProductResponse(product))
}

func (handler *Handler) GetByID(context *gin.Context) {
	productID := context.Param("productID")

	product, errorValue := handler.getProductUseCase.Execute(
		context.Request.Context(),
		productID,
	)
	if errorValue != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(errorValue, domain.ErrProductNotFound) {
			statusCode = http.StatusNotFound
		}

		context.JSON(statusCode, httpresponse.ErrorResponse{Error: errorValue.Error()})
		return
	}

	context.JSON(http.StatusOK, toProductResponse(product))
}

func (handler *Handler) List(context *gin.Context) {
	pageNumber, errorValue := integerQueryValue(context, "page", usecase.DefaultPageNumber)
	if errorValue != nil {
		context.JSON(http.StatusBadRequest, httpresponse.ErrorResponse{Error: usecase.ErrInvalidPagination.Error()})
		return
	}

	pageSize, errorValue := integerQueryValue(context, "page_size", usecase.DefaultPageSize)
	if errorValue != nil {
		context.JSON(http.StatusBadRequest, httpresponse.ErrorResponse{Error: usecase.ErrInvalidPagination.Error()})
		return
	}

	pageRequest, errorValue := usecase.NewProductPageRequest(pageNumber, pageSize)
	if errorValue != nil {
		context.JSON(http.StatusBadRequest, httpresponse.ErrorResponse{Error: errorValue.Error()})
		return
	}

	productPage, errorValue := handler.listProductsUseCase.Execute(
		context.Request.Context(),
		pageRequest,
	)
	if errorValue != nil {
		context.JSON(http.StatusInternalServerError, httpresponse.ErrorResponse{Error: errorValue.Error()})
		return
	}

	productResponses := make([]dto.ProductResponse, 0, len(productPage.Products))
	for _, product := range productPage.Products {
		productResponses = append(productResponses, toProductResponse(product))
	}

	context.JSON(http.StatusOK, dto.ProductPageResponse{
		Products:   productResponses,
		PageNumber: productPage.PageNumber,
		PageSize:   productPage.PageSize,
		TotalItems: productPage.TotalItems,
		TotalPages: productPage.TotalPages,
	})
}

func integerQueryValue(context *gin.Context, name string, fallback int) (int, error) {
	value := context.Query(name)
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
