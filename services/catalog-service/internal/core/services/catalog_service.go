package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"bangungrid/services/catalog-service/internal/core/domain"
	"bangungrid/services/catalog-service/internal/core/ports"
)

type CatalogService struct {
	categoryRepo    ports.CategoryRepositoryPort
	productRepo     ports.ProductRepositoryPort
	bulkDensityRepo ports.BulkDensityRepositoryPort
}

func NewCatalogService(
	catRepo ports.CategoryRepositoryPort,
	prodRepo ports.ProductRepositoryPort,
	bulkRepo ports.BulkDensityRepositoryPort,
) ports.CatalogUseCasePort {
	return &CatalogService{
		categoryRepo:    catRepo,
		productRepo:     prodRepo,
		bulkDensityRepo: bulkRepo,
	}
}

// ==========================================
// Kategori
// ==========================================

func (s *CatalogService) CreateCategory(ctx context.Context, cmd ports.CreateCategoryCommand) (*domain.Category, error) {
	category := &domain.Category{
		ID:        uuid.New().String(),
		TenantID:  cmd.TenantID,
		ParentID:  cmd.ParentID,
		Name:      cmd.Name,
		CreatedAt: time.Now().UTC(),
	}

	if err := category.Validate(); err != nil {
		return nil, err
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CatalogService) ListCategories(ctx context.Context, tenantID string) ([]domain.Category, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantRequired
	}
	return s.categoryRepo.List(ctx, tenantID)
}

// ==========================================
// Produk & Multi-UoM
// ==========================================

func (s *CatalogService) CreateProduct(ctx context.Context, cmd ports.CreateProductCommand) (*domain.Product, error) {
	now := time.Now().UTC()
	prodID := uuid.New().String()

	product := &domain.Product{
		ID:             prodID,
		TenantID:       cmd.TenantID,
		CategoryID:     cmd.CategoryID,
		BrandName:      strings.TrimSpace(cmd.BrandName),
		Name:           strings.TrimSpace(cmd.Name),
		Description:    cmd.Description,
		IsBulkMaterial: cmd.IsBulkMaterial,
		BaseUoM:        strings.ToUpper(strings.TrimSpace(cmd.BaseUoM)),
		Specifications: cmd.Specifications,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := product.Validate(); err != nil {
		return nil, err
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	// Simpan konversi UoM (Piramida Satuan)
	for _, conv := range cmd.Conversions {
		if conv.ConversionMultiplier <= 0 {
			conv.ConversionMultiplier = 1.0
		}
		uomObj := &domain.ProductUOMConversion{
			ID:                   uuid.New().String(),
			ProductID:            prodID,
			UOMCode:              strings.ToUpper(strings.TrimSpace(conv.UOMCode)),
			ConversionMultiplier: conv.ConversionMultiplier,
			WeightKg:             conv.WeightKg,
			VolumeM3:             conv.VolumeM3,
			RetailPrice:          conv.RetailPrice,
		}
		if err := s.productRepo.AddUOMConversion(ctx, uomObj); err != nil {
			return nil, fmt.Errorf("gagal menambahkan uom %s: %w", conv.UOMCode, err)
		}
		product.UOMConversions = append(product.UOMConversions, *uomObj)
	}

	// Simpan Tier Harga
	for _, tier := range cmd.PriceTiers {
		tierObj := &domain.ProductPriceTier{
			ID:          uuid.New().String(),
			ProductID:   prodID,
			UOMCode:     strings.ToUpper(strings.TrimSpace(tier.UOMCode)),
			TierName:    tier.TierName,
			MinQuantity: tier.MinQuantity,
			UnitPrice:   tier.UnitPrice,
		}
		if err := s.productRepo.AddPriceTier(ctx, tierObj); err != nil {
			return nil, fmt.Errorf("gagal menambahkan tier harga: %w", err)
		}
		product.PriceTiers = append(product.PriceTiers, *tierObj)
	}

	return product, nil
}

func (s *CatalogService) GetProductByID(ctx context.Context, tenantID, productID string) (*domain.Product, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantRequired
	}
	return s.productRepo.FindByID(ctx, tenantID, productID)
}

func (s *CatalogService) ListProducts(ctx context.Context, query ports.ListProductsQuery) ([]domain.Product, int64, error) {
	if query.TenantID == "" {
		return nil, 0, domain.ErrTenantRequired
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 20
	}
	return s.productRepo.List(ctx, query)
}

func (s *CatalogService) AddUOMConversion(ctx context.Context, productID string, cmd ports.CreateUOMConversionCommand) (*domain.ProductUOMConversion, error) {
	if cmd.ConversionMultiplier <= 0 {
		return nil, domain.ErrInvalidMultiplier
	}

	conv := &domain.ProductUOMConversion{
		ID:                   uuid.New().String(),
		ProductID:            productID,
		UOMCode:              strings.ToUpper(strings.TrimSpace(cmd.UOMCode)),
		ConversionMultiplier: cmd.ConversionMultiplier,
		WeightKg:             cmd.WeightKg,
		VolumeM3:             cmd.VolumeM3,
		RetailPrice:          cmd.RetailPrice,
	}

	if err := s.productRepo.AddUOMConversion(ctx, conv); err != nil {
		return nil, err
	}

	return conv, nil
}

// ==========================================
// Material Curah (Bulk Density & Dual Measurement)
// ==========================================

func (s *CatalogService) ListBulkDensityMatrices(ctx context.Context) ([]domain.BulkDensityMatrix, error) {
	return s.bulkDensityRepo.List(ctx)
}

func (s *CatalogService) CalculateBulkMaterialVolume(ctx context.Context, materialCode string, volumeM3 float64) (*domain.VolumeCalculationResult, error) {
	matrix, err := s.bulkDensityRepo.FindByCode(ctx, materialCode)
	if err != nil {
		return nil, domain.ErrBulkMaterialNotFound
	}

	result := matrix.CalculateWeightFromVolume(volumeM3)
	return &result, nil
}

func (s *CatalogService) ValidateWeighbridgeTolerance(ctx context.Context, materialCode string, volumeM3, actualWeightKg float64) (bool, float64, string, error) {
	matrix, err := s.bulkDensityRepo.FindByCode(ctx, materialCode)
	if err != nil {
		return false, 0, "", domain.ErrBulkMaterialNotFound
	}

	isNormal, diffPct, msg := matrix.IsWeightWithinTolerance(volumeM3, actualWeightKg)
	return isNormal, diffPct, msg, nil
}
