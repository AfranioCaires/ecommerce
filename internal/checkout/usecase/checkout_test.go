package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	authenticationdomain "github.com/afraniocaires/ecommerce/internal/authentication/domain"
	catalogdomain "github.com/afraniocaires/ecommerce/internal/catalog/domain"
	inventoryusecase "github.com/afraniocaires/ecommerce/internal/inventory/usecase"
	orderdomain "github.com/afraniocaires/ecommerce/internal/order/domain"
)

type checkoutDependencies struct {
	products         []*catalogdomain.Product
	customerError    error
	findError        error
	reserveError     error
	saveError        error
	transactionError error
	productIDs       []string
	stockItems       []inventoryusecase.StockItem
	savedOrder       *orderdomain.Order
}

func (dependencies *checkoutDependencies) FindByID(_ context.Context, userID string) (*authenticationdomain.User, error) {
	if dependencies.customerError != nil {
		return nil, dependencies.customerError
	}
	return &authenticationdomain.User{ID: userID}, nil
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
	dependencies.stockItems = append([]inventoryusecase.StockItem(nil), stockItems...)
	return nil
}

func (dependencies *checkoutDependencies) Save(_ context.Context, order *orderdomain.Order) error {
	dependencies.savedOrder = order
	return dependencies.saveError
}

func (dependencies *checkoutDependencies) UpdateStatus(_ context.Context, order *orderdomain.Order) error {
	return nil
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
	}{
		{name: "empty checkout", input: CheckoutInput{UserID: "user-1"}, dependencies: &checkoutDependencies{}, wantErr: ErrEmptyCheckoutItems},
		{name: "invalid checkout item", input: CheckoutInput{UserID: "user-1", Items: []CheckoutItemInput{{ProductID: "", Quantity: 0}}}, dependencies: &checkoutDependencies{}, wantErr: ErrInvalidCheckoutItem},
		{name: "transaction manager failure", input: validInput, dependencies: &checkoutDependencies{transactionError: dependencyError}, wantErr: dependencyError},
		{name: "missing customer", input: validInput, dependencies: &checkoutDependencies{customerError: authenticationdomain.ErrUserNotFound}, wantErr: ErrCheckoutCustomerNotFound},
		{name: "product repository failure", input: validInput, dependencies: &checkoutDependencies{findError: dependencyError}, wantErr: dependencyError},
		{name: "missing product by count", input: validInput, dependencies: &checkoutDependencies{products: nil}, wantErr: ErrCheckoutProductNotFound},
		{name: "missing requested product by ID", input: validInput, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product2}}, wantErr: ErrCheckoutProductNotFound},
		{name: "inactive product", input: validInput, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{&inactiveProduct}}, wantErr: ErrInactiveCheckoutProduct},
		{name: "reserve failure", input: validInput, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product1}, reserveError: dependencyError}, wantErr: dependencyError},
		{name: "invalid order", input: CheckoutInput{UserID: "", Items: validInput.Items}, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product1}}, wantErr: orderdomain.ErrEmptyOrderUserID},
		{name: "order save failure", input: validInput, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product1}, saveError: dependencyError}, wantErr: dependencyError},
		{name: "pending aggregated checkout", input: CheckoutInput{UserID: "user-1", Items: []CheckoutItemInput{{ProductID: "product-2", Quantity: 1}, {ProductID: "product-1", Quantity: 2}, {ProductID: "product-2", Quantity: 3}}}, dependencies: &checkoutDependencies{products: []*catalogdomain.Product{product2, product1}}, wantStatus: orderdomain.OrderStatusPending},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			useCase := NewCheckoutUseCase(testCase.dependencies, testCase.dependencies, testCase.dependencies, testCase.dependencies, testCase.dependencies, func() time.Time { return fixedTime })
			output, err := useCase.Execute(context.Background(), testCase.input)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("Execute() error = %v, want %v", err, testCase.wantErr)
			}
			if testCase.wantErr == nil {
				if output == nil || output.Order.Status != testCase.wantStatus {
					t.Fatalf("Execute() output = %#v", output)
				}
			}
			if testCase.name == "pending aggregated checkout" {
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
