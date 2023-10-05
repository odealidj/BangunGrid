package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"bangungrid/services/order-service/internal/core/domain"
)

type CreateOrderItemCommand struct {
	SKUID               uuid.UUID `json:"sku_id"`
	ProductNameSnapshot string    `json:"product_name"`
	UOMCodeSnapshot     string    `json:"uom_code"`
	Quantity            float64   `json:"quantity"`
	UnitPriceSnapshot   float64   `json:"unit_price"`
	WeightKg            float64   `json:"weight_kg"`
	VolumeM3            float64   `json:"volume_m3"`
}

type CreateOrderCommand struct {
	TenantID          uuid.UUID                `json:"tenant_id"`
	BranchID          uuid.UUID                `json:"branch_id"`
	CustomerID        uuid.UUID                `json:"customer_id"`
	CustomerType      domain.OrderType         `json:"customer_type"`
	DeliveryAddress   string                   `json:"delivery_address"`
	DeliveryLatitude  float64                  `json:"delivery_latitude"`
	DeliveryLongitude float64                  `json:"delivery_longitude"`
	IncludeUnloading  bool                     `json:"include_unloading"`
	NarrowRoadAccess  bool                     `json:"narrow_road_access"`
	ShippingFee       float64                  `json:"shipping_fee"`
	Notes             string                   `json:"notes"`
	Items             []CreateOrderItemCommand `json:"items"`
}

type EstimateOrderCommand struct {
	Items            []CreateOrderItemCommand `json:"items"`
	IncludeUnloading bool                     `json:"include_unloading"`
	NarrowRoadAccess bool                     `json:"narrow_road_access"`
}

type OrderEstimate struct {
	TotalWeightKg          float64 `json:"total_weight_kg"`
	TotalVolumeM3          float64 `json:"total_volume_m3"`
	EstimatedItemsSubtotal float64 `json:"estimated_items_subtotal"`
	RecommendedFleet       string  `json:"recommended_fleet"`
	EstimatedUnloadingFee  float64 `json:"estimated_unloading_fee"`
	EstimatedNarrowRoadFee float64 `json:"estimated_narrow_road_fee"`
	EstimatedGrandTotal    float64 `json:"estimated_grand_total"`
}

type ListOrdersFilter struct {
	TenantID   uuid.UUID           `json:"tenant_id"`
	CustomerID *uuid.UUID          `json:"customer_id,omitempty"`
	Status     *domain.OrderStatus `json:"status,omitempty"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
}

type CreateRFQItemCommand struct {
	MaterialName string  `json:"material_name"`
	TargetUOM    string  `json:"target_uom"`
	EstimatedQty float64 `json:"estimated_qty"`
	TargetPrice  float64 `json:"target_price,omitempty"`
	Notes        string  `json:"notes,omitempty"`
}

type CreateRFQCommand struct {
	TenantID        uuid.UUID              `json:"tenant_id"`
	CustomerID      uuid.UUID              `json:"customer_id"`
	ProjectName     string                 `json:"project_name"`
	ProjectAddress  string                 `json:"project_address"`
	BoQDocumentURL  string                 `json:"boq_document_url,omitempty"`
	EstimatedBudget float64                `json:"estimated_budget"`
	Notes           string                 `json:"notes"`
	Items           []CreateRFQItemCommand `json:"items"`
}

type CreateCallOffItemCommand struct {
	SKUID       uuid.UUID `json:"sku_id"`
	ProductName string    `json:"product_name"`
	Quantity    float64   `json:"quantity"`
	UOM         string    `json:"uom"`
}

type CreateCallOffCommand struct {
	OrderID               uuid.UUID                  `json:"order_id"`
	SequenceNo            int                        `json:"sequence_no"`
	RequestedDeliveryDate time.Time                  `json:"requested_delivery_date"`
	Notes                 string                     `json:"notes"`
	Items                 []CreateCallOffItemCommand `json:"items"`
}

type OrderUseCasePort interface {
	CreateOrder(ctx context.Context, cmd CreateOrderCommand) (*domain.Order, error)
	GetOrderByID(ctx context.Context, tenantID, orderID uuid.UUID) (*domain.Order, error)
	GetOrderByOrderNo(ctx context.Context, tenantID uuid.UUID, orderNo string) (*domain.Order, error)
	ListOrders(ctx context.Context, filter ListOrdersFilter) ([]*domain.Order, int64, error)
	UpdateOrderStatus(ctx context.Context, tenantID, orderID uuid.UUID, status domain.OrderStatus, reason string) (*domain.Order, error)
	CancelOrder(ctx context.Context, tenantID, orderID uuid.UUID, reason string) error
	EstimateOrder(ctx context.Context, cmd EstimateOrderCommand) (*OrderEstimate, error)

	CreateRFQ(ctx context.Context, cmd CreateRFQCommand) (*domain.B2BRFQ, error)
	ListRFQs(ctx context.Context, tenantID uuid.UUID, customerID *uuid.UUID) ([]*domain.B2BRFQ, error)
	GetRFQByID(ctx context.Context, tenantID, rfqID uuid.UUID) (*domain.B2BRFQ, error)
	UpdateRFQStatus(ctx context.Context, tenantID, rfqID uuid.UUID, status domain.RFQStatus, quotedAmount float64, quotationNotes string) (*domain.B2BRFQ, error)

	CreateCallOffDelivery(ctx context.Context, cmd CreateCallOffCommand) (*domain.B2BCallOffDelivery, error)
	ListCallOffDeliveries(ctx context.Context, orderID uuid.UUID) ([]*domain.B2BCallOffDelivery, error)
	UpdateCallOffStatus(ctx context.Context, deliveryID uuid.UUID, status domain.CallOffStatus) error
}
