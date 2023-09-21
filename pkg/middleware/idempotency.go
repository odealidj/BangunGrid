package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"bangungrid/pkg/response"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *responseRecorder) Write(b []byte) (int, error) {
	rec.body.Write(b)
	return rec.ResponseWriter.Write(b)
}

// Idempotency menyediakan middleware anti-duplikasi request berbasis Redis
func Idempotency(redisClient *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Idempotency hanya dipaksakan untuk method mutasi
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
				next.ServeHTTP(w, r)
				return
			}

			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			lockKey := fmt.Sprintf("idempotency:%s", requestID)

			// Kunci request di Redis selama 120 detik
			acquired, err := redisClient.SetNX(ctx, lockKey, "IN_PROGRESS", 120*time.Second).Result()
			if err != nil {
				response.JSONError(w, r, http.StatusInternalServerError, "Gagal memproses idempotency", "REDIS_ERROR", nil)
				return
			}

			if !acquired {
				val, _ := redisClient.Get(ctx, lockKey).Result()
				if val == "IN_PROGRESS" {
					response.JSONError(w, r, http.StatusConflict, "Permintaan sedang diproses, silakan tunggu", "DUPLICATE_REQUEST_IN_PROGRESS", nil)
					return
				}

				// Kembalikan response body yang tersimpan
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Idempotent-Replay", "true")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(val))
				return
			}

			// Rekam respons asli untuk disimpan ke Redis
			rec := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           bytes.NewBuffer(nil),
			}

			next.ServeHTTP(rec, r)

			// Jika berhasil (2xx), simpan respons di Redis selama 24 jam
			if rec.statusCode >= 200 && rec.statusCode < 300 {
				_ = redisClient.Set(ctx, lockKey, rec.body.String(), 24*time.Hour).Err()
			} else {
				// Jika gagal, lepas lock agar klien bisa mencoba kembali
				_ = redisClient.Del(ctx, lockKey).Err()
			}
		})
	}
}
