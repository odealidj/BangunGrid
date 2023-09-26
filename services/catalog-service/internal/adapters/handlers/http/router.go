package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"bangungrid/pkg/response"
)

func NewRouter(handler *CatalogHandler) http.Handler {
	r := chi.NewRouter()

	// Global Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Healthcheck
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		response.JSONSuccess(w, r, http.StatusOK, "catalog-service is healthy", map[string]string{
			"status": "UP",
		})
	})

	r.Route("/api/v1/catalog", func(r chi.Router) {
		// Kategori
		r.Post("/categories", handler.CreateCategory)
		r.Get("/categories", handler.ListCategories)

		// Produk & Multi-UoM
		r.Post("/products", handler.CreateProduct)
		r.Get("/products", handler.ListProducts)
		r.Get("/products/{id}", handler.GetProductByID)
		r.Post("/products/{id}/uom", handler.AddUOMConversion)

		// Material Curah (Bulk Materials)
		r.Get("/bulk/matrices", handler.ListBulkDensity)
		r.Post("/bulk/calculate", handler.CalculateBulkVolume)
		r.Post("/bulk/verify-weighbridge", handler.VerifyWeighbridge)
	})

	return r
}
