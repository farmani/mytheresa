package models

import "gorm.io/gorm"

// CategoriesRepositoryInterface defines the contract for category data access
type CategoriesRepositoryInterface interface {
}

type CategoriesRepository struct {
	db *gorm.DB
}

// Ensure CategoriesRepository implements the interface
var _ CategoriesRepositoryInterface = (*CategoriesRepository)(nil)

func NewCategoriesRepository(db *gorm.DB) *CategoriesRepository {
	return &CategoriesRepository{db: db}
}
