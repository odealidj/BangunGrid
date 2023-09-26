package ports

import (
	"context"

	"bangungrid/services/catalog-service/internal/core/domain"
)

type CreateCategoryCommand struct {
	TenantID string
	ParentID *string
	Name     string
}

type CreateProductCommand struct {
	TenantID       string
	CategoryID     string
	BrandName      string
	Name           string
	Description    string
	IsBulkMaterial bool
	BaseUoM        string
	Specifications map[string]any
	Conversions    []CreateUOMConversionCommand
	PriceTiers     []CreatePriceTierCommand
}

type CreateUOMConversionCommand struct {
	UOMCode              string
	ConversionMultiplier float64
	WeightKg             float64
	VolumeM3             float64
	RetailPrice          float64
}

type CreatePriceTierCommand struct {
	UOMCode     string
	TierName    string
	MinQuantity float64
	UnitPrice   float64
}

type ListProductsQuery struct {
	TenantID   string
	CategoryID string
	Search     string
	IsBulkOnly *bool
	Page       int
	Limit      int
}

type CatalogUseCasePort interface {
	// Kategori
	CreateCategory(ctx context.Context, cmd CreateCategoryCommand) (*domain.Category, error)
	ListCategories(ctx context.Context, tenantID string) ([]domain.Category, error)

	// Produk & Multi-UoM
	CreateProduct(ctx context.Context, cmd CreateProductCommand) (*domain.Product, error)
	GetProductByID(ctx context.Context, tenantID, productID string) (*domain.Product, error)
	ListProducts(ctx context.Context, query ListProductsQuery) ([]domain.Product, int64, error)
	AddUOMConversion(ctx context.Context, productID string, cmd CreateUOMConversionCommand) (*domain.ProductUOMConversion, error)

	// Material Curah (Bulk Density & Dual Measurement)
	ListBulkDensityMatrices(ctx context.Context) ([]domain.BulkDensityMatrix, error)
	CalculateBulkMaterialVolume(ctx context.Context, materialCode string, volumeM3 float64) (*domain.VolumeCalculationResult, error)
	ValidateWeighbridgeTolerance(ctx context.Context, materialCode string, volumeM3, actualWeightKg float64) (bool, float64, string, error)
}
