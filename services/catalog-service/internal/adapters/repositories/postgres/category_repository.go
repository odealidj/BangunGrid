package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"bangungrid/services/catalog-service/internal/core/domain"
	"bangungrid/services/catalog-service/internal/core/ports"
)

type PostgresCategoryRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresCategoryRepository(pool *pgxpool.Pool) ports.CategoryRepositoryPort {
	return &PostgresCategoryRepository{pool: pool}
}

func (r *PostgresCategoryRepository) Create(ctx context.Context, c *domain.Category) error {
	query := `
		INSERT INTO product_categories (id, tenant_id, parent_id, name, slug, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query, c.ID, c.TenantID, c.ParentID, c.Name, c.Slug, c.CreatedAt)
	if err != nil {
		return fmt.Errorf("gagal membuat kategori: %w", err)
	}
	return nil
}

func (r *PostgresCategoryRepository) FindByID(ctx context.Context, tenantID, id string) (*domain.Category, error) {
	query := `
		SELECT id, tenant_id, parent_id, name, slug, created_at
		FROM product_categories
		WHERE id = $1 AND tenant_id = $2
	`
	var c domain.Category
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&c.ID, &c.TenantID, &c.ParentID, &c.Name, &c.Slug, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("gagal mencari kategori: %w", err)
	}
	return &c, nil
}

func (r *PostgresCategoryRepository) List(ctx context.Context, tenantID string) ([]domain.Category, error) {
	query := `
		SELECT id, tenant_id, parent_id, name, slug, created_at
		FROM product_categories
		WHERE tenant_id = $1
		ORDER BY name ASC
	`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat daftar kategori: %w", err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.TenantID, &c.ParentID, &c.Name, &c.Slug, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}
