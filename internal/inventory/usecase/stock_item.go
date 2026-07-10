package usecase

import "sort"

type StockItem struct {
	ProductID string
	Quantity  int
}

func AggregateStockItems(stockItems []StockItem) []StockItem {
	quantitiesByProductID := make(map[string]int, len(stockItems))

	for _, stockItem := range stockItems {
		quantitiesByProductID[stockItem.ProductID] += stockItem.Quantity
	}

	aggregatedStockItems := make([]StockItem, 0, len(quantitiesByProductID))

	for productID, quantity := range quantitiesByProductID {
		aggregatedStockItems = append(aggregatedStockItems, StockItem{
			ProductID: productID,
			Quantity:  quantity,
		})
	}

	sort.Slice(aggregatedStockItems, func(firstIndex int, secondIndex int) bool {
		return aggregatedStockItems[firstIndex].ProductID < aggregatedStockItems[secondIndex].ProductID
	})

	return aggregatedStockItems
}
