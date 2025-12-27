package models

import (
	"context"

	"gorm.io/gorm"
)

// ProductsRepositoryInterface defines the contract for product data access
type ProductsRepositoryInterface interface {
	GetAllProducts(ctx context.Context) ([]Product, error)
	GetProducts(ctx context.Context, opts ProductQueryParameters) ([]Product, int64, error)
	GetProductByCode(ctx context.Context, code string) (*Product, error)
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

func (r *ProductsRepository) GetAllProducts(ctx context.Context) ([]Product, error) {
	var products []Product
	if err := r.db.WithContext(ctx).Preload("Category").Preload("Variants").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductsRepository) GetProducts(ctx context.Context, opts ProductQueryParameters) ([]Product, int64, error) {
	var products []Product
	var total int64

	query := r.db.WithContext(ctx).Model(&Product{}).Preload("Category").Preload("Variants")

	if opts.Category != "" {
		query = query.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.code = ?", opts.Category)
	}

	if opts.PriceLessThan != nil {
		query = query.Where("products.price < ?", opts.PriceLessThan)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(opts.Offset).Limit(opts.Limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductsRepository) GetProductByCode(ctx context.Context, code string) (*Product, error) {
	var product Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Variants").
		Where("code = ?", code).
		First(&product).
		Error

	if err != nil {
		return nil, err
	}

	return &product, nil
}
