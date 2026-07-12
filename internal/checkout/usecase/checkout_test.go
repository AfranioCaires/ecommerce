package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	catalogdomain "github.com/afraniocaires/ecommerce/internal/catalog/domain"
	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	orderdomain "github.com/afraniocaires/ecommerce/internal/order/domain"
	paymentdomain "github.com/afraniocaires/ecommerce/internal/payment/domain"
)

type checkoutDependencies struct {
	products      []*catalogdomain.Product
	paymentStatus paymentdomain.PaymentStatus
	released      bool
	updatedOrder  *orderdomain.Order
}

func (dependencies *checkoutDependencies) FindByIDs(applicationContext context.Context, productIDs []string) ([]*catalogdomain.Product, error) {
	return dependencies.products, nil
}

func (dependencies *checkoutDependencies) Reserve(applicationContext context.Context, stockItems []inventoryusecase.StockItem) error {
	return nil
}

func (dependencies *checkoutDependencies) Release(applicationContext context.Context, stockItems []inventoryusecase.StockItem) error {
	dependencies.released = true
	return nil
}

func (dependencies *checkoutDependencies) Save(applicationContext context.Context, order *orderdomain.Order) error {
	return nil
}

func (dependencies *checkoutDependencies) UpdateStatus(applicationContext context.Context, order *orderdomain.Order) error {
	dependencies.updatedOrder = order
	return nil
}

func (dependencies *checkoutDependencies) Process(applicationContext context.Context, orderID string, amountCents int64) (*paymentdomain.Payment, error) {
	return paymentdomain.NewPayment("payment-1", orderID, amountCents, dependencies.paymentStatus, time.Now())
}

func (dependencies *checkoutDependencies) Execute(applicationContext context.Context, operation func(transactionContext context.Context) error) error {
	return operation(applicationContext)
}

func TestCheckoutUseCase(t *testing.T) {
	product, _ := catalogdomain.NewProduct("product-1", "Keyboard", "", 1013, time.Now())

	t.Run("it should mark an approved checkout as paid", func(t *testing.T) {
		dependencies := &checkoutDependencies{products: []*catalogdomain.Product{product}, paymentStatus: paymentdomain.PaymentStatusApproved}
		useCase := NewCheckoutUseCase(dependencies, dependencies, dependencies, dependencies, dependencies, time.Now)
		output, errorValue := useCase.Execute(context.Background(), CheckoutInput{UserID: "user-1", Items: []CheckoutItemInput{{ProductID: product.ID, Quantity: 1}}})
		if errorValue != nil || output.Order.Status != orderdomain.OrderStatusPaid || dependencies.released {
			t.Fatalf("unexpected approved checkout: %#v, %v", output, errorValue)
		}
	})

	t.Run("it should restore stock and fail a declined checkout", func(t *testing.T) {
		dependencies := &checkoutDependencies{products: []*catalogdomain.Product{product}, paymentStatus: paymentdomain.PaymentStatusDeclined}
		useCase := NewCheckoutUseCase(dependencies, dependencies, dependencies, dependencies, dependencies, time.Now)
		output, errorValue := useCase.Execute(context.Background(), CheckoutInput{UserID: "user-1", Items: []CheckoutItemInput{{ProductID: product.ID, Quantity: 1}}})
		if errorValue != nil || output.Order.Status != orderdomain.OrderStatusFailed || !dependencies.released {
			t.Fatalf("unexpected declined checkout: %#v, %v", output, errorValue)
		}
	})

	t.Run("it should reject an empty checkout", func(t *testing.T) {
		dependencies := &checkoutDependencies{}
		useCase := NewCheckoutUseCase(dependencies, dependencies, dependencies, dependencies, dependencies, time.Now)
		_, errorValue := useCase.Execute(context.Background(), CheckoutInput{UserID: "user-1"})
		if !errors.Is(errorValue, ErrEmptyCheckoutItems) {
			t.Fatalf("expected empty checkout error, received %v", errorValue)
		}
	})
}
