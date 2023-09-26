package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"bangungrid/services/catalog-service/internal/core/domain"
	"bangungrid/services/catalog-service/internal/core/ports"
)

type PostgresProductRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresProductRepository(pool *pgxpool.Pool) ports.ProductRepositoryPort {
	return &PostgresProductRepository{pool: pool}
}

func (r *PostgresProductRepository) Create(ctx context.Context, p *domain.Product) error {
	specJSON, err := json.Marshal(p.Specifications)
	if err != nil {
		specJSON = []byte("{}")
	}

	query := `
		INSERT INTO products (
			id, tenant_id, category_id, brand_name, name, slug, description,
			is_bulk_material, base_uom, specifications, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = r.pool.Exec(
		ctx, query,
		p.ID, p.TenantID, p.CategoryID, p.BrandName, p.Name, p.Slug, p.Description,
		p.IsBulkMaterial, p.BaseUoM, specJSON, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal membuat produk: %w", err)
	}
	return nil
}

func (r *PostgresProductRepository) AddUOMConversion(ctx context.Context, c *domain.ProductUOMConversion) error {
	query := `
		INSERT INTO product_uom_conversions (
			id, product_id, uom_code, conversion_multiplier, weight_kg, volume_m3, retail_price
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(
		ctx, query,
		c.ID, c.ProductID, c.UOMCode, c.ConversionMultiplier, c.WeightKg, c.VolumeM3, c.RetailPrice,
	)
	if err != nil {
		return fmt.Errorf("gagal menambah konversi uom: %w", err)
	}
	return nil
}

func (r *PostgresProductRepository) AddPriceTier(ctx context.Context, t *domain.ProductPriceTier) error {
	query := `
		INSERT INTO product_price_tiers (
			id, product_id, uom_code, tier_name, min_quantity, unit_price
		) VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(
		ctx, query,
		t.ID, t.ProductID, t.UOMCode, t.TierName, t.MinQuantity, t.UnitPrice,
	)
	if err != nil {
		return fmt.Errorf("gagal menambah tier harga: %w", err)
	}
	return nil
}

func (r *PostgresProductRepository) FindByID(ctx context.Context, tenantID, id string) (*domain.Product, error) {
	query := `
		SELECT id, tenant_id, category_id, brand_name, name, slug, description,
		       is_bulk_material, base_uom, specifications, created_at, updated_at
		FROM products
		WHERE id = $1 AND tenant_id = $2
	`
	var p domain.Product
	var specRaw []byte
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&p.ID, &p.TenantID, &p.CategoryID, &p.BrandName, &p.Name, &p.Slug, &p.Description,
		&p.IsBulkMaterial, &p.BaseUoM, &specRaw, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("gagal mencari produk: %w", err)
	}

	if len(specRaw) > 0 {
		_ = json.Unmarshal(specRaw, &p.Specifications)
	}

	// Muat Konversi UoM
	convs, err := r.loadConversions(ctx, p.ID)
	if err == nil {
		p.UOMConversions = convs
	}

	// Muat Price Tiers
	tiers, err := r.loadPriceTiers(ctx, p.ID)
	if err == nil {
		p.PriceTiers = tiers
	}

	return &p, nil
}

func (r *PostgresProductRepository) List(ctx context.Context, q ports.ListProductsQuery) ([]domain.Product, int64, error) {
	whereClause := "WHERE tenant_id = $1"
	args := []any{q.TenantID}
	argIdx := 2

	if q.CategoryID != "" {
		whereClause += fmt.Sprintf(" AND category_id = $%d", argIdx)
		args = append(args, q.CategoryID)
		argIdx++
	}

	if q.IsBulkOnly != nil {
		whereClause += fmt.Sprintf(" AND is_bulk_material = $%d", argIdx)
		args = append(args, *q.IsBulkOnly)
		argIdx++
	}

	if q.Search != "" {
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR brand_name ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+q.Search+"%")
		argIdx++
	}

	// Hitung total items
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereClause)
	var totalItems int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalItems); err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total produk: %w", err)
	}

	// Ambil data halaman
	offset := (q.Page - 1) * q.Limit
	selectQuery := fmt.Sprintf(`
		SELECT id, tenant_id, category_id, brand_name, name, slug, description,
		       is_bulk_material, base_uom, specifications, created_at, updated_at
		FROM products
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, q.Limit, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal memuat daftar produk: %w", err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		var specRaw []byte
		err := rows.Scan(
			&p.ID, &p.TenantID, &p.CategoryID, &p.BrandName, &p.Name, &p.Slug, &p.Description,
			&p.IsBulkMaterial, &p.BaseUoM, &specRaw, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		if len(specRaw) > 0 {
			_ = json.Unmarshal(specRaw, &p.Specifications)
		}

		convs, _ := r.loadConversions(ctx, p.ID)
		p.UOMConversions = convs

		tiers, _ := r.loadPriceTiers(ctx, p.ID)
		p.PriceTiers = tiers

		products = append(products, p)
	}

	return products, totalItems, nil
}

func (r *PostgresProductRepository) loadConversions(ctx context.Context, productID string) ([]domain.ProductUOMConversion, error) {
	query := `
		SELECT id, product_id, uom_code, conversion_multiplier, weight_kg, volume_m3, retail_price
		FROM product_uom_conversions
		WHERE product_id = $1
		ORDER BY conversion_multiplier ASC
	`
	rows, err := r.pool.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.ProductUOMConversion
	for rows.Next() {
		var c domain.ProductUOMConversion
		if err := rows.Scan(&c.ID, &c.ProductID, &c.UOMCode, &c.ConversionMultiplier, &c.WeightKg, &c.VolumeM3, &c.RetailPrice); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *PostgresProductRepository) loadPriceTiers(ctx context.Context, productID string) ([]domain.ProductPriceTier, error) {
	query := `
		SELECT id, product_id, uom_code, tier_name, min_quantity, unit_price
		FROM product_price_tiers
		WHERE product_id = $1
		ORDER BY min_quantity ASC
	`
	rows, err := r.pool.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.ProductPriceTier
	for rows.Next() {
		var t domain.ProductPriceTier
		if err := rows.Scan(&t.ID, &t.ProductID, &t.UOMCode, &t.TierName, &t.MinQuantity, &t.UnitPrice); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}
