package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"bangungrid/pkg/response"
)

func NewRouter(handler *InventoryHandler) http.Handler {
	r := chi.NewRouter()

	// Global Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Healthcheck
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		response.JSONSuccess(w, r, http.StatusOK, "inventory-service is healthy", map[string]string{
			"status": "UP",
		})
	})

	r.Route("/api/v1/inventory", func(r chi.Router) {
		// Gudang & Stockyard
		r.Post("/warehouses", handler.CreateWarehouse)
		r.Get("/warehouses", handler.ListWarehouses)
		r.Get("/warehouses/{id}", handler.GetWarehouseByID)

		// Manajemen Stok & ATP
		r.Post("/stock", handler.AddStock)
		r.Post("/stock/shrinkage", handler.AdjustShrinkage)
		r.Get("/stock/{warehouseId}", handler.ListStockByWarehouse)
		r.Get("/stock/{warehouseId}/{productId}", handler.GetStock)

		// Reservasi & Distributed Soft-locking
		r.Post("/reservations", handler.ReserveStock)
		r.Post("/reservations/{id}/commit", handler.CommitReservation)
		r.Post("/reservations/{id}/release", handler.ReleaseReservation)

		// Audit Ledger / Movements
		r.Get("/movements", handler.ListMovements)
	})

	return r
}
