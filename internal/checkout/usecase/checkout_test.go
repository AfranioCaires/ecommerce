package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	catalogdomain "github.com/afraniocaires/ecommerce/internal/catalog/domain"
	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	orderdomain "github.com/afraniocaires/ecommerce/internal/order/domain"
	paymentdomain "github.com/afraniocaires/ecommerce/internal/payment/domain"
)

type checkoutDependencies struct {
	products         []*catalogdomain.Product
	paymentStatus    paymentdomain.PaymentStatus
	findError        error
	reserveError     error
	releaseError     error
	saveError        error
	processError     error
	updateError      error
	transactionError error
	productIDs       []string
	stockItems       []inventoryusecase.StockItem
	released         bool
	savedOrder       *orderdomain.Order
	updatedOrder     *orderdomain.Order
}

func (dependencies *checkoutDependencies) FindByIDs(_ context.Context, productIDs []string) ([]*catalogdomain.Product, error) {
	dependencies.productIDs = append([]string(nil), productIDs...)
	return dependencies.products, dependencies.findError
}

func (dependencies *checkoutDependencies) Reserve(_ context.Context, stockItems []inventoryusecase.StockItem) error {
	dependencies.stockItems = append([]inventoryusecase.StockItem(nil), stockItems...)
	return dependencies.reserveError
}

func (dependencies *checkoutDependencies) Release(_ context.Context, stockItems []inventoryusecase.StockItem) error {
	dependencies.released = true
	dependencies.stockItems = append([]inventoryusecase.StockItem(nil), stockItems...)
	return dependencies.releaseError
}

func (dependencies *checkoutDependencies) Save(_ context.Context, order *orderdomain.Order) error {
	dependencies.savedOrder = order
	return dependencies.saveError
}

func (dependencies *checkoutDependencies) UpdateStatus(_ context.Context, order *orderdomain.Order) error {
	dependencies.updatedOrder = order
	return dependencies.updateError
}

func (dependencies *checkoutDependencies) Process(_ context.Context, orderID string, amountCents int64) (*paymentdomain.Payment, error) {
	if dependencies.processError != nil {
		return nil, dependencies.processError
	}
	return paymentdomain.NewPayment("payment-1", orderID, amountCents, dependencies.paymentStatus, time.Now())
}

func (dependencies *checkoutDependencies) Execute(applicationContext context.Context, operation func(context.Context) error) error {
	if dependencies.transactionError != nil {
		return dependencies.transactionError
	}
	return operation(applicationContext)
}

func TestCheckoutUseCase(t *testing.T) {
	fixedTime := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	product1, _ := catalogdomain.NewProduct("product-1", "Keyboard", "", 1013, fixedTime)
	product2, _ := catalogdomain.NewProduct("product-2", "Mouse", "", 500, fixedTime)
	inactiveProduct := *product1
	inactiveProduct.Status = catalogdomain.ProductStatusInactive
	dependencyError := errors.New("dependency failed")
	validInput := CheckoutInput{UserID: "user-1", Items: []CheckoutItemInput{{ProductID: product1.ID, Quantity: 1}}}

	for _, testCase := range []struct {
		name         string
		input        CheckoutInput
		dependencies *checkoutDependencies
		wantErr      error
		wantStatus   orderdomain.OrderStatus
		wantReleased bool
	}{
		{name: "empty checkout", input: CheckoutInput{UserID: "user-1"}, dependencies: &checkoutDependencies{}, wantErr: ErrEmptyCheckoutItems},
		{name: "invalid checkout item", input: CheckoutInput{UserID: "user-1", Items: []CheckoutItemInput{{ProductID: "", Quantity: 0}}}, dependencies: &checkoutDependencies{}, wantErr: ErrInvalidCheckoutItem},
		{name: "transaction manager failure", input: validInput, dependencies: &checkoutDependencies{transactionError: dependencyError}, wantErr: dependencyError},
		{name: "product repository failure", input: validInput, dependencies: &checkoutDependencies{findError: dependencyError}, wantErr: dependencyError},
		{name: "missing product by count", input: validInput, dependencies: &checkoutDependencies{products: nil}, wantErr: ErrCheckoutProductNotFound},
		{name: "missing requested product by ID", input: validInput, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product2}}, wantErr: ErrCheckoutProductNotFound},
		{name: "inactive product", input: validInput, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{&inactiveProduct}}, wantErr: ErrInactiveCheckoutProduct},
		{name: "reserve failure", input: validInput, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product1}, reserveError: dependencyError}, wantErr: dependencyError},
		{name: "invalid order", input: CheckoutInput{UserID: "", Items: validInput.Items}, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product1}}, wantErr: orderdomain.ErrEmptyOrderUserID},
		{name: "order save failure", input: validInput, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product1}, saveError: dependencyError}, wantErr: dependencyError},
		{name: "payment failure", input: validInput, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product1}, processError: dependencyError}, wantErr: dependencyError},
		{name: "release failure", input: validInput, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product1}, paymentStatus: paymentdomain.PaymentStatusDeclined, releaseError: dependencyError}, wantErr: dependencyError, wantReleased: true},
		{name: "declined payment", input: validInput, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product1}, paymentStatus: paymentdomain.PaymentStatusDeclined}, wantStatus: orderdomain.OrderStatusFailed, wantReleased: true},
		{name: "status update failure", input: validInput, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product1}, paymentStatus: paymentdomain.PaymentStatusApproved, updateError: dependencyError}, wantErr: dependencyError, wantStatus: orderdomain.OrderStatusPaid},
		{name: "approved aggregated checkout", input: CheckoutInput{UserID: "user-1", Items: []CheckoutItemInput{{ProductID: "product-2", Quantity: 1}, {ProductID: "product-1", Quantity: 2}, {ProductID: "product-2", Quantity: 3}}}, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product2, product1}, paymentStatus: paymentdomain.PaymentStatusApproved}, wantStatus: orderdomain.OrderStatusPaid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			useCase := NewCheckoutUseCase(testCase.dependencies, testCase.dependencies, testCase.dependencies, testCase.dependencies, testCase.dependencies, func() time.Time { return fixedTime })
			output, err := useCase.Execute(context.Background(), testCase.input)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("Execute() error = %v, want %v", err, testCase.wantErr)
			}
			if testCase.dependencies.released != testCase.wantReleased {
				t.Fatalf("released = %v, want %v", testCase.dependencies.released, testCase.wantReleased)
			}
			if testCase.wantErr == nil {
				if output == nil || output.Order.Status != testCase.wantStatus || output.Payment.Status != testCase.dependencies.paymentStatus {
					t.Fatalf("Execute() output = %#v", output)
				}
			}
			if testCase.name == "status update failure" && (testCase.dependencies.updatedOrder == nil || testCase.dependencies.updatedOrder.Status != testCase.wantStatus) {
				t.Fatalf("updated order = %#v", testCase.dependencies.updatedOrder)
			}
			if testCase.name == "approved aggregated checkout" {
				if !reflect.DeepEqual(testCase.dependencies.productIDs, []string{"product-1", "product-2"}) {
					t.Fatalf("product IDs = %#v", testCase.dependencies.productIDs)
				}
				wantStock := []inventoryusecase.StockItem{{ProductID: "product-1", Quantity: 2}, {ProductID: "product-2", Quantity: 4}}
				if !reflect.DeepEqual(testCase.dependencies.stockItems, wantStock) || output.Order.TotalAmountCents != 4026 {
					t.Fatalf("stock = %#v, total = %d", testCase.dependencies.stockItems, output.Order.TotalAmountCents)
				}
			}
		})
	}
}

func TestAggregateCheckoutItemsRejectsNonPositiveQuantity(t *testing.T) {
	if _, _, err := aggregateCheckoutItems([]CheckoutItemInput{{ProductID: "product-1", Quantity: 0}}); !errors.Is(err, ErrInvalidCheckoutItem) {
		t.Fatalf("aggregateCheckoutItems() error = %v", err)
	}
}
