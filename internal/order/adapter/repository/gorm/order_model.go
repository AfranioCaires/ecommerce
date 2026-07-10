package orderrepository

import "time"

type OrderModel struct {
	ID               string `gorm:"primaryKey"`
	UserID           string `gorm:"not null;index"`
	TotalAmountCents int64  `gorm:"not null"`
	Status           string `gorm:"not null;index"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Items            []OrderItemModel `gorm:"foreignKey:OrderID"`
}

func (OrderModel) TableName() string {
	return "orders"
}

type OrderItemModel struct {
	ID             uint   `gorm:"primaryKey"`
	OrderID        string `gorm:"not null;index"`
	ProductID      string `gorm:"not null"`
	ProductName    string `gorm:"not null"`
	UnitPriceCents int64  `gorm:"not null"`
	Quantity       int    `gorm:"not null"`
}

func (OrderItemModel) TableName() string {
	return "order_items"
}
