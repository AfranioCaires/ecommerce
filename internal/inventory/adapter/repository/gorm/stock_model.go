package inventoryrepository

import "time"

type StockModel struct {
	ProductID string `gorm:"primaryKey"`
	Quantity  int    `gorm:"not null"`
	UpdatedAt time.Time
}

func (StockModel) TableName() string {
	return "stocks"
}
