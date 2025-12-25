package models

import (
	"gorm.io/gorm"
)

// ProductsRepositoryInterface defines the contract for product data access
type ProductsRepositoryInterface interface {
	GetAllProducts() ([]Product, error)
}

type ProductsRepository struct {
	db *gorm.DB
}

// Ensure ProductsRepository implements the interface
var _ ProductsRepositoryInterface = (*ProductsRepository)(nil)

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

func (r *ProductsRepository) GetAllProducts() ([]Product, error) {
	var products []Product
	if err := r.db.Preload("Category").Preload("Variants").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}
