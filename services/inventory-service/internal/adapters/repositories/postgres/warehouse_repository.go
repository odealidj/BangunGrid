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

type PostgresWarehouseRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresWarehouseRepository(pool *pgxpool.Pool) ports.WarehouseRepositoryPort {
	return &PostgresWarehouseRepository{pool: pool}
}

func (r *PostgresWarehouseRepository) Create(ctx context.Context, w *domain.Warehouse) error {
	query := `
		INSERT INTO warehouses (
			id, tenant_id, code, name, type, address, city, postal_code, latitude, longitude, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.pool.Exec(ctx, query,
		w.ID, w.TenantID, w.Code, w.Name, w.Type, w.Address, w.City, w.PostalCode,
		w.Latitude, w.Longitude, w.IsActive, w.CreatedAt, w.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal membuat warehouse: %w", err)
	}
	return nil
}

func (r *PostgresWarehouseRepository) FindByID(ctx context.Context, tenantID, id string) (*domain.Warehouse, error) {
	query := `
		SELECT id, tenant_id, code, name, type, address, city, postal_code, latitude, longitude, is_active, created_at, updated_at
		FROM warehouses
		WHERE id = $1 AND tenant_id = $2
	`
	var w domain.Warehouse
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&w.ID, &w.TenantID, &w.Code, &w.Name, &w.Type, &w.Address, &w.City, &w.PostalCode,
		&w.Latitude, &w.Longitude, &w.IsActive, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWarehouseNotFound
		}
		return nil, fmt.Errorf("gagal mencari warehouse by id: %w", err)
	}
	return &w, nil
}

func (r *PostgresWarehouseRepository) FindByCode(ctx context.Context, tenantID, code string) (*domain.Warehouse, error) {
	query := `
		SELECT id, tenant_id, code, name, type, address, city, postal_code, latitude, longitude, is_active, created_at, updated_at
		FROM warehouses
		WHERE code = $1 AND tenant_id = $2
	`
	var w domain.Warehouse
	err := r.pool.QueryRow(ctx, query, code, tenantID).Scan(
		&w.ID, &w.TenantID, &w.Code, &w.Name, &w.Type, &w.Address, &w.City, &w.PostalCode,
		&w.Latitude, &w.Longitude, &w.IsActive, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWarehouseNotFound
		}
		return nil, fmt.Errorf("gagal mencari warehouse by code: %w", err)
	}
	return &w, nil
}

func (r *PostgresWarehouseRepository) List(ctx context.Context, tenantID string) ([]domain.Warehouse, error) {
	query := `
		SELECT id, tenant_id, code, name, type, address, city, postal_code, latitude, longitude, is_active, created_at, updated_at
		FROM warehouses
		WHERE tenant_id = $1
		ORDER BY name ASC
	`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat daftar warehouse: %w", err)
	}
	defer rows.Close()

	var list []domain.Warehouse
	for rows.Next() {
		var w domain.Warehouse
		if err := rows.Scan(
			&w.ID, &w.TenantID, &w.Code, &w.Name, &w.Type, &w.Address, &w.City, &w.PostalCode,
			&w.Latitude, &w.Longitude, &w.IsActive, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, w)
	}
	return list, nil
}
