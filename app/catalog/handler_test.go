package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockProductsRepository implements ProductsRepositoryInterface for testing,
// unit tests should not depend on the database,
// so we use this mock repository to simulate repository behavior
type MockProductsRepository struct {
	products []models.Product
	getErr   error
}

func (m *MockProductsRepository) GetAllProducts() ([]models.Product, error) {
	return m.products, m.getErr
}

func TestHandleGet(t *testing.T) {
	t.Run("get all products successfully", func(t *testing.T) {
		mockRepo := &MockProductsRepository{
			products: []models.Product{},
			getErr:   nil,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("returns internal server error on repository error", func(t *testing.T) {
		mockRepo := &MockProductsRepository{
			products: nil,
			getErr:   assert.AnError,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest("GET", "/catalog", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
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
			getErr: nil,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var body struct {
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

	t.Run("product without category returns nil category in response", func(t *testing.T) {
		mockRepo := &MockProductsRepository{
			products: []models.Product{
				{
					Code:     "PROD009",
					Price:    decimal.NewFromFloat(49.99),
					Category: nil,
				},
			},
			getErr: nil,
		}

		handler := NewCatalogHandler(mockRepo)
		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var body struct {
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
}
