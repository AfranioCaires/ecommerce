package usecase

import (
	"context"
	"time"

	"github.com/afraniocaires/ecommerce/internal/inventory/domain"
)

type InventoryService struct {
	stockRepository StockRepository
	currentTime     func() time.Time
}

func NewInventoryService(
	stockRepository StockRepository,
	currentTime func() time.Time,
) *InventoryService {
	return &InventoryService{
		stockRepository: stockRepository,
		currentTime:     currentTime,
	}
}

func (service *InventoryService) SetQuantity(
	context context.Context,
	productID string,
	quantity int,
) (*domain.Stock, error) {
	stock, errorValue := domain.NewStock(
		productID,
		quantity,
		service.currentTime(),
	)
	if errorValue != nil {
		return nil, errorValue
	}

	if errorValue := service.stockRepository.Save(
		context,
		stock,
	); errorValue != nil {
		return nil, errorValue
	}

	return stock, nil
}

func (service *InventoryService) Reserve(
	context context.Context,
	stockItems []StockItem,
) error {
	aggregatedStockItems := AggregateStockItems(stockItems)

	for _, stockItem := range aggregatedStockItems {
		stock, errorValue := service.stockRepository.FindByProductIDForUpdate(
			context,
			stockItem.ProductID,
		)
		if errorValue != nil {
			return errorValue
		}

		if errorValue := stock.Reserve(
			stockItem.Quantity,
			service.currentTime(),
		); errorValue != nil {
			return errorValue
		}

		if errorValue := service.stockRepository.Save(
			context,
			stock,
		); errorValue != nil {
			return errorValue
		}
	}

	return nil
}

func (service *InventoryService) Release(
	context context.Context,
	stockItems []StockItem,
) error {
	aggregatedStockItems := AggregateStockItems(stockItems)

	for _, stockItem := range aggregatedStockItems {
		stock, errorValue := service.stockRepository.FindByProductIDForUpdate(
			context,
			stockItem.ProductID,
		)
		if errorValue != nil {
			return errorValue
		}

		if errorValue := stock.Release(
			stockItem.Quantity,
			service.currentTime(),
		); errorValue != nil {
			return errorValue
		}

		if errorValue := service.stockRepository.Save(
			context,
			stock,
		); errorValue != nil {
			return errorValue
		}
	}

	return nil
}
