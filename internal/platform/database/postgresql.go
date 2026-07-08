package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgreSQLConnection(dataSource string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dataSource), &gorm.Config{})
}
