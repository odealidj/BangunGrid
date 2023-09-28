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

type PostgresStockRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresStockRepository(pool *pgxpool.Pool) ports.StockRepositoryPort {
	return &PostgresStockRepository{pool: pool}
}

func (r *PostgresStockRepository) Find(ctx context.Context, tenantID, warehouseID, productID string) (*domain.StockItem, error) {
	query := `
		SELECT id, tenant_id, warehouse_id, product_id, uom_code,
		       physical_qty, reserved_qty, available_qty, safety_stock, min_reorder_qty, updated_at
		FROM stock_items
		WHERE tenant_id = $1 AND warehouse_id = $2 AND product_id = $3
	`
	var s domain.StockItem
	err := r.pool.QueryRow(ctx, query, tenantID, warehouseID, productID).Scan(
		&s.ID, &s.TenantID, &s.WarehouseID, &s.ProductID, &s.UOMCode,
		&s.PhysicalQty, &s.ReservedQty, &s.AvailableQty, &s.SafetyStock, &s.MinReorderQty, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrStockNotFound
		}
		return nil, fmt.Errorf("gagal mencari data stok: %w", err)
	}
	return &s, nil
}

func (r *PostgresStockRepository) Upsert(ctx context.Context, item *domain.StockItem) error {
	item.Recalculate()
	query := `
		INSERT INTO stock_items (
			id, tenant_id, warehouse_id, product_id, uom_code,
			physical_qty, reserved_qty, available_qty, safety_stock, min_reorder_qty, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (tenant_id, warehouse_id, product_id)
		DO UPDATE SET
			physical_qty = EXCLUDED.physical_qty,
			reserved_qty = EXCLUDED.reserved_qty,
			available_qty = EXCLUDED.available_qty,
			safety_stock = EXCLUDED.safety_stock,
			min_reorder_qty = EXCLUDED.min_reorder_qty,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.pool.Exec(ctx, query,
		item.ID, item.TenantID, item.WarehouseID, item.ProductID, item.UOMCode,
		item.PhysicalQty, item.ReservedQty, item.AvailableQty, item.SafetyStock, item.MinReorderQty, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal upsert data stok: %w", err)
	}
	return nil
}

func (r *PostgresStockRepository) ListByWarehouse(ctx context.Context, tenantID, warehouseID string) ([]domain.StockItem, error) {
	query := `
		SELECT id, tenant_id, warehouse_id, product_id, uom_code,
		       physical_qty, reserved_qty, available_qty, safety_stock, min_reorder_qty, updated_at
		FROM stock_items
		WHERE tenant_id = $1 AND warehouse_id = $2
		ORDER BY updated_at DESC
	`
	rows, err := r.pool.Query(ctx, query, tenantID, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat daftar stok gudang: %w", err)
	}
	defer rows.Close()

	var list []domain.StockItem
	for rows.Next() {
		var s domain.StockItem
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.WarehouseID, &s.ProductID, &s.UOMCode,
			&s.PhysicalQty, &s.ReservedQty, &s.AvailableQty, &s.SafetyStock, &s.MinReorderQty, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}
