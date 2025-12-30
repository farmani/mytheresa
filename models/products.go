package models

import (
	"github.com/shopspring/decimal"
)

// ProductQueryParameters holds filtering parameters includes pagination options
type ProductQueryParameters struct {
	PaginationQueryParameters
	Category      string
	PriceLessThan *decimal.Decimal
}

// Product represents a product in the catalog.
// It includes a unique code and a price.
type Product struct {
	ID         uint            `gorm:"primaryKey"`
	Code       string          `gorm:"uniqueIndex;not null"`
	Price      decimal.Decimal `gorm:"type:decimal(10,2);not null"`
	CategoryID *uint           `gorm:"column:category_id"`
	Category   *Category       `gorm:"foreignKey:CategoryID"`
	Variants   []Variant       `gorm:"foreignKey:ProductID"`
}

func (p *Product) TableName() string {
	return "products"
}
