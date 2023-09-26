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

		`CREATE TABLE IF NOT EXISTS product_categories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			parent_id UUID,
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			category_id UUID NOT NULL,
			brand_name VARCHAR(100),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) NOT NULL,
			description TEXT,
			is_bulk_material BOOLEAN DEFAULT FALSE,
			base_uom VARCHAR(50) NOT NULL,
			specifications JSONB DEFAULT '{}'::jsonb,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS product_uom_conversions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			product_id UUID NOT NULL,
			uom_code VARCHAR(50) NOT NULL,
			conversion_multiplier DOUBLE PRECISION NOT NULL,
			weight_kg DOUBLE PRECISION NOT NULL DEFAULT 0,
			volume_m3 DOUBLE PRECISION NOT NULL DEFAULT 0,
			retail_price NUMERIC(15,2) NOT NULL DEFAULT 0
		);`,

		`CREATE TABLE IF NOT EXISTS product_price_tiers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			product_id UUID NOT NULL,
			uom_code VARCHAR(50) NOT NULL,
			tier_name VARCHAR(100) NOT NULL,
			min_quantity DOUBLE PRECISION NOT NULL,
			unit_price NUMERIC(15,2) NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS bulk_density_matrices (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			material_code VARCHAR(50) UNIQUE NOT NULL,
			material_name VARCHAR(100) NOT NULL,
			standard_density_kg_per_m3 DOUBLE PRECISION NOT NULL,
			normal_tolerance_pct DOUBLE PRECISION DEFAULT 2.0
		);`,

		`CREATE INDEX IF NOT EXISTS idx_products_tenant ON products(tenant_id);`,
		`CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);`,
		`CREATE INDEX IF NOT EXISTS idx_categories_tenant ON product_categories(tenant_id);`,
		`CREATE INDEX IF NOT EXISTS idx_uom_conversions_product ON product_uom_conversions(product_id);`,
	}

	for _, q := range queries {
		if _, err := pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("gagal mengeksekusi migrasi catalog: %w", err)
		}
	}

	return nil
}
