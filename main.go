package database

import (
	"github.com/mikemeh/ecommerce-api_V2/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(dbURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&models.User{}, &models.Product{}, &models.Order{}, &models.OrderItem{}); err != nil {
		return nil, err
	}

	return db, nil
}
