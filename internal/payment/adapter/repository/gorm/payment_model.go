package paymentrepository

import "time"

type PaymentModel struct {
	ID          string `gorm:"primaryKey"`
	OrderID     string `gorm:"not null;uniqueIndex"`
	AmountCents int64  `gorm:"not null"`
	Status      string `gorm:"not null;index"`
	CreatedAt   time.Time
}

func (PaymentModel) TableName() string {
	return "payments"
}
