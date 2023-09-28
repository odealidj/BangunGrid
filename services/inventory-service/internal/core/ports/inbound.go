package ports

import (
	"context"
	"time"

	"bangungrid/services/inventory-service/internal/core/domain"
)

type CreateWarehouseCommand struct {
	TenantID   string
	Code       string
	Name       string
	Type       domain.WarehouseType
	Address    string
	City       string
	PostalCode string
	Latitude   float64
	Longitude  float64
}

type AddStockCommand struct {
	TenantID    string
	WarehouseID string
	ProductID   string
	UOMCode     string
	Quantity    float64
	ReferenceID string // PO / Inbound Delivery Order
	Notes       string
}

type AdjustShrinkageCommand struct {
	TenantID    string
	WarehouseID string
	ProductID   string
	UOMCode     string
	ShrinkageQty float64 // Berapa unit/kg yang susut (nilai positif)
	Reason      string  // e.g. "Penyusutan kelembaban pasir pasca cuaca panas", "Opname berkala"
	ReferenceID string  // Audit Opname ID
}

type ReserveStockCommand struct {
	TenantID    string
	WarehouseID string
	ProductID   string
	OrderID     string
	Quantity    float64
	UOMCode     string
	TTL         time.Duration // e.g. 15 minutes default
}

type InventoryUseCasePort interface {
	// Warehouse Operations
	CreateWarehouse(ctx context.Context, cmd CreateWarehouseCommand) (*domain.Warehouse, error)
	GetWarehouseByID(ctx context.Context, tenantID, id string) (*domain.Warehouse, error)
	ListWarehouses(ctx context.Context, tenantID string) ([]domain.Warehouse, error)

	// Stock & ATP (Available-to-Promise)
	GetStock(ctx context.Context, tenantID, warehouseID, productID string) (*domain.StockItem, error)
	ListStockByWarehouse(ctx context.Context, tenantID, warehouseID string) ([]domain.StockItem, error)
	AddStock(ctx context.Context, cmd AddStockCommand) (*domain.StockItem, error)
	AdjustShrinkage(ctx context.Context, cmd AdjustShrinkageCommand) (*domain.StockItem, error)

	// Distributed Reservation
	ReserveStock(ctx context.Context, cmd ReserveStockCommand) (*domain.StockReservation, error)
	CommitReservation(ctx context.Context, tenantID, reservationID string) error
	ReleaseReservation(ctx context.Context, tenantID, reservationID string) error

	// Audit Trail / Movements
	ListMovements(ctx context.Context, tenantID, warehouseID, productID string) ([]domain.StockMovement, error)
}
