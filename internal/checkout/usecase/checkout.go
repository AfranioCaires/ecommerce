package usecase

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"

	catalogdomain "github.com/afraniocaires/ecommerce/internal/catalog/domain"
	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	orderdomain "github.com/afraniocaires/ecommerce/internal/order/domain"
	paymentdomain "github.com/afraniocaires/ecommerce/internal/payment/domain"
)

var (
	ErrEmptyCheckoutItems      = errors.New("the checkout must contain at least one item.")
	ErrInvalidCheckoutItem     = errors.New("a checkout item is invalid.")
	ErrCheckoutProductNotFound = errors.New("a requested product was not found.")
	ErrInactiveCheckoutProduct = errors.New("a requested product is inactive.")
)

type CheckoutUseCase struct {
	productReader      ProductReader
	inventoryManager   InventoryManager
	orderWriter        OrderWriter
	paymentProcessor   PaymentProcessor
	transactionManager TransactionManager
	currentTime        func() time.Time
}

func NewCheckoutUseCase(
	productReader ProductReader,
	inventoryManager InventoryManager,
	orderWriter OrderWriter,
	paymentProcessor PaymentProcessor,
	transactionManager TransactionManager,
	currentTime func() time.Time,
) *CheckoutUseCase {
	return &CheckoutUseCase{
		productReader:      productReader,
		inventoryManager:   inventoryManager,
		orderWriter:        orderWriter,
		paymentProcessor:   paymentProcessor,
		transactionManager: transactionManager,
		currentTime:        currentTime,
	}
}

func (useCase *CheckoutUseCase) Execute(
	applicationContext context.Context,
	input CheckoutInput,
) (*CheckoutOutput, error) {
	if len(input.Items) == 0 {
		return nil, ErrEmptyCheckoutItems
	}

	aggregatedItems, productIDs, errorValue := aggregateCheckoutItems(input.Items)
	if errorValue != nil {
		return nil, errorValue
	}

	var checkoutOutput CheckoutOutput

	errorValue = useCase.transactionManager.Execute(
		applicationContext,
		func(transactionContext context.Context) error {
			products, errorValue := useCase.productReader.FindByIDs(
				transactionContext,
				productIDs,
			)
			if errorValue != nil {
				return errorValue
			}

			productsByID := make(map[string]*catalogdomain.Product, len(products))

			for _, product := range products {
				productsByID[product.ID] = product
			}

			if len(productsByID) != len(productIDs) {
				return ErrCheckoutProductNotFound
			}

			orderItems := make([]orderdomain.OrderItem, 0, len(aggregatedItems))
			stockItems := make([]inventoryusecase.StockItem, 0, len(aggregatedItems))

			for _, checkoutItem := range aggregatedItems {
				product, available := productsByID[checkoutItem.ProductID]
				if !available {
					return ErrCheckoutProductNotFound
				}

				if product.Status != catalogdomain.ProductStatusActive {
					return ErrInactiveCheckoutProduct
				}

				orderItems = append(orderItems, orderdomain.OrderItem{
					ProductID:      product.ID,
					ProductName:    product.Name,
					UnitPriceCents: product.PriceCents,
					Quantity:       checkoutItem.Quantity,
				})

				stockItems = append(stockItems, inventoryusecase.StockItem{
					ProductID: product.ID,
					Quantity:  checkoutItem.Quantity,
				})
			}

			if errorValue := useCase.inventoryManager.Reserve(
				transactionContext,
				stockItems,
			); errorValue != nil {
				return errorValue
			}

			order, errorValue := orderdomain.NewOrder(
				uuid.NewString(),
				input.UserID,
				orderItems,
				useCase.currentTime(),
			)
			if errorValue != nil {
				return errorValue
			}

			if errorValue := useCase.orderWriter.Save(
				transactionContext,
				order,
			); errorValue != nil {
				return errorValue
			}

			payment, errorValue := useCase.paymentProcessor.Process(
				transactionContext,
				order.ID,
				order.TotalAmountCents,
			)
			if errorValue != nil {
				return errorValue
			}

			if payment.Status == paymentdomain.PaymentStatusDeclined {
				if errorValue := useCase.inventoryManager.Release(
					transactionContext,
					stockItems,
				); errorValue != nil {
					return errorValue
				}

				order.MarkAsFailed(useCase.currentTime())
			} else {
				order.MarkAsPaid(useCase.currentTime())
			}

			if errorValue := useCase.orderWriter.UpdateStatus(
				transactionContext,
				order,
			); errorValue != nil {
				return errorValue
			}

			checkoutOutput = CheckoutOutput{
				Order:   order,
				Payment: payment,
			}

			return nil
		},
	)
	if errorValue != nil {
		return nil, errorValue
	}

	return &checkoutOutput, nil
}

func aggregateCheckoutItems(
	checkoutItems []CheckoutItemInput,
) ([]CheckoutItemInput, []string, error) {
	quantitiesByProductID := make(map[string]int, len(checkoutItems))

	for _, checkoutItem := range checkoutItems {
		if checkoutItem.ProductID == "" || checkoutItem.Quantity <= 0 {
			return nil, nil, ErrInvalidCheckoutItem
		}

		quantitiesByProductID[checkoutItem.ProductID] += checkoutItem.Quantity
	}

	aggregatedItems := make([]CheckoutItemInput, 0, len(quantitiesByProductID))
	productIDs := make([]string, 0, len(quantitiesByProductID))

	for productID, quantity := range quantitiesByProductID {
		aggregatedItems = append(aggregatedItems, CheckoutItemInput{
			ProductID: productID,
			Quantity:  quantity,
		})
	}

	sort.Slice(aggregatedItems, func(firstIndex int, secondIndex int) bool {
		return aggregatedItems[firstIndex].ProductID < aggregatedItems[secondIndex].ProductID
	})

	for _, aggregatedItem := range aggregatedItems {
		productIDs = append(productIDs, aggregatedItem.ProductID)
	}

	return aggregatedItems, productIDs, nil
}
