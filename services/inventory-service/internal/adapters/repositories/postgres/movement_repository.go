package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"bangungrid/services/inventory-service/internal/core/domain"
	"bangungrid/services/inventory-service/internal/core/ports"
)

type PostgresMovementRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresMovementRepository(pool *pgxpool.Pool) ports.MovementRepositoryPort {
	return &PostgresMovementRepository{pool: pool}
}

func (r *PostgresMovementRepository) Create(ctx context.Context, m *domain.StockMovement) error {
	query := `
		INSERT INTO stock_movements (
			id, tenant_id, warehouse_id, product_id, uom_code, quantity, movement_type, reference_id, notes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		m.ID, m.TenantID, m.WarehouseID, m.ProductID, m.UOMCode,
		m.Quantity, m.MovementType, m.ReferenceID, m.Notes, m.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal mencatat movement stok: %w", err)
	}
	return nil
}

func (r *PostgresMovementRepository) List(ctx context.Context, tenantID, warehouseID, productID string) ([]domain.StockMovement, error) {
	query := `
		SELECT id, tenant_id, warehouse_id, product_id, uom_code, quantity, movement_type, reference_id, notes, created_at
		FROM stock_movements
		WHERE tenant_id = $1 AND warehouse_id = $2 AND product_id = $3
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, tenantID, warehouseID, productID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat daftar movement: %w", err)
	}
	defer rows.Close()

	var list []domain.StockMovement
	for rows.Next() {
		var m domain.StockMovement
		if err := rows.Scan(
			&m.ID, &m.TenantID, &m.WarehouseID, &m.ProductID, &m.UOMCode,
			&m.Quantity, &m.MovementType, &m.ReferenceID, &m.Notes, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}
