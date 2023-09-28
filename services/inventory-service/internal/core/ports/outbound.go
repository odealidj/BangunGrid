package ports

import (
	"context"
	"time"

	"bangungrid/services/inventory-service/internal/core/domain"
)

type WarehouseRepositoryPort interface {
	Create(ctx context.Context, w *domain.Warehouse) error
	FindByID(ctx context.Context, tenantID, id string) (*domain.Warehouse, error)
	FindByCode(ctx context.Context, tenantID, code string) (*domain.Warehouse, error)
	List(ctx context.Context, tenantID string) ([]domain.Warehouse, error)
}

type StockRepositoryPort interface {
	Find(ctx context.Context, tenantID, warehouseID, productID string) (*domain.StockItem, error)
	Upsert(ctx context.Context, item *domain.StockItem) error
	ListByWarehouse(ctx context.Context, tenantID, warehouseID string) ([]domain.StockItem, error)
}

type MovementRepositoryPort interface {
	Create(ctx context.Context, m *domain.StockMovement) error
	List(ctx context.Context, tenantID, warehouseID, productID string) ([]domain.StockMovement, error)
}

type ReservationRepositoryPort interface {
	Create(ctx context.Context, r *domain.StockReservation) error
	FindByID(ctx context.Context, tenantID, id string) (*domain.StockReservation, error)
	Update(ctx context.Context, r *domain.StockReservation) error
}

type DistributedLockPort interface {
	AcquireLock(ctx context.Context, lockKey string, ttl time.Duration) (bool, func(), error)
}

type EventPublisherPort interface {
	PublishStockEvent(ctx context.Context, eventType string, payload any) error
}
