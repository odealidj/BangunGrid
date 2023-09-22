package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"bangungrid/pkg/middleware"
	"bangungrid/pkg/response"
)

func NewRouter(h *AuthHandler, redisClient *redis.Client) http.Handler {
	r := chi.NewRouter()

	// 1. Middleware Standar Chi
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))

	// 2. Ekstraksi Header Gateway ke Context
	r.Use(middleware.ExtractHeaders())

	// 3. Idempotency Middleware (jika Redis terhubung)
	if redisClient != nil {
		r.Use(middleware.Idempotency(redisClient))
	}

	// 4. Healthcheck
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		response.JSONSuccess(w, r, http.StatusOK, "auth-service is healthy", map[string]string{
			"status": "UP",
		})
	})

	// 5. JWKS Endpoint untuk KrakenD Gateway
	r.Get("/.well-known/jwks.json", h.GetJWKS)

	// 6. Rute Autentikasi
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.RefreshToken)
		r.Get("/profile", h.GetProfile)
	})

	return r
}
