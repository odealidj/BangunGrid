package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"bangungrid/services/inventory-service/internal/core/domain"
	"bangungrid/services/inventory-service/internal/core/ports"
)

type PostgresReservationRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresReservationRepository(pool *pgxpool.Pool) ports.ReservationRepositoryPort {
	return &PostgresReservationRepository{pool: pool}
}

func (r *PostgresReservationRepository) Create(ctx context.Context, res *domain.StockReservation) error {
	query := `
		INSERT INTO stock_reservations (
			id, tenant_id, warehouse_id, product_id, order_id, quantity, uom_code, status, expires_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		res.ID, res.TenantID, res.WarehouseID, res.ProductID, res.OrderID,
		res.Quantity, res.UOMCode, res.Status, res.ExpiresAt, res.CreatedAt, res.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal mencatat reservasi: %w", err)
	}
	return nil
}

func (r *PostgresReservationRepository) FindByID(ctx context.Context, tenantID, id string) (*domain.StockReservation, error) {
	query := `
		SELECT id, tenant_id, warehouse_id, product_id, order_id, quantity, uom_code, status, expires_at, created_at, updated_at
		FROM stock_reservations
		WHERE id = $1 AND tenant_id = $2
	`
	var res domain.StockReservation
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&res.ID, &res.TenantID, &res.WarehouseID, &res.ProductID, &res.OrderID,
		&res.Quantity, &res.UOMCode, &res.Status, &res.ExpiresAt, &res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrReservationNotFound
		}
		return nil, fmt.Errorf("gagal mencari reservasi: %w", err)
	}
	return &res, nil
}

func (r *PostgresReservationRepository) Update(ctx context.Context, res *domain.StockReservation) error {
	query := `
		UPDATE stock_reservations
		SET status = $1, updated_at = $2
		WHERE id = $3 AND tenant_id = $4
	`
	_, err := r.pool.Exec(ctx, query, res.Status, res.UpdatedAt, res.ID, res.TenantID)
	if err != nil {
		return fmt.Errorf("gagal update status reservasi: %w", err)
	}
	return nil
}
