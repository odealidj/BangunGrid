package ports

import (
	"context"

	"bangungrid/services/catalog-service/internal/core/domain"
)

type CategoryRepositoryPort interface {
	Create(ctx context.Context, category *domain.Category) error
	FindByID(ctx context.Context, tenantID, id string) (*domain.Category, error)
	List(ctx context.Context, tenantID string) ([]domain.Category, error)
}

type ProductRepositoryPort interface {
	Create(ctx context.Context, product *domain.Product) error
	FindByID(ctx context.Context, tenantID, id string) (*domain.Product, error)
	List(ctx context.Context, query ListProductsQuery) ([]domain.Product, int64, error)
	AddUOMConversion(ctx context.Context, conv *domain.ProductUOMConversion) error
	AddPriceTier(ctx context.Context, tier *domain.ProductPriceTier) error
}

type BulkDensityRepositoryPort interface {
	FindByCode(ctx context.Context, materialCode string) (*domain.BulkDensityMatrix, error)
	List(ctx context.Context) ([]domain.BulkDensityMatrix, error)
	SeedDefaults(ctx context.Context) error
}
