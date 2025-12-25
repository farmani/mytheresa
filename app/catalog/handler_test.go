package catalog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
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
}
