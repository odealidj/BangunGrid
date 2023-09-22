package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AutoMigrate membuat tabel basis data auth-service jika belum ada
func AutoMigrate(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
	CREATE TABLE IF NOT EXISTS organizations (
		id UUID PRIMARY KEY,
		tenant_id UUID NOT NULL,
		name VARCHAR(255) NOT NULL,
		npwp VARCHAR(50),
		nib VARCHAR(50),
		legal_address TEXT,
		contact_email VARCHAR(255),
		contact_phone VARCHAR(50),
		verification_status VARCHAR(50) DEFAULT 'PENDING',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		tenant_id UUID NOT NULL,
		organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
		email VARCHAR(255) NOT NULL,
		phone VARCHAR(50) NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		user_type VARCHAR(50) NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'CUSTOMER',
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		CONSTRAINT uq_tenant_email UNIQUE(tenant_id, email),
		CONSTRAINT uq_tenant_phone UNIQUE(tenant_id, phone)
	);

	CREATE INDEX IF NOT EXISTS idx_users_tenant_email ON users(tenant_id, email);
	CREATE INDEX IF NOT EXISTS idx_users_tenant_phone ON users(tenant_id, phone);
	`

	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("gagal menjalankan migrasi auth-service: %w", err)
	}
	return nil
}
