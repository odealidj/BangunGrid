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

		`CREATE SEQUENCE IF NOT EXISTS order_no_seq START 1001;`,

		`CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			branch_id UUID NOT NULL,
			order_no VARCHAR(100) NOT NULL,
			customer_id UUID NOT NULL,
			customer_type VARCHAR(50) NOT NULL DEFAULT 'B2C_RETAIL',
			status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
			total_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
			total_weight_kg DOUBLE PRECISION NOT NULL DEFAULT 0,
			total_volume_m3 DOUBLE PRECISION NOT NULL DEFAULT 0,
			unloading_fee NUMERIC(15,2) NOT NULL DEFAULT 0,
			narrow_road_fee NUMERIC(15,2) NOT NULL DEFAULT 0,
			shipping_fee NUMERIC(15,2) NOT NULL DEFAULT 0,
			grand_total NUMERIC(15,2) NOT NULL DEFAULT 0,
			delivery_address TEXT,
			delivery_latitude DOUBLE PRECISION DEFAULT 0,
			delivery_longitude DOUBLE PRECISION DEFAULT 0,
			notes TEXT,
			cancel_reason TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			CONSTRAINT uq_order_tenant_no UNIQUE (tenant_id, order_no)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_orders_tenant_cust ON orders(tenant_id, customer_id);`,
		`CREATE INDEX IF NOT EXISTS idx_orders_tenant_status ON orders(tenant_id, status);`,

		`CREATE TABLE IF NOT EXISTS order_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			sku_id UUID NOT NULL,
			product_name_snapshot VARCHAR(255) NOT NULL,
			uom_code_snapshot VARCHAR(50) NOT NULL,
			quantity DOUBLE PRECISION NOT NULL,
			unit_price_snapshot NUMERIC(15,2) NOT NULL,
			subtotal NUMERIC(15,2) NOT NULL,
			weight_kg_snapshot DOUBLE PRECISION NOT NULL DEFAULT 0,
			volume_m3_snapshot DOUBLE PRECISION NOT NULL DEFAULT 0
		);`,

		`CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);`,

		`CREATE TABLE IF NOT EXISTS b2b_rfqs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			customer_id UUID NOT NULL,
			project_name VARCHAR(255) NOT NULL,
			project_address TEXT NOT NULL,
			boq_document_url TEXT,
			estimated_budget NUMERIC(15,2) DEFAULT 0,
			status VARCHAR(50) NOT NULL DEFAULT 'SUBMITTED',
			notes TEXT,
			quoted_amount NUMERIC(15,2) DEFAULT 0,
			quotation_notes TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);`,

		`CREATE INDEX IF NOT EXISTS idx_b2b_rfqs_tenant_cust ON b2b_rfqs(tenant_id, customer_id);`,

		`CREATE TABLE IF NOT EXISTS b2b_rfq_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			rfq_id UUID NOT NULL REFERENCES b2b_rfqs(id) ON DELETE CASCADE,
			material_name VARCHAR(255) NOT NULL,
			target_uom VARCHAR(50) NOT NULL,
			estimated_qty DOUBLE PRECISION NOT NULL,
			target_price NUMERIC(15,2) DEFAULT 0,
			notes TEXT
		);`,

		`CREATE INDEX IF NOT EXISTS idx_rfq_items_rfq ON b2b_rfq_items(rfq_id);`,

		`CREATE TABLE IF NOT EXISTS b2b_call_off_deliveries (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			sequence_no INT NOT NULL,
			requested_delivery_date DATE NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'SCHEDULED',
			notes TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);`,

		`CREATE INDEX IF NOT EXISTS idx_call_off_order ON b2b_call_off_deliveries(order_id);`,

		`CREATE TABLE IF NOT EXISTS b2b_call_off_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			call_off_id UUID NOT NULL REFERENCES b2b_call_off_deliveries(id) ON DELETE CASCADE,
			sku_id UUID NOT NULL,
			product_name VARCHAR(255) NOT NULL,
			quantity DOUBLE PRECISION NOT NULL,
			uom VARCHAR(50) NOT NULL
		);`,

		`CREATE INDEX IF NOT EXISTS idx_call_off_items_calloff ON b2b_call_off_items(call_off_id);`,
	}

	for _, query := range queries {
		if _, err := pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("gagal menjalankan query migrasi: %v. Query: %s", err, query)
		}
	}

	return nil
}
