package catalogrepository

import (
	"context"
	"sort"
	"sync"

	"github.com/afraniocaires/ecommerce/internal/catalog/domain"
	"github.com/afraniocaires/ecommerce/internal/catalog/usecase"
)

type ProductRepository struct {
	mutex    sync.RWMutex
	products map[string]domain.Product
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{products: make(map[string]domain.Product)}
}

func (repository *ProductRepository) Save(applicationContext context.Context, product *domain.Product) error {
	_ = applicationContext
	repository.mutex.Lock()
	defer repository.mutex.Unlock()
	repository.products[product.ID] = *product
	return nil
}

func (repository *ProductRepository) FindByID(applicationContext context.Context, productID string) (*domain.Product, error) {
	_ = applicationContext
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()
	product, available := repository.products[productID]
	if !available {
		return nil, domain.ErrProductNotFound
	}
	return &product, nil
}

func (repository *ProductRepository) FindByIDs(applicationContext context.Context, productIDs []string) ([]*domain.Product, error) {
	_ = applicationContext
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()
	products := make([]*domain.Product, 0, len(productIDs))
	for _, productID := range productIDs {
		if product, available := repository.products[productID]; available {
			productCopy := product
			products = append(products, &productCopy)
		}
	}
	return products, nil
}

func (repository *ProductRepository) FindPage(applicationContext context.Context, pageRequest usecase.ProductPageRequest) ([]*domain.Product, int64, error) {
	_ = applicationContext
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()
	products := make([]domain.Product, 0, len(repository.products))
	for _, product := range repository.products {
		if product.Status == domain.ProductStatusActive {
			products = append(products, product)
		}
	}
	sort.Slice(products, func(firstIndex int, secondIndex int) bool {
		if products[firstIndex].CreatedAt.Equal(products[secondIndex].CreatedAt) {
			return products[firstIndex].ID < products[secondIndex].ID
		}
		return products[firstIndex].CreatedAt.After(products[secondIndex].CreatedAt)
	})
	totalItems := int64(len(products))
	startIndex := (pageRequest.PageNumber - 1) * pageRequest.PageSize
	if startIndex >= len(products) {
		return []*domain.Product{}, totalItems, nil
	}
	endIndex := min(startIndex+pageRequest.PageSize, len(products))
	pageProducts := make([]*domain.Product, 0, endIndex-startIndex)
	for index := startIndex; index < endIndex; index++ {
		productCopy := products[index]
		pageProducts = append(pageProducts, &productCopy)
	}
	return pageProducts, totalItems, nil
}
