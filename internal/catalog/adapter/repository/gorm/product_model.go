package catalogrepository

import "time"

type ProductModel struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string
	PriceCents  int64  `gorm:"not null"`
	Status      string `gorm:"not null;index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (ProductModel) TableName() string {
	return "products"
}
