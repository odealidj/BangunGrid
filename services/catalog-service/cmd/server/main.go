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

	httpadapter "bangungrid/services/catalog-service/internal/adapters/handlers/http"
	postgresrepo "bangungrid/services/catalog-service/internal/adapters/repositories/postgres"
	"bangungrid/services/catalog-service/internal/core/services"
)

func main() {
	_ = godotenv.Load()

	port := getEnv("PORT", "8082")
	postgresURL := getEnv("DATABASE_URL", "postgres://bangungrid:bangungrid_secret_pass@localhost:5434/db_catalog?sslmode=disable")

	log.Printf("[STARTUP] Inisialisasi BangunGrid catalog-service pada port %s...\n", port)

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
			log.Println("[DATABASE] Berhasil terhubung ke PostgreSQL (db_catalog)")
			if err := postgresrepo.AutoMigrate(ctx, pgPool); err != nil {
				log.Printf("[DATABASE] Gagal migrasi tabel: %v\n", err)
			} else {
				log.Println("[DATABASE] Migrasi tabel catalog database berhasil!")
			}
		}
		pingCancel()
	}

	// 2. Repositories
	catRepo := postgresrepo.NewPostgresCategoryRepository(pgPool)
	prodRepo := postgresrepo.NewPostgresProductRepository(pgPool)
	bulkRepo := postgresrepo.NewPostgresBulkDensityRepository(pgPool)

	// Seed data standar material curah jika koneksi aktif
	if pgPool != nil {
		if err := bulkRepo.SeedDefaults(ctx); err != nil {
			log.Printf("[SEED] Peringatan saat seeding bulk density: %v\n", err)
		} else {
			log.Println("[SEED] Seeding matriks densitas material curah berhasil!")
		}
	}

	// 3. Core Services
	catalogService := services.NewCatalogService(catRepo, prodRepo, bulkRepo)

	// 4. HTTP Handler & Router
	handler := httpadapter.NewCatalogHandler(catalogService)
	router := httpadapter.NewRouter(handler)

	// 5. Server Setup & Graceful Shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[SERVER] catalog-service aktif melayani request di http://localhost:%s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[SERVER] Gagal menjalankan HTTP server: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[SHUTDOWN] Menghentikan catalog-service secara aman (graceful shutdown)...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[SHUTDOWN] Error saat shutdown server: %v\n", err)
	}
	if pgPool != nil {
		pgPool.Close()
	}
	log.Println("[SHUTDOWN] catalog-service berhasil dihentikan.")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
