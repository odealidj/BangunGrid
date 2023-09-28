package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func AutoMigrate(ctx context.Context, pool *pgxpool.Pool) error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`,

		`CREATE TABLE IF NOT EXISTS warehouses (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			code VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL DEFAULT 'CENTRAL_WAREHOUSE',
			address TEXT,
			city VARCHAR(100),
			postal_code VARCHAR(20),
			latitude DOUBLE PRECISION DEFAULT 0,
			longitude DOUBLE PRECISION DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			CONSTRAINT uq_warehouse_tenant_code UNIQUE (tenant_id, code)
		);`,

		`CREATE TABLE IF NOT EXISTS stock_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
			product_id UUID NOT NULL,
			uom_code VARCHAR(50) NOT NULL,
			physical_qty DOUBLE PRECISION NOT NULL DEFAULT 0,
			reserved_qty DOUBLE PRECISION NOT NULL DEFAULT 0,
			available_qty DOUBLE PRECISION NOT NULL DEFAULT 0,
			safety_stock DOUBLE PRECISION NOT NULL DEFAULT 0,
			min_reorder_qty DOUBLE PRECISION NOT NULL DEFAULT 0,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			CONSTRAINT uq_stock_tenant_wh_prod UNIQUE (tenant_id, warehouse_id, product_id)
		);`,

		`CREATE TABLE IF NOT EXISTS stock_reservations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			warehouse_id UUID NOT NULL,
			product_id UUID NOT NULL,
			order_id VARCHAR(100) NOT NULL,
			quantity DOUBLE PRECISION NOT NULL,
			uom_code VARCHAR(50) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS stock_movements (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			warehouse_id UUID NOT NULL,
			product_id UUID NOT NULL,
			uom_code VARCHAR(50) NOT NULL,
			quantity DOUBLE PRECISION NOT NULL,
			movement_type VARCHAR(50) NOT NULL,
			reference_id VARCHAR(100),
			notes TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);`,

		`CREATE INDEX IF NOT EXISTS idx_stock_items_wh ON stock_items(warehouse_id);`,
		`CREATE INDEX IF NOT EXISTS idx_stock_items_prod ON stock_items(product_id);`,
		`CREATE INDEX IF NOT EXISTS idx_reservations_order ON stock_reservations(order_id);`,
		`CREATE INDEX IF NOT EXISTS idx_reservations_status ON stock_reservations(status);`,
		`CREATE INDEX IF NOT EXISTS idx_movements_prod ON stock_movements(warehouse_id, product_id);`,
	}

	for _, q := range queries {
		if _, err := pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("gagal mengeksekusi migrasi inventory: %w", err)
		}
	}

	return nil
}
