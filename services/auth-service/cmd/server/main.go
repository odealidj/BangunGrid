package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	natsclient "github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	httpadapter "bangungrid/services/auth-service/internal/adapters/handlers/http"
	natspublisher "bangungrid/services/auth-service/internal/adapters/publishers/nats"
	postgresrepo "bangungrid/services/auth-service/internal/adapters/repositories/postgres"
	"bangungrid/services/auth-service/internal/adapters/security/hasher"
	"bangungrid/services/auth-service/internal/adapters/security/jwt"
	"bangungrid/services/auth-service/internal/core/services"
)

func main() {
	_ = godotenv.Load()

	port := getEnv("PORT", "8081")
	postgresURL := getEnv("DATABASE_URL", "postgres://bangungrid:bangungrid_secret_pass@localhost:5432/db_auth?sslmode=disable")
	redisHost := getEnv("REDIS_HOST", "localhost:6379")
	redisPass := getEnv("REDIS_PASSWORD", "redis_secret_pass")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

	log.Printf("[STARTUP] Inisialisasi BangunGrid auth-service pada port %s...\n", port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Inisialisasi PostgreSQL Connection Pool
	pgPool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		log.Printf("[DATABASE] Peringatan: Gagal membuat pool PostgreSQL: %v\n", err)
	} else {
		pingCtx, pingCancel := context.WithTimeout(ctx, 3*time.Second)
		if err := pgPool.Ping(pingCtx); err != nil {
			log.Printf("[DATABASE] Peringatan: Tidak dapat terhubung ke PostgreSQL: %v\n", err)
		} else {
			log.Println("[DATABASE] Berhasil terhubung ke PostgreSQL (db_auth)")
			if err := postgresrepo.AutoMigrate(ctx, pgPool); err != nil {
				log.Printf("[DATABASE] Gagal migrasi tabel: %v\n", err)
			} else {
				log.Println("[DATABASE] Migrasi tabel database berhasil!")
			}
		}
		pingCancel()
	}

	// 2. Inisialisasi Redis Client
	var redisClient *redis.Client
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPass,
		DB:       0,
	})
	rPingCtx, rPingCancel := context.WithTimeout(ctx, 2*time.Second)
	if err := rdb.Ping(rPingCtx).Err(); err != nil {
		log.Printf("[REDIS] Peringatan: Tidak dapat terhubung ke Redis: %v\n", err)
	} else {
		log.Println("[REDIS] Berhasil terhubung ke Redis")
		redisClient = rdb
	}
	rPingCancel()

	// 3. Inisialisasi NATS JetStream
	var js natsclient.JetStreamContext
	nc, err := natsclient.Connect(natsURL, natsclient.Timeout(2*time.Second))
	if err != nil {
		log.Printf("[NATS] Peringatan: Tidak dapat terhubung ke NATS: %v\n", err)
	} else {
		defer nc.Close()
		jsCtx, err := nc.JetStream()
		if err != nil {
			log.Printf("[NATS] Gagal menginisialisasi JetStream: %v\n", err)
		} else {
			log.Println("[NATS] Berhasil terhubung ke NATS JetStream")
			js = jsCtx
		}
	}

	// 4. Inisialisasi Adapters
	passwordHasher := hasher.NewBcryptHasher(10)
	tokenService, err := jwt.NewRSAJWTService(24, 7)
	if err != nil {
		log.Fatalf("[SECURITY] Gagal menginisialisasi RSA JWT Service: %v", err)
	}

	var userRepo *postgresrepo.UserRepository
	var orgRepo *postgresrepo.OrganizationRepository
	if pgPool != nil {
		userRepo = postgresrepo.NewUserRepository(pgPool).(*postgresrepo.UserRepository)
		orgRepo = postgresrepo.NewOrganizationRepository(pgPool).(*postgresrepo.OrganizationRepository)
	}
	eventPublisher := natspublisher.NewAuthNatsPublisher(js)

	// 5. Inisialisasi Core Use Case Service
	authService := services.NewAuthService(
		userRepo,
		orgRepo,
		tokenService,
		passwordHasher,
		eventPublisher,
	)

	// 6. Inisialisasi HTTP Handler & Router
	authHandler := httpadapter.NewAuthHandler(authService)
	router := httpadapter.NewRouter(authHandler, redisClient)

	// 7. Menjalankan Server dengan Graceful Shutdown
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[SERVER] auth-service aktif melayani request di http://localhost:%s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[SERVER] Gagal menjalankan server: %v", err)
		}
	}()

	// Menangkap sinyal shutdown OS
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[SHUTDOWN] Menerima sinyal berhenti, mematikan server secara aman...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[SHUTDOWN] Server dipaksa berhenti: %v", err)
	}

	if pgPool != nil {
		pgPool.Close()
	}
	if redisClient != nil {
		_ = redisClient.Close()
	}

	log.Println("[SHUTDOWN] auth-service telah berhasil dihentikan dengan aman.")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
