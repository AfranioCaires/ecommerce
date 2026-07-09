package authenticationrepository

import "time"

type UserModel struct {
	ID           string `gorm:"primaryKey"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	Roles        string `gorm:"not null"`
	CreatedAt    time.Time
}

func (UserModel) TableName() string {
	return "users"
}
