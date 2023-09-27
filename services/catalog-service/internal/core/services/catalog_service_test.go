package services_test

import (
	"context"
	"testing"

	"bangungrid/services/catalog-service/internal/core/domain"
	"bangungrid/services/catalog-service/internal/core/ports"
	"bangungrid/services/catalog-service/internal/core/services"
)

// ==========================================
// Mocks for Ports
// ==========================================

type mockCategoryRepo struct {
	categories map[string]*domain.Category
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{categories: make(map[string]*domain.Category)}
}

func (m *mockCategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	m.categories[c.ID] = c
	return nil
}

func (m *mockCategoryRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.Category, error) {
	c, ok := m.categories[id]
	if !ok {
		return nil, domain.ErrCategoryNotFound
	}
	return c, nil
}

func (m *mockCategoryRepo) List(ctx context.Context, tenantID string) ([]domain.Category, error) {
	var list []domain.Category
	for _, c := range m.categories {
		if c.TenantID == tenantID {
			list = append(list, *c)
		}
	}
	return list, nil
}

type mockProductRepo struct {
	products    map[string]*domain.Product
	conversions map[string][]domain.ProductUOMConversion
	tiers       map[string][]domain.ProductPriceTier
}

func newMockProductRepo() *mockProductRepo {
	return &mockProductRepo{
		products:    make(map[string]*domain.Product),
		conversions: make(map[string][]domain.ProductUOMConversion),
		tiers:       make(map[string][]domain.ProductPriceTier),
	}
}

func (m *mockProductRepo) Create(ctx context.Context, p *domain.Product) error {
	m.products[p.ID] = p
	return nil
}

func (m *mockProductRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	res := *p
	res.UOMConversions = m.conversions[id]
	res.PriceTiers = m.tiers[id]
	return &res, nil
}

func (m *mockProductRepo) List(ctx context.Context, q ports.ListProductsQuery) ([]domain.Product, int64, error) {
	var list []domain.Product
	for _, p := range m.products {
		if p.TenantID == q.TenantID {
			res := *p
			res.UOMConversions = m.conversions[p.ID]
			res.PriceTiers = m.tiers[p.ID]
			list = append(list, res)
		}
	}
	return list, int64(len(list)), nil
}

func (m *mockProductRepo) AddUOMConversion(ctx context.Context, c *domain.ProductUOMConversion) error {
	m.conversions[c.ProductID] = append(m.conversions[c.ProductID], *c)
	return nil
}

func (m *mockProductRepo) AddPriceTier(ctx context.Context, t *domain.ProductPriceTier) error {
	m.tiers[t.ProductID] = append(m.tiers[t.ProductID], *t)
	return nil
}

type mockBulkDensityRepo struct {
	matrices map[string]*domain.BulkDensityMatrix
}

func newMockBulkDensityRepo() *mockBulkDensityRepo {
	repo := &mockBulkDensityRepo{matrices: make(map[string]*domain.BulkDensityMatrix)}
	// Default matrix
	repo.matrices["BULK-SAND-01"] = &domain.BulkDensityMatrix{
		ID:                     "m-001",
		MaterialCode:           "BULK-SAND-01",
		MaterialName:           "Pasir Pasang Halus",
		StandardDensityKgPerM3: 1400.0,
		NormalTolerancePct:     2.0,
	}
	return repo
}

func (m *mockBulkDensityRepo) FindByCode(ctx context.Context, code string) (*domain.BulkDensityMatrix, error) {
	mat, ok := m.matrices[code]
	if !ok {
		return nil, domain.ErrBulkMaterialNotFound
	}
	return mat, nil
}

func (m *mockBulkDensityRepo) List(ctx context.Context) ([]domain.BulkDensityMatrix, error) {
	var list []domain.BulkDensityMatrix
	for _, mat := range m.matrices {
		list = append(list, *mat)
	}
	return list, nil
}

func (m *mockBulkDensityRepo) SeedDefaults(ctx context.Context) error {
	return nil
}

// ==========================================
// Test Cases
// ==========================================

func setupCatalogService() (*services.CatalogService, *mockCategoryRepo, *mockProductRepo, *mockBulkDensityRepo) {
	catRepo := newMockCategoryRepo()
	prodRepo := newMockProductRepo()
	bulkRepo := newMockBulkDensityRepo()
	svc := services.NewCatalogService(catRepo, prodRepo, bulkRepo).(*services.CatalogService)
	return svc, catRepo, prodRepo, bulkRepo
}

func TestCatalogService_CreateCategory_Success(t *testing.T) {
	svc, _, _, _ := setupCatalogService()
	ctx := context.Background()

	cmd := ports.CreateCategoryCommand{
		TenantID: "tenant-001",
		Name:     "Semen & Mortar",
	}

	cat, err := svc.CreateCategory(ctx, cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cat.ID == "" {
		t.Errorf("expected non-empty category ID")
	}
	if cat.Slug != "semen-&-mortar" {
		t.Errorf("expected category slug semen-&-mortar, got %s", cat.Slug)
	}
}

func TestCatalogService_CreateProduct_WithMultiUoM(t *testing.T) {
	svc, _, prodRepo, _ := setupCatalogService()
	ctx := context.Background()

	cmd := ports.CreateProductCommand{
		TenantID:       "tenant-001",
		CategoryID:     "cat-001",
		BrandName:      "Semen Gresik",
		Name:           "Semen Gresik PCC 40kg",
		Description:    "Semen Portland Composite untuk plesteran dan acian",
		IsBulkMaterial: false,
		BaseUoM:        "SAK",
		Specifications: map[string]any{
			"tipe": "PCC",
			"berat_kemasan_kg": 40,
			"sni": "SNI 7064:2014",
		},
		Conversions: []ports.CreateUOMConversionCommand{
			{
				UOMCode:              "SAK",
				ConversionMultiplier: 1.0,
				WeightKg:             40.0,
				VolumeM3:             0.03,
				RetailPrice:          65000,
			},
			{
				UOMCode:              "PALET",
				ConversionMultiplier: 40.0, // 1 Palet = 40 Sak
				WeightKg:             1600.0,
				VolumeM3:             1.2,
				RetailPrice:          2520000, // Harga grosir per palet
			},
		},
		PriceTiers: []ports.CreatePriceTierCommand{
			{
				UOMCode:     "SAK",
				TierName:    "CONTRACTOR_TIER_1",
				MinQuantity: 100,
				UnitPrice:   61000, // Diskon kontraktor min 100 sak
			},
		},
	}

	prod, err := svc.CreateProduct(ctx, cmd)
	if err != nil {
		t.Fatalf("expected no error creating product, got: %v", err)
	}

	if prod.ID == "" {
		t.Errorf("expected non-empty product ID")
	}
	if len(prod.UOMConversions) != 2 {
		t.Errorf("expected 2 UOM conversions, got %d", len(prod.UOMConversions))
	}
	if len(prod.PriceTiers) != 1 {
		t.Errorf("expected 1 price tier, got %d", len(prod.PriceTiers))
	}

	// Verify in repository
	saved, err := prodRepo.FindByID(ctx, "tenant-001", prod.ID)
	if err != nil || saved == nil {
		t.Fatalf("expected product to be saved in repository")
	}
}

func TestCatalogService_BulkMaterial_Calculation(t *testing.T) {
	svc, _, _, _ := setupCatalogService()
	ctx := context.Background()

	// 1 Dump Truck Pasir Pasang = 6.8 m3
	// Standard density = 1400 kg/m3 (1.4 Ton/m3)
	// Expected weight = 6.8 * 1400 = 9520 Kg (9.52 Ton)
	// Tolerance 2% = 190.4 Kg
	res, err := svc.CalculateBulkMaterialVolume(ctx, "BULK-SAND-01", 6.8)
	if err != nil {
		t.Fatalf("expected calculation success, got: %v", err)
	}

	if res.ExpectedWeightKg != 9520.0 {
		t.Errorf("expected 9520 kg, got %.2f", res.ExpectedWeightKg)
	}
	if res.ExpectedWeightTon != 9.52 {
		t.Errorf("expected 9.52 ton, got %.2f", res.ExpectedWeightTon)
	}
	if res.MinToleranceKg > res.ExpectedWeightKg || res.MaxToleranceKg < res.ExpectedWeightKg {
		t.Errorf("tolerance range invalid: [%.2f, %.2f]", res.MinToleranceKg, res.MaxToleranceKg)
	}
}

func TestCatalogService_Weighbridge_ToleranceCheck(t *testing.T) {
	svc, _, _, _ := setupCatalogService()
	ctx := context.Background()

	volumeM3 := 6.8
	// Expected = 9520 Kg

	// Kasus 1: Bobot normal wajar (9450 Kg, selisih -0.74%, masih dalam batas 2.0%)
	isNormal, diffPct, _, err := svc.ValidateWeighbridgeTolerance(ctx, "BULK-SAND-01", volumeM3, 9450.0)
	if err != nil {
		t.Fatalf("expected validation success: %v", err)
	}
	if !isNormal {
		t.Errorf("expected 9450 kg to be normal, got diffPct: %.2f%%", diffPct)
	}

	// Kasus 2: Anomali susut berlebih (9100 Kg, selisih -4.41%, melebihi 2.0%)
	isNormal, diffPct, msg, err := svc.ValidateWeighbridgeTolerance(ctx, "BULK-SAND-01", volumeM3, 9100.0)
	if err != nil {
		t.Fatalf("expected validation success: %v", err)
	}
	if isNormal {
		t.Errorf("expected 9100 kg to trigger anomaly, got normal: %s", msg)
	}
	if diffPct >= -2.0 {
		t.Errorf("expected diffPct < -2.0, got %.2f", diffPct)
	}
}
