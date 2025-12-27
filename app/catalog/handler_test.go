package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// MockProductsRepository implements ProductsRepositoryInterface for testing,
// unit tests should not depend on the database,
// so we use this mock repository to simulate repository behavior
type MockProductsRepository struct {
	products      []models.Product
	total         int64
	productByCode *models.Product
	getErr        error
	getByCodeErr  error
}

func (m *MockProductsRepository) GetAllProducts(ctx context.Context) ([]models.Product, error) {
	return m.products, m.getErr
}

func (m *MockProductsRepository) GetProducts(ctx context.Context, opts models.ProductQueryParameters) ([]models.Product, int64, error) {
	if m.getErr != nil {
		return nil, 0, m.getErr
	}

	return m.products, m.total, nil
}

func (m *MockProductsRepository) GetProductByCode(ctx context.Context, code string) (*models.Product, error) {
	if m.getByCodeErr != nil {
		return nil, m.getByCodeErr
	}
	return m.productByCode, nil
}

func TestHandleGet(t *testing.T) {
	t.Run("get all products successfully", func(t *testing.T) {
		mockRepo := &MockProductsRepository{
			products: []models.Product{},
			total:    0,
			getErr:   nil,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockRepo := &MockProductsRepository{
			getErr: errors.New("database error"),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "database error")
	})

	t.Run("includes category in response when product has category", func(t *testing.T) {
		mockRepo := &MockProductsRepository{
			products: []models.Product{
				{
					Code:  "PROD001",
					Price: decimal.NewFromFloat(99.99),
					Category: &models.Category{
						Code: "CLOTHING",
						Name: "Clothing",
					},
				},
			},
			total:  1,
			getErr: nil,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var body struct {
			Total    int64 `json:"total"`
			Offset   int   `json:"offset"`
			Limit    int   `json:"limit"`
			Products []struct {
				Category *struct {
					Code string `json:"code"`
					Name string `json:"name"`
				} `json:"category"`
			} `json:"products"`
		}

		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.NotEmpty(t, body.Products)
		require.NotNil(t, body.Products[0].Category)

		assert.Equal(t, "CLOTHING", body.Products[0].Category.Code)
		assert.Equal(t, "Clothing", body.Products[0].Category.Name)
	})

	t.Run("return product without category", func(t *testing.T) {
		mockRepo := &MockProductsRepository{
			products: []models.Product{
				{
					Code:  "PROD009",
					Price: decimal.NewFromFloat(49.99),
				},
			},
			total:  1,
			getErr: nil,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var body struct {
			Total    int64 `json:"total"`
			Offset   int   `json:"offset"`
			Limit    int   `json:"limit"`
			Products []struct {
				Category *struct {
					Code string `json:"code"`
					Name string `json:"name"`
				} `json:"category"`
			} `json:"products"`
		}

		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.Len(t, body.Products, 1)
		assert.Nil(t, body.Products[0].Category)
	})

	t.Run("uses default pagination values", func(t *testing.T) {
		mockRepo := &MockProductsRepository{
			products: []models.Product{},
			total:    0,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"offset":0`)
		assert.Contains(t, rec.Body.String(), `"limit":10`)
	})

	t.Run("returns paginated products with category", func(t *testing.T) {
		mockRepo := &MockProductsRepository{
			products: []models.Product{
				{
					Code:  "PROD001",
					Price: decimal.NewFromFloat(10.99),
					Category: &models.Category{
						Code: "CLOTHING",
						Name: "Clothing",
					},
				},
			},
			total: 1,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog?limit=5&offset=0", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var body struct {
			Total    int64 `json:"total"`
			Offset   int   `json:"offset"`
			Limit    int   `json:"limit"`
			Products []struct {
				Category *struct {
					Code string `json:"code"`
					Name string `json:"name"`
				} `json:"category"`
			} `json:"products"`
		}

		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.Len(t, body.Products, 1)
		assert.Equal(t, "CLOTHING", body.Products[0].Category.Code)
		assert.Equal(t, "Clothing", body.Products[0].Category.Name)
		assert.Equal(t, int64(1), body.Total)
		assert.Equal(t, 0, body.Offset)
		assert.Equal(t, 5, body.Limit)
	})

	t.Run("validates limit must be between 1 and 100", func(t *testing.T) {
		mockRepo := &MockProductsRepository{}
		handler := NewCatalogHandler(mockRepo)

		// Test limit > 100
		req := httptest.NewRequest("GET", "/catalog?limit=101", nil)
		rec := httptest.NewRecorder()
		handler.HandleGet(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "limit must be between 1 and 100")

		// Test limit < 1
		req = httptest.NewRequest("GET", "/catalog?limit=0", nil)
		rec = httptest.NewRecorder()
		handler.HandleGet(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "limit must be between 1 and 100")
	})

	t.Run("validates offset must be non-negative", func(t *testing.T) {
		mockRepo := &MockProductsRepository{}
		handler := NewCatalogHandler(mockRepo)

		req := httptest.NewRequest("GET", "/catalog?offset=-1", nil)
		rec := httptest.NewRecorder()
		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid offset parameter")
	})

	t.Run("validates category parameter", func(t *testing.T) {
		mockRepo := &MockProductsRepository{}
		handler := NewCatalogHandler(mockRepo)

		req := httptest.NewRequest("GET", "/catalog?category=Hats", nil)
		rec := httptest.NewRecorder()
		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var body struct {
			Error string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.Equal(t, `invalid category "Hats"`, body.Error)
	})

	t.Run("validates price_less_than parameter", func(t *testing.T) {
		mockRepo := &MockProductsRepository{}
		handler := NewCatalogHandler(mockRepo)

		req := httptest.NewRequest("GET", "/catalog?price_less_than=invalid", nil)
		rec := httptest.NewRecorder()
		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid price_less_than parameter")
	})
}

func TestHandleGetByCode(t *testing.T) {
	t.Run("returns product with variants", func(t *testing.T) {
		mockRepo := &MockProductsRepository{
			productByCode: &models.Product{
				Code:  "PROD001",
				Price: decimal.NewFromFloat(100.00),
				Category: &models.Category{
					Code: "CLOTHING",
					Name: "Clothing",
				},
				Variants: []models.Variant{
					{
						Name:  "Small",
						SKU:   "PROD001-S",
						Price: decimal.NewFromFloat(95.00),
					},
					{
						Name:  "Medium",
						SKU:   "PROD001-M",
						Price: decimal.NewFromFloat(100.00),
					},
				},
			},
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog/PROD001", nil)
		req.SetPathValue("code", "PROD001")
		rec := httptest.NewRecorder()

		handler.HandleGetByCode(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "PROD001")
		assert.Contains(t, rec.Body.String(), "CLOTHING")
		assert.Contains(t, rec.Body.String(), "PROD001-S")
		assert.Contains(t, rec.Body.String(), "PROD001-M")
	})

	t.Run("variant inherits product price when no variant price", func(t *testing.T) {
		mockRepo := &MockProductsRepository{
			productByCode: &models.Product{
				Code:  "PROD001",
				Price: decimal.NewFromFloat(100.00),
				Variants: []models.Variant{
					{
						Name:  "Default",
						SKU:   "PROD001-DEF",
						Price: decimal.Decimal{}, // Zero value - should inherit
					},
				},
			},
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog/PROD001", nil)
		req.SetPathValue("code", "PROD001")
		rec := httptest.NewRecorder()

		handler.HandleGetByCode(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"price":100`)
	})

	t.Run("returns 404 when product not found", func(t *testing.T) {
		mockRepo := &MockProductsRepository{
			getByCodeErr: gorm.ErrRecordNotFound,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog/NOTFOUND", nil)
		req.SetPathValue("code", "NOTFOUND")
		rec := httptest.NewRecorder()

		handler.HandleGetByCode(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "product not found")
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockRepo := &MockProductsRepository{
			getByCodeErr: errors.New("database error"),
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog/PROD001", nil)
		req.SetPathValue("code", "PROD001")
		rec := httptest.NewRecorder()

		handler.HandleGetByCode(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "database error")
	})

	t.Run("returns error when code is empty", func(t *testing.T) {
		mockRepo := &MockProductsRepository{}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog/", nil)
		// Not setting path value simulates empty code
		rec := httptest.NewRecorder()

		handler.HandleGetByCode(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "product code is required")
	})
}
