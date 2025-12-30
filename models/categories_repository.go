package models

import (
	"context"

	"gorm.io/gorm"
)

// CategoriesRepositoryInterface defines the contract for category data access
type CategoriesRepositoryInterface interface {
	GetAllCategories(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, category *Category) error
}

type CategoriesRepository struct {
	db *gorm.DB
}

// Ensure CategoriesRepository implements the interface
var _ CategoriesRepositoryInterface = (*CategoriesRepository)(nil)

func NewCategoriesRepository(db *gorm.DB) *CategoriesRepository {
	return &CategoriesRepository{db: db}
}

func (r *CategoriesRepository) GetAllCategories(ctx context.Context) ([]Category, error) {
	var categories []Category
	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoriesRepository) CreateCategory(ctx context.Context, category *Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}
