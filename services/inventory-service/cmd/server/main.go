package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	handlers "bangungrid/services/inventory-service/internal/adapters/handlers/http"
	redislock "bangungrid/services/inventory-service/internal/adapters/locks/redis"
	"bangungrid/services/inventory-service/internal/adapters/repositories/postgres"
	"bangungrid/services/inventory-service/internal/core/services"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://bangungrid:bangungrid_secret_pass@localhost:5434/db_inventory?sslmode=disable"
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost:6380"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	log.Printf("[STARTUP] Inisialisasi BangunGrid inventory-service pada port %s...", port)

	// 1. Koneksi PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pgPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("[FATAL] Gagal menghubungkan PostgreSQL pool: %v", err)
	}
	defer pgPool.Close()

	if err := pgPool.Ping(ctx); err != nil {
		log.Fatalf("[FATAL] Gagal ping PostgreSQL: %v", err)
	}
	log.Println("[DATABASE] Berhasil terhubung ke PostgreSQL (db_inventory)")

	// 2. Eksekusi Migrasi Tabel Inventory
	if err := postgres.AutoMigrate(ctx, pgPool); err != nil {
		log.Fatalf("[FATAL] Gagal migrasi tabel inventory database: %v", err)
	}
	log.Println("[DATABASE] Migrasi tabel inventory database berhasil!")

	// 3. Koneksi Redis untuk Distributed Lock
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPassword,
		DB:       0,
	})
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer pingCancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		log.Printf("[WARN] Redis tidak tersedia (%v), distributed lock akan fallback ke in-memory", err)
	} else {
		log.Println("[REDIS] Berhasil terhubung ke Redis untuk Distributed Stock Lock")
	}
	defer rdb.Close()

	// 4. Inisialisasi Repositori & Lock
	warehouseRepo := postgres.NewPostgresWarehouseRepository(pgPool)
	stockRepo := postgres.NewPostgresStockRepository(pgPool)
	movementRepo := postgres.NewPostgresMovementRepository(pgPool)
	reservationRepo := postgres.NewPostgresReservationRepository(pgPool)
	distLock := redislock.NewRedisDistributedLock(rdb)

	// 5. Inisialisasi Use Case Service
	inventoryService := services.NewInventoryService(
		warehouseRepo,
		stockRepo,
		movementRepo,
		reservationRepo,
		distLock,
		nil, // EventPublisher
	)

	// 6. Router & HTTP Handler
	handler := handlers.NewInventoryHandler(inventoryService)
	router := handlers.NewRouter(handler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 7. Background Server Listener & Graceful Shutdown
	go func() {
		log.Printf("[SERVER] inventory-service aktif melayani request di http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] Kesalahan pada HTTP server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[SHUTDOWN] Memulai pematian server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[FATAL] Server dipaksa berhenti: %v", err)
	}
	log.Println("[SHUTDOWN] inventory-service berhasil dimatikan dengan aman.")
}
