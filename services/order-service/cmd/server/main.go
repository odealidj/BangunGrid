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
	"github.com/nats-io/nats.go"
	natsconsumer "bangungrid/services/order-service/internal/adapters/consumers/nats"
	handlers "bangungrid/services/order-service/internal/adapters/handlers/http"
	natspub "bangungrid/services/order-service/internal/adapters/publishers/nats"
	"bangungrid/services/order-service/internal/adapters/repositories/postgres"
	"bangungrid/services/order-service/internal/core/services"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://bangungrid:bangungrid_secret_pass@localhost:5434/db_order?sslmode=disable"
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	log.Printf("[STARTUP] Inisialisasi BangunGrid order-service pada port %s...", port)

	// 1. Koneksi PostgreSQL Pool (db_order)
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
	log.Println("[DATABASE] Berhasil terhubung ke PostgreSQL (db_order)")

	// 2. Eksekusi Migrasi Tabel Order, RFQ, dan Call-Off
	if err := postgres.AutoMigrate(ctx, pgPool); err != nil {
		log.Fatalf("[FATAL] Gagal migrasi tabel order database: %v", err)
	}
	log.Println("[DATABASE] Migrasi tabel order database berhasil!")

	// 3. Koneksi NATS JetStream untuk Asynchronous Saga
	var nc *nats.Conn
	nc, err = nats.Connect(natsURL, nats.Timeout(3*time.Second))
	if err != nil {
		log.Printf("[WARN] NATS tidak tersedia (%v). Event publishing & consumer akan dinonaktifkan", err)
	} else {
		log.Println("[NATS] Berhasil terhubung ke NATS Event Bus")
		defer nc.Close()
	}

	orderPublisher := natspub.NewNatsOrderPublisher(nc)

	// 4. Inisialisasi Repositori
	orderRepo := postgres.NewPostgresOrderRepository(pgPool)
	rfqRepo := postgres.NewPostgresRFQRepository(pgPool)
	callOffRepo := postgres.NewPostgresCallOffRepository(pgPool)

	// 5. Inisialisasi Use Case Service
	orderService := services.NewOrderService(
		orderRepo,
		rfqRepo,
		callOffRepo,
		orderPublisher,
	)

	// 6. Jalankan NATS Saga Consumer
	if nc != nil {
		consumer := natsconsumer.NewInventoryEventConsumer(nc, orderService)
		if err := consumer.Start(context.Background()); err != nil {
			log.Printf("[WARN] Gagal memulai NATS consumer: %v", err)
		}
	}

	// 7. Router & HTTP Handler
	handler := handlers.NewOrderHandler(orderService)
	router := handlers.NewRouter(handler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Graceful Shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[READY] order-service aktif dan siap melayani permintaan HTTP di :%s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] Kesalahan server HTTP: %v", err)
		}
	}()

	<-stopChan
	log.Println("[SHUTDOWN] Menerima sinyal terminasi, mematikan order-service secara anggun...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERROR] Gagal mematikan HTTP server secara anggun: %v", err)
	}

	log.Println("[SHUTDOWN] order-service berhasil dimatikan dengan aman.")
}

