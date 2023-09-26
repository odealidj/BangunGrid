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

type PostgresBulkDensityRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresBulkDensityRepository(pool *pgxpool.Pool) ports.BulkDensityRepositoryPort {
	return &PostgresBulkDensityRepository{pool: pool}
}

func (r *PostgresBulkDensityRepository) FindByCode(ctx context.Context, materialCode string) (*domain.BulkDensityMatrix, error) {
	query := `
		SELECT id, material_code, material_name, standard_density_kg_per_m3, normal_tolerance_pct
		FROM bulk_density_matrices
		WHERE material_code = $1
	`
	var m domain.BulkDensityMatrix
	err := r.pool.QueryRow(ctx, query, materialCode).Scan(
		&m.ID, &m.MaterialCode, &m.MaterialName, &m.StandardDensityKgPerM3, &m.NormalTolerancePct,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBulkMaterialNotFound
		}
		return nil, fmt.Errorf("gagal mencari matriks densitas: %w", err)
	}
	return &m, nil
}

func (r *PostgresBulkDensityRepository) List(ctx context.Context) ([]domain.BulkDensityMatrix, error) {
	query := `
		SELECT id, material_code, material_name, standard_density_kg_per_m3, normal_tolerance_pct
		FROM bulk_density_matrices
		ORDER BY material_code ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat daftar matriks densitas: %w", err)
	}
	defer rows.Close()

	var list []domain.BulkDensityMatrix
	for rows.Next() {
		var m domain.BulkDensityMatrix
		if err := rows.Scan(&m.ID, &m.MaterialCode, &m.MaterialName, &m.StandardDensityKgPerM3, &m.NormalTolerancePct); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

// SeedDefaults menyuntikkan matriks standar SNI & best practice industri konstruksi
func (r *PostgresBulkDensityRepository) SeedDefaults(ctx context.Context) error {
	defaults := []domain.BulkDensityMatrix{
		{MaterialCode: "BULK-SAND-01", MaterialName: "Pasir Pasang Halus", StandardDensityKgPerM3: 1400.0, NormalTolerancePct: 2.0},
		{MaterialCode: "BULK-SAND-02", MaterialName: "Pasir Cor / Beton Kasar", StandardDensityKgPerM3: 1450.0, NormalTolerancePct: 2.5},
		{MaterialCode: "BULK-STN-01", MaterialName: "Batu Split 1-2 (10-20mm)", StandardDensityKgPerM3: 1350.0, NormalTolerancePct: 1.5},
		{MaterialCode: "BULK-STN-02", MaterialName: "Batu Split 2-3 (20-30mm)", StandardDensityKgPerM3: 1380.0, NormalTolerancePct: 1.5},
		{MaterialCode: "BULK-DUST-01", MaterialName: "Abu Batu", StandardDensityKgPerM3: 1300.0, NormalTolerancePct: 2.0},
		{MaterialCode: "BULK-FILL-01", MaterialName: "Tanah Urug / Sirtu", StandardDensityKgPerM3: 1500.0, NormalTolerancePct: 3.0},
	}

	query := `
		INSERT INTO bulk_density_matrices (
			material_code, material_name, standard_density_kg_per_m3, normal_tolerance_pct
		) VALUES ($1, $2, $3, $4)
		ON CONFLICT (material_code) DO NOTHING
	`

	for _, d := range defaults {
		_, err := r.pool.Exec(ctx, query, d.MaterialCode, d.MaterialName, d.StandardDensityKgPerM3, d.NormalTolerancePct)
		if err != nil {
			return fmt.Errorf("gagal seeding bulk density %s: %w", d.MaterialCode, err)
		}
	}

	return nil
}
