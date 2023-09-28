package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"bangungrid/services/inventory-service/internal/core/domain"
	"bangungrid/services/inventory-service/internal/core/ports"
)

type InventoryService struct {
	warehouseRepo   ports.WarehouseRepositoryPort
	stockRepo       ports.StockRepositoryPort
	movementRepo    ports.MovementRepositoryPort
	reservationRepo ports.ReservationRepositoryPort
	distLock        ports.DistributedLockPort
	eventPublisher  ports.EventPublisherPort
}

func NewInventoryService(
	wRepo ports.WarehouseRepositoryPort,
	sRepo ports.StockRepositoryPort,
	mRepo ports.MovementRepositoryPort,
	rRepo ports.ReservationRepositoryPort,
	lock ports.DistributedLockPort,
	pub ports.EventPublisherPort,
) ports.InventoryUseCasePort {
	return &InventoryService{
		warehouseRepo:   wRepo,
		stockRepo:       sRepo,
		movementRepo:    mRepo,
		reservationRepo: rRepo,
		distLock:        lock,
		eventPublisher:  pub,
	}
}

// ==========================================
// Warehouse Operations
// ==========================================

func (s *InventoryService) CreateWarehouse(ctx context.Context, cmd ports.CreateWarehouseCommand) (*domain.Warehouse, error) {
	now := time.Now().UTC()
	warehouse := &domain.Warehouse{
		ID:         uuid.New().String(),
		TenantID:   cmd.TenantID,
		Code:       strings.ToUpper(strings.TrimSpace(cmd.Code)),
		Name:       strings.TrimSpace(cmd.Name),
		Type:       cmd.Type,
		Address:    cmd.Address,
		City:       cmd.City,
		PostalCode: cmd.PostalCode,
		Latitude:   cmd.Latitude,
		Longitude:  cmd.Longitude,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := warehouse.Validate(); err != nil {
		return nil, err
	}

	if err := s.warehouseRepo.Create(ctx, warehouse); err != nil {
		return nil, err
	}

	return warehouse, nil
}

func (s *InventoryService) GetWarehouseByID(ctx context.Context, tenantID, id string) (*domain.Warehouse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantRequired
	}
	return s.warehouseRepo.FindByID(ctx, tenantID, id)
}

func (s *InventoryService) ListWarehouses(ctx context.Context, tenantID string) ([]domain.Warehouse, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantRequired
	}
	return s.warehouseRepo.List(ctx, tenantID)
}

// ==========================================
// Stock & ATP Operations
// ==========================================

func (s *InventoryService) GetStock(ctx context.Context, tenantID, warehouseID, productID string) (*domain.StockItem, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantRequired
	}
	item, err := s.stockRepo.Find(ctx, tenantID, warehouseID, productID)
	if err != nil {
		return nil, err
	}
	item.Recalculate()
	return item, nil
}

func (s *InventoryService) ListStockByWarehouse(ctx context.Context, tenantID, warehouseID string) ([]domain.StockItem, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantRequired
	}
	return s.stockRepo.ListByWarehouse(ctx, tenantID, warehouseID)
}

func (s *InventoryService) AddStock(ctx context.Context, cmd ports.AddStockCommand) (*domain.StockItem, error) {
	if cmd.Quantity <= 0 {
		return nil, domain.ErrInvalidQuantity
	}
	if cmd.TenantID == "" {
		return nil, domain.ErrTenantRequired
	}

	lockKey := fmt.Sprintf("lock:stock:%s:%s:%s", cmd.TenantID, cmd.WarehouseID, cmd.ProductID)
	locked, unlock, err := s.distLock.AcquireLock(ctx, lockKey, 5*time.Second)
	if err != nil || !locked {
		return nil, domain.ErrDistributedLockFailed
	}
	defer unlock()

	stockItem, err := s.stockRepo.Find(ctx, cmd.TenantID, cmd.WarehouseID, cmd.ProductID)
	if err != nil {
		// Buat item stok baru jika belum ada di gudang
		now := time.Now().UTC()
		stockItem = &domain.StockItem{
			ID:          uuid.New().String(),
			TenantID:    cmd.TenantID,
			WarehouseID: cmd.WarehouseID,
			ProductID:   cmd.ProductID,
			UOMCode:     strings.ToUpper(cmd.UOMCode),
			PhysicalQty: 0,
			ReservedQty: 0,
			UpdatedAt:   now,
		}
	}

	if err := stockItem.AddPhysical(cmd.Quantity); err != nil {
		return nil, err
	}

	if err := s.stockRepo.Upsert(ctx, stockItem); err != nil {
		return nil, fmt.Errorf("gagal memperbarui stok: %w", err)
	}

	// Catat pergerakan stok (Movement Ledger)
	movement := &domain.StockMovement{
		ID:           uuid.New().String(),
		TenantID:     cmd.TenantID,
		WarehouseID:  cmd.WarehouseID,
		ProductID:    cmd.ProductID,
		UOMCode:      stockItem.UOMCode,
		Quantity:     cmd.Quantity,
		MovementType: domain.MovementInboundPO,
		ReferenceID:  cmd.ReferenceID,
		Notes:        cmd.Notes,
		CreatedAt:    time.Now().UTC(),
	}
	_ = s.movementRepo.Create(ctx, movement)

	if s.eventPublisher != nil {
		_ = s.eventPublisher.PublishStockEvent(ctx, "inventory.stock.added", stockItem)
	}

	return stockItem, nil
}

func (s *InventoryService) AdjustShrinkage(ctx context.Context, cmd ports.AdjustShrinkageCommand) (*domain.StockItem, error) {
	if cmd.ShrinkageQty <= 0 {
		return nil, domain.ErrInvalidQuantity
	}
	if cmd.TenantID == "" {
		return nil, domain.ErrTenantRequired
	}

	lockKey := fmt.Sprintf("lock:stock:%s:%s:%s", cmd.TenantID, cmd.WarehouseID, cmd.ProductID)
	locked, unlock, err := s.distLock.AcquireLock(ctx, lockKey, 5*time.Second)
	if err != nil || !locked {
		return nil, domain.ErrDistributedLockFailed
	}
	defer unlock()

	stockItem, err := s.stockRepo.Find(ctx, cmd.TenantID, cmd.WarehouseID, cmd.ProductID)
	if err != nil {
		return nil, domain.ErrStockNotFound
	}

	if stockItem.PhysicalQty < cmd.ShrinkageQty {
		return nil, fmt.Errorf("kuantitas penyusutan (%.2f) melebihi stok fisik yang ada (%.2f)", cmd.ShrinkageQty, stockItem.PhysicalQty)
	}

	stockItem.PhysicalQty -= cmd.ShrinkageQty
	stockItem.Recalculate()
	stockItem.UpdatedAt = time.Now().UTC()

	if err := s.stockRepo.Upsert(ctx, stockItem); err != nil {
		return nil, fmt.Errorf("gagal memperbarui penyesuaian susut: %w", err)
	}

	// Catat pergerakan penyusutan dengan nilai kuantitas negatif
	movement := &domain.StockMovement{
		ID:           uuid.New().String(),
		TenantID:     cmd.TenantID,
		WarehouseID:  cmd.WarehouseID,
		ProductID:    cmd.ProductID,
		UOMCode:      stockItem.UOMCode,
		Quantity:     -cmd.ShrinkageQty,
		MovementType: domain.MovementShrinkageAdjustment,
		ReferenceID:  cmd.ReferenceID,
		Notes:        cmd.Reason,
		CreatedAt:    time.Now().UTC(),
	}
	_ = s.movementRepo.Create(ctx, movement)

	if s.eventPublisher != nil {
		_ = s.eventPublisher.PublishStockEvent(ctx, "inventory.stock.shrinkage_adjusted", stockItem)
	}

	return stockItem, nil
}

// ==========================================
// Stock Reservation & Locking
// ==========================================

func (s *InventoryService) ReserveStock(ctx context.Context, cmd ports.ReserveStockCommand) (*domain.StockReservation, error) {
	if cmd.Quantity <= 0 {
		return nil, domain.ErrInvalidQuantity
	}
	if cmd.TenantID == "" {
		return nil, domain.ErrTenantRequired
	}
	if cmd.TTL <= 0 {
		cmd.TTL = 15 * time.Minute
	}

	lockKey := fmt.Sprintf("lock:stock:%s:%s:%s", cmd.TenantID, cmd.WarehouseID, cmd.ProductID)
	locked, unlock, err := s.distLock.AcquireLock(ctx, lockKey, 5*time.Second)
	if err != nil || !locked {
		return nil, domain.ErrDistributedLockFailed
	}
	defer unlock()

	stockItem, err := s.stockRepo.Find(ctx, cmd.TenantID, cmd.WarehouseID, cmd.ProductID)
	if err != nil {
		return nil, domain.ErrStockNotFound
	}

	if !stockItem.CanReserve(cmd.Quantity) {
		return nil, domain.ErrInsufficientStock
	}

	if err := stockItem.Reserve(cmd.Quantity); err != nil {
		return nil, err
	}

	if err := s.stockRepo.Upsert(ctx, stockItem); err != nil {
		return nil, fmt.Errorf("gagal memperbarui reservasi stok: %w", err)
	}

	now := time.Now().UTC()
	reservation := &domain.StockReservation{
		ID:          uuid.New().String(),
		TenantID:    cmd.TenantID,
		WarehouseID: cmd.WarehouseID,
		ProductID:   cmd.ProductID,
		OrderID:     cmd.OrderID,
		Quantity:    cmd.Quantity,
		UOMCode:     cmd.UOMCode,
		Status:      domain.ReservationStatusActive,
		ExpiresAt:   now.Add(cmd.TTL),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.reservationRepo.Create(ctx, reservation); err != nil {
		return nil, fmt.Errorf("gagal mencatat reservasi: %w", err)
	}

	if s.eventPublisher != nil {
		_ = s.eventPublisher.PublishStockEvent(ctx, "inventory.stock.reserved", reservation)
	}

	return reservation, nil
}

func (s *InventoryService) CommitReservation(ctx context.Context, tenantID, reservationID string) error {
	reservation, err := s.reservationRepo.FindByID(ctx, tenantID, reservationID)
	if err != nil {
		return domain.ErrReservationNotFound
	}

	if reservation.Status != domain.ReservationStatusActive {
		return domain.ErrReservationAlreadyProcessed
	}

	lockKey := fmt.Sprintf("lock:stock:%s:%s:%s", tenantID, reservation.WarehouseID, reservation.ProductID)
	locked, unlock, err := s.distLock.AcquireLock(ctx, lockKey, 5*time.Second)
	if err != nil || !locked {
		return domain.ErrDistributedLockFailed
	}
	defer unlock()

	stockItem, err := s.stockRepo.Find(ctx, tenantID, reservation.WarehouseID, reservation.ProductID)
	if err != nil {
		return domain.ErrStockNotFound
	}

	if err := stockItem.DeductCommitted(reservation.Quantity); err != nil {
		return err
	}

	if err := s.stockRepo.Upsert(ctx, stockItem); err != nil {
		return fmt.Errorf("gagal memotong stok fisik: %w", err)
	}

	reservation.Status = domain.ReservationStatusCommitted
	reservation.UpdatedAt = time.Now().UTC()
	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		return fmt.Errorf("gagal memperbarui status reservasi: %w", err)
	}

	// Catat pergerakan stok keluar (Outbound SO)
	movement := &domain.StockMovement{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		WarehouseID:  reservation.WarehouseID,
		ProductID:    reservation.ProductID,
		UOMCode:      reservation.UOMCode,
		Quantity:     -reservation.Quantity,
		MovementType: domain.MovementOutboundSO,
		ReferenceID:  reservation.OrderID,
		Notes:        "Pengiriman pesanan (Sales Order committed)",
		CreatedAt:    time.Now().UTC(),
	}
	_ = s.movementRepo.Create(ctx, movement)

	if s.eventPublisher != nil {
		_ = s.eventPublisher.PublishStockEvent(ctx, "inventory.stock.committed", reservation)
	}

	return nil
}

func (s *InventoryService) ReleaseReservation(ctx context.Context, tenantID, reservationID string) error {
	reservation, err := s.reservationRepo.FindByID(ctx, tenantID, reservationID)
	if err != nil {
		return domain.ErrReservationNotFound
	}

	if reservation.Status != domain.ReservationStatusActive {
		return domain.ErrReservationAlreadyProcessed
	}

	lockKey := fmt.Sprintf("lock:stock:%s:%s:%s", tenantID, reservation.WarehouseID, reservation.ProductID)
	locked, unlock, err := s.distLock.AcquireLock(ctx, lockKey, 5*time.Second)
	if err != nil || !locked {
		return domain.ErrDistributedLockFailed
	}
	defer unlock()

	stockItem, err := s.stockRepo.Find(ctx, tenantID, reservation.WarehouseID, reservation.ProductID)
	if err != nil {
		return domain.ErrStockNotFound
	}

	if err := stockItem.ReleaseReservation(reservation.Quantity); err != nil {
		return err
	}

	if err := s.stockRepo.Upsert(ctx, stockItem); err != nil {
		return fmt.Errorf("gagal merilis kuantitas reservasi: %w", err)
	}

	reservation.Status = domain.ReservationStatusCancelled
	reservation.UpdatedAt = time.Now().UTC()
	if err := s.reservationRepo.Update(ctx, reservation); err != nil {
		return fmt.Errorf("gagal membatalkan reservasi: %w", err)
	}

	if s.eventPublisher != nil {
		_ = s.eventPublisher.PublishStockEvent(ctx, "inventory.stock.released", reservation)
	}

	return nil
}

// ==========================================
// Audit Ledger Movements
// ==========================================

func (s *InventoryService) ListMovements(ctx context.Context, tenantID, warehouseID, productID string) ([]domain.StockMovement, error) {
	if tenantID == "" {
		return nil, domain.ErrTenantRequired
	}
	return s.movementRepo.List(ctx, tenantID, warehouseID, productID)
}
