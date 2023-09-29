package services_test

import (
	"context"
	"testing"
	"time"

	"bangungrid/services/inventory-service/internal/core/domain"
	"bangungrid/services/inventory-service/internal/core/ports"
	"bangungrid/services/inventory-service/internal/core/services"
)

// ==========================================
// Mocks
// ==========================================

type mockWarehouseRepo struct {
	warehouses map[string]*domain.Warehouse
}

func newMockWarehouseRepo() *mockWarehouseRepo {
	return &mockWarehouseRepo{warehouses: make(map[string]*domain.Warehouse)}
}

func (m *mockWarehouseRepo) Create(ctx context.Context, w *domain.Warehouse) error {
	m.warehouses[w.ID] = w
	return nil
}

func (m *mockWarehouseRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.Warehouse, error) {
	w, ok := m.warehouses[id]
	if !ok || w.TenantID != tenantID {
		return nil, domain.ErrWarehouseNotFound
	}
	return w, nil
}

func (m *mockWarehouseRepo) FindByCode(ctx context.Context, tenantID, code string) (*domain.Warehouse, error) {
	for _, w := range m.warehouses {
		if w.TenantID == tenantID && w.Code == code {
			return w, nil
		}
	}
	return nil, domain.ErrWarehouseNotFound
}

func (m *mockWarehouseRepo) List(ctx context.Context, tenantID string) ([]domain.Warehouse, error) {
	var list []domain.Warehouse
	for _, w := range m.warehouses {
		if w.TenantID == tenantID {
			list = append(list, *w)
		}
	}
	return list, nil
}

type mockStockRepo struct {
	stocks map[string]*domain.StockItem
}

func newMockStockRepo() *mockStockRepo {
	return &mockStockRepo{stocks: make(map[string]*domain.StockItem)}
}

func (m *mockStockRepo) key(tenantID, warehouseID, productID string) string {
	return tenantID + ":" + warehouseID + ":" + productID
}

func (m *mockStockRepo) Find(ctx context.Context, tenantID, warehouseID, productID string) (*domain.StockItem, error) {
	s, ok := m.stocks[m.key(tenantID, warehouseID, productID)]
	if !ok {
		return nil, domain.ErrStockNotFound
	}
	item := *s
	return &item, nil
}

func (m *mockStockRepo) Upsert(ctx context.Context, item *domain.StockItem) error {
	k := m.key(item.TenantID, item.WarehouseID, item.ProductID)
	copied := *item
	m.stocks[k] = &copied
	return nil
}

func (m *mockStockRepo) ListByWarehouse(ctx context.Context, tenantID, warehouseID string) ([]domain.StockItem, error) {
	var list []domain.StockItem
	for _, s := range m.stocks {
		if s.TenantID == tenantID && s.WarehouseID == warehouseID {
			list = append(list, *s)
		}
	}
	return list, nil
}

type mockMovementRepo struct {
	movements []domain.StockMovement
}

func (m *mockMovementRepo) Create(ctx context.Context, mv *domain.StockMovement) error {
	m.movements = append(m.movements, *mv)
	return nil
}

func (m *mockMovementRepo) List(ctx context.Context, tenantID, warehouseID, productID string) ([]domain.StockMovement, error) {
	var list []domain.StockMovement
	for _, mv := range m.movements {
		if mv.TenantID == tenantID && mv.WarehouseID == warehouseID && mv.ProductID == productID {
			list = append(list, mv)
		}
	}
	return list, nil
}

type mockReservationRepo struct {
	reservations map[string]*domain.StockReservation
}

func newMockReservationRepo() *mockReservationRepo {
	return &mockReservationRepo{reservations: make(map[string]*domain.StockReservation)}
}

func (m *mockReservationRepo) Create(ctx context.Context, r *domain.StockReservation) error {
	copied := *r
	m.reservations[r.ID] = &copied
	return nil
}

func (m *mockReservationRepo) FindByID(ctx context.Context, tenantID, id string) (*domain.StockReservation, error) {
	r, ok := m.reservations[id]
	if !ok || r.TenantID != tenantID {
		return nil, domain.ErrReservationNotFound
	}
	copied := *r
	return &copied, nil
}

func (m *mockReservationRepo) Update(ctx context.Context, r *domain.StockReservation) error {
	copied := *r
	m.reservations[r.ID] = &copied
	return nil
}

type mockDistLock struct{}

func (m *mockDistLock) AcquireLock(ctx context.Context, lockKey string, ttl time.Duration) (bool, func(), error) {
	return true, func() {}, nil
}

type mockEventPub struct {
	events []string
}

func (m *mockEventPub) PublishStockEvent(ctx context.Context, eventType string, payload any) error {
	m.events = append(m.events, eventType)
	return nil
}

// ==========================================
// Tests
// ==========================================

func setupInventoryService() (*services.InventoryService, *mockWarehouseRepo, *mockStockRepo, *mockReservationRepo, *mockMovementRepo) {
	wRepo := newMockWarehouseRepo()
	sRepo := newMockStockRepo()
	mRepo := &mockMovementRepo{}
	rRepo := newMockReservationRepo()
	lock := &mockDistLock{}
	pub := &mockEventPub{}

	svc := services.NewInventoryService(wRepo, sRepo, mRepo, rRepo, lock, pub).(*services.InventoryService)
	return svc, wRepo, sRepo, rRepo, mRepo
}

func TestInventoryService_CreateWarehouse_Success(t *testing.T) {
	svc, _, _, _, _ := setupInventoryService()
	ctx := context.Background()

	cmd := ports.CreateWarehouseCommand{
		TenantID:   "tenant-001",
		Code:       "WH-CKG-01",
		Name:       "Gudang Utama Semen Cakung",
		Type:       domain.WarehouseTypeCentral,
		Address:    "Kawasan Industri Cakung Blok B",
		City:       "Jakarta Timur",
		PostalCode: "13910",
		Latitude:   -6.1954,
		Longitude:  106.9452,
	}

	wh, err := svc.CreateWarehouse(ctx, cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if wh.ID == "" {
		t.Errorf("expected non-empty ID")
	}
	if wh.Code != "WH-CKG-01" {
		t.Errorf("expected code WH-CKG-01, got %s", wh.Code)
	}
}

func TestInventoryService_AddStock_And_ATP(t *testing.T) {
	svc, _, _, _, mRepo := setupInventoryService()
	ctx := context.Background()

	cmd := ports.AddStockCommand{
		TenantID:    "tenant-001",
		WarehouseID: "wh-001",
		ProductID:   "prod-semen-01",
		UOMCode:     "SAK",
		Quantity:    500.0,
		ReferenceID: "PO-INBOUND-202309-001",
		Notes:       "Penerimaan 500 sak Semen Gresik dari pabrik Tuban",
	}

	stock, err := svc.AddStock(ctx, cmd)
	if err != nil {
		t.Fatalf("expected add stock success, got: %v", err)
	}

	if stock.PhysicalQty != 500.0 {
		t.Errorf("expected physical 500, got %.2f", stock.PhysicalQty)
	}
	if stock.AvailableQty != 500.0 {
		t.Errorf("expected available 500, got %.2f", stock.AvailableQty)
	}
	if len(mRepo.movements) != 1 {
		t.Errorf("expected 1 movement logged, got %d", len(mRepo.movements))
	}
}

func TestInventoryService_Reserve_Commit_Lifecycle(t *testing.T) {
	svc, _, _, _, _ := setupInventoryService()
	ctx := context.Background()

	// 1. Inisialisasi stok 200 Sak
	_, err := svc.AddStock(ctx, ports.AddStockCommand{
		TenantID:    "tenant-001",
		WarehouseID: "wh-001",
		ProductID:   "prod-semen-01",
		UOMCode:     "SAK",
		Quantity:    200.0,
	})
	if err != nil {
		t.Fatalf("failed setup stock: %v", err)
	}

	// 2. Kontraktor memesan 80 Sak (Soft-lock reservation)
	resCmd := ports.ReserveStockCommand{
		TenantID:    "tenant-001",
		WarehouseID: "wh-001",
		ProductID:   "prod-semen-01",
		OrderID:     "SO-B2B-1001",
		Quantity:    80.0,
		UOMCode:     "SAK",
		TTL:         15 * time.Minute,
	}

	reservation, err := svc.ReserveStock(ctx, resCmd)
	if err != nil {
		t.Fatalf("expected reservation success, got: %v", err)
	}

	// Cek status stok setelah reservasi:
	// Physical = 200, Reserved = 80, Available = 120
	currentStock, _ := svc.GetStock(ctx, "tenant-001", "wh-001", "prod-semen-01")
	if currentStock.PhysicalQty != 200 || currentStock.ReservedQty != 80 || currentStock.AvailableQty != 120 {
		t.Errorf("invalid stock state post-reservation: P=%.0f, R=%.0f, A=%.0f",
			currentStock.PhysicalQty, currentStock.ReservedQty, currentStock.AvailableQty)
	}

	// 3. Coba reservasi melebihi Available (misal minta 150 sak padahal sisa 120 sak)
	_, errOver := svc.ReserveStock(ctx, ports.ReserveStockCommand{
		TenantID:    "tenant-001",
		WarehouseID: "wh-001",
		ProductID:   "prod-semen-01",
		OrderID:     "SO-B2B-1002",
		Quantity:    150.0,
		UOMCode:     "SAK",
	})
	if errOver != domain.ErrInsufficientStock {
		t.Errorf("expected ErrInsufficientStock, got: %v", errOver)
	}

	// 4. Pembayaran DP / Approval TOP B2B sukses -> Commit Reservation
	errCommit := svc.CommitReservation(ctx, "tenant-001", reservation.ID)
	if errCommit != nil {
		t.Fatalf("expected commit reservation success, got: %v", errCommit)
	}

	// Cek status stok pasca commit:
	// Physical = 120, Reserved = 0, Available = 120
	postCommitStock, _ := svc.GetStock(ctx, "tenant-001", "wh-001", "prod-semen-01")
	if postCommitStock.PhysicalQty != 120 || postCommitStock.ReservedQty != 0 || postCommitStock.AvailableQty != 120 {
		t.Errorf("invalid stock state post-commit: P=%.0f, R=%.0f, A=%.0f",
			postCommitStock.PhysicalQty, postCommitStock.ReservedQty, postCommitStock.AvailableQty)
	}
}

func TestInventoryService_ReleaseReservation_OnOrderCancelled(t *testing.T) {
	svc, _, _, _, _ := setupInventoryService()
	ctx := context.Background()

	// Inisialisasi stok 100 Sak
	_, _ = svc.AddStock(ctx, ports.AddStockCommand{
		TenantID:    "tenant-001",
		WarehouseID: "wh-001",
		ProductID:   "prod-semen-01",
		UOMCode:     "SAK",
		Quantity:    100.0,
	})

	// Reservasi 40 Sak
	res, _ := svc.ReserveStock(ctx, ports.ReserveStockCommand{
		TenantID:    "tenant-001",
		WarehouseID: "wh-001",
		ProductID:   "prod-semen-01",
		OrderID:     "SO-EXPIRED-99",
		Quantity:    40.0,
		UOMCode:     "SAK",
	})

	// Order dibatalkan atau checkout timeout -> Release
	err := svc.ReleaseReservation(ctx, "tenant-001", res.ID)
	if err != nil {
		t.Fatalf("expected release success: %v", err)
	}

	// Stok harus kembali utuh: Physical = 100, Reserved = 0, Available = 100
	stock, _ := svc.GetStock(ctx, "tenant-001", "wh-001", "prod-semen-01")
	if stock.PhysicalQty != 100 || stock.ReservedQty != 0 || stock.AvailableQty != 100 {
		t.Errorf("failed reverting reservation on release: P=%.0f, R=%.0f, A=%.0f",
			stock.PhysicalQty, stock.ReservedQty, stock.AvailableQty)
	}
}

func TestInventoryService_AdjustShrinkage(t *testing.T) {
	svc, _, _, _, mRepo := setupInventoryService()
	ctx := context.Background()

	// 1. Inisialisasi material curah: Pasir Pasang 14.000 Kg (10 m3) di Stockyard
	_, _ = svc.AddStock(ctx, ports.AddStockCommand{
		TenantID:    "tenant-001",
		WarehouseID: "stockyard-marunda",
		ProductID:   "bulk-sand-01",
		UOMCode:     "KG",
		Quantity:    14000.0,
	})

	// 2. Opname mingguan: terjadi penyusutan air/kelembaban 250 Kg
	adjCmd := ports.AdjustShrinkageCommand{
		TenantID:     "tenant-001",
		WarehouseID:  "stockyard-marunda",
		ProductID:    "bulk-sand-01",
		UOMCode:      "KG",
		ShrinkageQty: 250.0,
		Reason:       "Penyusutan kadar air alami pasir cuaca terik",
		ReferenceID:  "OPNAME-202309-W4",
	}

	stock, err := svc.AdjustShrinkage(ctx, adjCmd)
	if err != nil {
		t.Fatalf("expected adjustment success, got: %v", err)
	}

	if stock.PhysicalQty != 13750.0 {
		t.Errorf("expected physical 13750, got %.2f", stock.PhysicalQty)
	}

	// Verifikasi audit movement tercatat dengan kuantitas -250
	lastMovement := mRepo.movements[len(mRepo.movements)-1]
	if lastMovement.Quantity != -250.0 || lastMovement.MovementType != domain.MovementShrinkageAdjustment {
		t.Errorf("expected movement -250 shrinkage, got %.2f type %s", lastMovement.Quantity, lastMovement.MovementType)
	}
}
