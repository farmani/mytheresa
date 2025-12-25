package catalog

import (
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type Response struct {
	Products []ProductDTO `json:"products"`
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

type CatalogHandler struct {
	repo models.ProductsRepositoryInterface
}

func NewCatalogHandler(r models.ProductsRepositoryInterface) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	res, err := h.repo.GetAllProducts()
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Map response
	products := make([]ProductDTO, len(res))
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

		products[i] = dto
	}

	response := Response{
		Products: products,
	}

	api.OKResponse(w, response)
}
