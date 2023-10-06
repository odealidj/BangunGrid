package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"bangungrid/pkg/response"
)

func NewRouter(handler *OrderHandler) http.Handler {
	r := chi.NewRouter()

	// Global Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Healthcheck
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		response.JSONSuccess(w, r, http.StatusOK, "order-service is healthy", map[string]string{
			"status": "UP",
		})
	})

	r.Route("/api/v1/orders", func(r chi.Router) {
		// 1. Estimasi Bobot, Volume & Logistik (Kuli/Armada/Langsir)
		r.Post("/estimate", handler.EstimateOrder)

		// 2. B2B RFQ (Request for Quotation / Tender Proyek)
		r.Post("/rfq", handler.CreateRFQ)
		r.Get("/rfq", handler.ListRFQs)
		r.Get("/rfq/{id}", handler.GetRFQByID)
		r.Patch("/rfq/{id}", handler.UpdateRFQStatus)

		// 3. Update Status Parsial Pengiriman (Call-Off)
		r.Patch("/call-offs/{deliveryId}/status", handler.UpdateCallOffStatus)

		// 4. Sales Orders (B2C & B2B)
		r.Post("/", handler.CreateOrder)
		r.Get("/", handler.ListOrders)
		r.Get("/number/{orderNo}", handler.GetOrderByOrderNo)
		r.Get("/{id}", handler.GetOrderByID)
		r.Patch("/{id}/status", handler.UpdateOrderStatus)
		r.Post("/{id}/cancel", handler.CancelOrder)

		// 5. Call-Off Deliveries berelasi dengan Order
		r.Post("/{id}/call-offs", handler.CreateCallOffDelivery)
		r.Get("/{id}/call-offs", handler.ListCallOffDeliveries)
	})

	return r
}

