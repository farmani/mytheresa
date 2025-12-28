package categories

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCategoriesRepository implements CategoriesRepositoryInterface for testing
type MockCategoriesRepository struct {
	categories []models.Category
	createErr  error
	getAllErr  error
}

func (m *MockCategoriesRepository) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	return m.categories, nil
}

func (m *MockCategoriesRepository) CreateCategory(ctx context.Context, cat *models.Category) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.categories = append(m.categories, *cat)
	return nil
}

func TestHandleGetAll(t *testing.T) {
	t.Run("returns all categories", func(t *testing.T) {
		mockRepo := &MockCategoriesRepository{
			categories: []models.Category{
				{ID: 1, Code: "CLOTHING", Name: "Clothing"},
				{ID: 2, Code: "SHOES", Name: "Shoes"},
				{ID: 3, Code: "ACCESSORIES", Name: "Accessories"},
			},
		}

		handler := NewCategoriesHandler(mockRepo)
		req := httptest.NewRequest("GET", "/categories", nil)
		rec := httptest.NewRecorder()

		handler.HandleGetAll(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Body.String(), "CLOTHING")
		assert.Contains(t, rec.Body.String(), "SHOES")
		assert.Contains(t, rec.Body.String(), "ACCESSORIES")

		var body struct {
			Categories []struct {
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"categories"`
		}

		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.NotEmpty(t, body.Categories)
		require.Len(t, body.Categories, 3)
		for i, exp := range mockRepo.categories {
			assert.Equal(t, exp.Code, body.Categories[i].Code)
			assert.Equal(t, exp.Name, body.Categories[i].Name)
		}
	})

	t.Run("returns empty list when no categories", func(t *testing.T) {
		mockRepo := &MockCategoriesRepository{
			categories: []models.Category{},
		}

		handler := NewCategoriesHandler(mockRepo)
		req := httptest.NewRequest("GET", "/categories", nil)
		rec := httptest.NewRecorder()

		handler.HandleGetAll(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"categories":[]`)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockRepo := &MockCategoriesRepository{
			getAllErr: errors.New("database error"),
		}

		handler := NewCategoriesHandler(mockRepo)
		req := httptest.NewRequest("GET", "/categories", nil)
		rec := httptest.NewRecorder()

		handler.HandleGetAll(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "database error")
	})
}

func TestHandleCreate(t *testing.T) {
	t.Run("creates category successfully", func(t *testing.T) {
		mockRepo := &MockCategoriesRepository{
			categories: []models.Category{},
		}

		handler := NewCategoriesHandler(mockRepo)
		body := bytes.NewBufferString(`{"code":"ELECTRONICS","name":"Electronics"}`)
		req := httptest.NewRequest("POST", "/categories", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.HandleCreate(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Body.String(), "ELECTRONICS")
		assert.Contains(t, rec.Body.String(), "Electronics")
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		mockRepo := &MockCategoriesRepository{}

		handler := NewCategoriesHandler(mockRepo)
		body := bytes.NewBufferString(`{invalid json}`)
		req := httptest.NewRequest("POST", "/categories", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.HandleCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid request body")
	})

	t.Run("returns error when code is missing", func(t *testing.T) {
		mockRepo := &MockCategoriesRepository{}

		handler := NewCategoriesHandler(mockRepo)
		body := bytes.NewBufferString(`{"name":"Electronics"}`)
		req := httptest.NewRequest("POST", "/categories", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.HandleCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "code and name are required")
	})

	t.Run("returns error when name is missing", func(t *testing.T) {
		mockRepo := &MockCategoriesRepository{}

		handler := NewCategoriesHandler(mockRepo)
		body := bytes.NewBufferString(`{"code":"ELECTRONICS"}`)
		req := httptest.NewRequest("POST", "/categories", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.HandleCreate(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "code and name are required")
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockRepo := &MockCategoriesRepository{
			createErr: errors.New("database error"),
		}

		handler := NewCategoriesHandler(mockRepo)
		body := bytes.NewBufferString(`{"code":"ELECTRONICS","name":"Electronics"}`)
		req := httptest.NewRequest("POST", "/categories", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.HandleCreate(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "database error")
	})
}
