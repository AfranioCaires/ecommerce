package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/afraniocaires/ecommerce/internal/inventory/domain"
)

var errInventoryDependency = errors.New("inventory dependency failed")

type inventoryRepositoryFake struct {
	stocks    map[string]*domain.Stock
	findError error
	saveError error
	saved     []*domain.Stock
}

func (repository *inventoryRepositoryFake) Save(_ context.Context, stock *domain.Stock) error {
	if repository.saveError != nil {
		return repository.saveError
	}
	repository.saved = append(repository.saved, stock)
	if repository.stocks != nil {
		repository.stocks[stock.ProductID] = stock
	}
	return nil
}

func (repository *inventoryRepositoryFake) FindByProductID(_ context.Context, productID string) (*domain.Stock, error) {
	if repository.findError != nil {
		return nil, repository.findError
	}
	stock, found := repository.stocks[productID]
	if !found {
		return nil, domain.ErrStockNotFound
	}
	return stock, nil
}

func (repository *inventoryRepositoryFake) FindByProductIDForUpdate(applicationContext context.Context, productID string) (*domain.Stock, error) {
	return repository.FindByProductID(applicationContext, productID)
}

func TestAggregateStockItems(t *testing.T) {
	result := AggregateStockItems([]StockItem{
		{ProductID: "product-2", Quantity: 1},
		{ProductID: "product-1", Quantity: 2},
		{ProductID: "product-2", Quantity: 3},
	})
	want := []StockItem{{ProductID: "product-1", Quantity: 2}, {ProductID: "product-2", Quantity: 4}}
	if !reflect.DeepEqual(result, want) {
		t.Fatalf("AggregateStockItems() = %#v, want %#v", result, want)
	}
	if empty := AggregateStockItems(nil); len(empty) != 0 {
		t.Fatalf("AggregateStockItems(nil) = %#v", empty)
	}
}

func TestInventoryServiceSetQuantity(t *testing.T) {
	now := time.Date(2026, time.July, 15, 15, 0, 0, 0, time.FixedZone("test", -3*60*60))

	t.Run("rejects invalid stock", func(t *testing.T) {
		repository := &inventoryRepositoryFake{}
		service := NewInventoryService(repository, func() time.Time { return now })
		stock, errorValue := service.SetQuantity(context.Background(), "product-1", -1)
		if stock != nil || !errors.Is(errorValue, domain.ErrInvalidStockQuantity) || len(repository.saved) != 0 {
			t.Fatalf("SetQuantity() = %#v, %v; saves = %d", stock, errorValue, len(repository.saved))
		}
	})

	t.Run("propagates save error", func(t *testing.T) {
		repository := &inventoryRepositoryFake{saveError: errInventoryDependency}
		service := NewInventoryService(repository, func() time.Time { return now })
		stock, errorValue := service.SetQuantity(context.Background(), "product-1", 3)
		if stock != nil || !errors.Is(errorValue, errInventoryDependency) {
			t.Fatalf("SetQuantity() = %#v, %v", stock, errorValue)
		}
	})

	t.Run("saves and returns stock", func(t *testing.T) {
		repository := &inventoryRepositoryFake{}
		service := NewInventoryService(repository, func() time.Time { return now })
		stock, errorValue := service.SetQuantity(context.Background(), "product-1", 3)
		if errorValue != nil || len(repository.saved) != 1 || repository.saved[0] != stock || stock.UpdatedAt.Location() != time.UTC || !stock.UpdatedAt.Equal(now) {
			t.Fatalf("SetQuantity() = %#v, %v; saves = %#v", stock, errorValue, repository.saved)
		}
	})
}

func TestInventoryServiceReserve(t *testing.T) {
	testInventoryMutation(t, "reserve", func(service *InventoryService, items []StockItem) error {
		return service.Reserve(context.Background(), items)
	}, 3, domain.ErrInsufficientStock)
}

func TestInventoryServiceRelease(t *testing.T) {
	testInventoryMutation(t, "release", func(service *InventoryService, items []StockItem) error {
		return service.Release(context.Background(), items)
	}, 7, domain.ErrInvalidStockQuantity)
}

func testInventoryMutation(t *testing.T, operation string, mutate func(*InventoryService, []StockItem) error, successfulQuantity int, mutationError error) {
	t.Helper()
	now := time.Date(2026, time.July, 15, 15, 0, 0, 0, time.FixedZone("test", -3*60*60))

	t.Run(operation+" propagates lookup error", func(t *testing.T) {
		repository := &inventoryRepositoryFake{findError: errInventoryDependency}
		service := NewInventoryService(repository, func() time.Time { return now })
		if errorValue := mutate(service, []StockItem{{ProductID: "product-1", Quantity: 2}}); !errors.Is(errorValue, errInventoryDependency) {
			t.Fatalf("mutation error = %v", errorValue)
		}
	})

	t.Run(operation+" propagates domain error", func(t *testing.T) {
		quantity := 2
		if operation == "release" {
			quantity = -1
		}
		repository := &inventoryRepositoryFake{stocks: map[string]*domain.Stock{"product-1": {ProductID: "product-1", Quantity: 1}}}
		service := NewInventoryService(repository, func() time.Time { return now })
		if errorValue := mutate(service, []StockItem{{ProductID: "product-1", Quantity: quantity}}); !errors.Is(errorValue, mutationError) {
			t.Fatalf("mutation error = %v", errorValue)
		}
	})

	t.Run(operation+" propagates save error", func(t *testing.T) {
		repository := &inventoryRepositoryFake{
			stocks:    map[string]*domain.Stock{"product-1": {ProductID: "product-1", Quantity: 5}},
			saveError: errInventoryDependency,
		}
		service := NewInventoryService(repository, func() time.Time { return now })
		if errorValue := mutate(service, []StockItem{{ProductID: "product-1", Quantity: 2}}); !errors.Is(errorValue, errInventoryDependency) {
			t.Fatalf("mutation error = %v", errorValue)
		}
	})

	t.Run(operation+" saves aggregated items", func(t *testing.T) {
		repository := &inventoryRepositoryFake{stocks: map[string]*domain.Stock{
			"product-1": {ProductID: "product-1", Quantity: 5},
			"product-2": {ProductID: "product-2", Quantity: 5},
		}}
		service := NewInventoryService(repository, func() time.Time { return now })
		errorValue := mutate(service, []StockItem{{ProductID: "product-2", Quantity: 1}, {ProductID: "product-1", Quantity: 1}, {ProductID: "product-2", Quantity: 1}})
		if errorValue != nil || len(repository.saved) != 2 || repository.stocks["product-2"].Quantity != successfulQuantity || repository.stocks["product-1"].UpdatedAt.Location() != time.UTC {
			t.Fatalf("mutation error = %v; stocks = %#v; saves = %#v", errorValue, repository.stocks, repository.saved)
		}
	})
}
