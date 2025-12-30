package catalog

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Response struct {
	Products []ProductDTO `json:"products"`
	Total    int64        `json:"total"`
	Offset   int          `json:"offset"`
	Limit    int          `json:"limit"`
}

type CategoryDTO struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ProductDTO struct {
	Code     string       `json:"code"`
	Price    float64      `json:"price"`
	Category *CategoryDTO `json:"category,omitempty"`
}

type ProductDetailResponse struct {
	Code     string       `json:"code"`
	Price    float64      `json:"price"`
	Category *CategoryDTO `json:"category,omitempty"`
	Variants []VariantDTO `json:"variants"`
}

type VariantDTO struct {
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
}

type CatalogHandler struct {
	repo models.ProductsRepositoryInterface
}

func NewCatalogHandler(r models.ProductsRepositoryInterface) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	opts, err := parseQueryOptions(r)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	res, total, err := h.repo.GetProducts(r.Context(), opts)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	productDTOs := make([]ProductDTO, len(res))
	for i, p := range res {
		dto := ProductDTO{
			Code:  p.Code,
			Price: p.Price.InexactFloat64(),
		}
		if p.Category != nil {
			dto.Category = &CategoryDTO{
				Code: p.Category.Code,
				Name: p.Category.Name,
			}
		}

		productDTOs[i] = dto
	}

	api.OKResponse(w, Response{
		Products: productDTOs,
		Total:    total,
		Offset:   opts.Offset,
		Limit:    opts.Limit,
	})
}

func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "product code is required")
		return
	}

	product, err := h.repo.GetProductByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			api.ErrorResponse(w, http.StatusNotFound, "product not found")
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Map variants with price inheritance
	variants := make([]VariantDTO, len(product.Variants))
	for i, v := range product.Variants {
		price := v.Price
		if price.IsZero() {
			price = product.Price
		}
		variants[i] = VariantDTO{
			Name:  v.Name,
			SKU:   v.SKU,
			Price: price.InexactFloat64(),
		}
	}

	response := ProductDetailResponse{
		Code:     product.Code,
		Price:    product.Price.InexactFloat64(),
		Variants: variants,
	}

	if product.Category != nil {
		response.Category = &CategoryDTO{
			Code: product.Category.Code,
			Name: product.Category.Name,
		}
	}

	api.OKResponse(w, response)
}

func parseQueryOptions(r *http.Request) (models.ProductQueryParameters, error) {
	opts := models.ProductQueryParameters{
		PaginationQueryParameters: models.PaginationQueryParameters{
			Offset: 0,
			Limit:  10,
		},
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return opts, fmt.Errorf("invalid offset parameter")
		}
		opts.Offset = offset
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			return opts, fmt.Errorf("limit must be between 1 and 100")
		}
		opts.Limit = limit
	}

	if category := r.URL.Query().Get("category"); category != "" {
		categoryNormalized := strings.TrimSpace(strings.ToUpper(category))
		if !validCategory(categoryNormalized) {
			return opts, fmt.Errorf("invalid category %q", category)
		}
		opts.Category = categoryNormalized
	}

	if priceStr := r.URL.Query().Get("price_less_than"); priceStr != "" {
		price, err := decimal.NewFromString(priceStr)
		if err != nil {
			return opts, fmt.Errorf("invalid price_less_than parameter")
		}
		opts.PriceLessThan = &price
	}

	return opts, nil
}

func validCategory(category string) bool {
	allowed := map[string]struct{}{
		"CLOTHING":    {},
		"SHOES":       {},
		"ACCESSORIES": {},
	}
	_, ok := allowed[category]
	return ok
}
