package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type OrderType string

const (
	CustomerTypeB2CRetail     OrderType = "B2C_RETAIL"
	CustomerTypeB2BContractor OrderType = "B2B_CONTRACTOR"
)

type OrderStatus string

const (
	OrderStatusDraft              OrderStatus = "DRAFT"
	OrderStatusPending            OrderStatus = "PENDING"
	OrderStatusAwaitingPayment    OrderStatus = "AWAITING_PAYMENT"
	OrderStatusPaid               OrderStatus = "PAID"
	OrderStatusProcessing         OrderStatus = "PROCESSING"
	OrderStatusPartiallyDelivered OrderStatus = "PARTIALLY_DELIVERED"
	OrderStatusCompleted          OrderStatus = "COMPLETED"
	OrderStatusCancelled          OrderStatus = "CANCELLED"
	OrderStatusRejectedStock      OrderStatus = "REJECTED_OUT_OF_STOCK"
)

type Order struct {
	ID                   uuid.UUID   `json:"id"`
	TenantID             uuid.UUID   `json:"tenant_id"`
	BranchID             uuid.UUID   `json:"branch_id"`
	OrderNo              string      `json:"order_no"`
	CustomerID           uuid.UUID   `json:"customer_id"`
	CustomerType         OrderType   `json:"customer_type"`
	Status               OrderStatus `json:"status"`
	Items                []OrderItem `json:"items"`
	TotalAmount          float64     `json:"total_amount"`
	TotalWeightKg        float64     `json:"total_weight_kg"`
	TotalVolumeM3        float64     `json:"total_volume_m3"`
	UnloadingFee         float64     `json:"unloading_fee"`
	NarrowRoadFee        float64     `json:"narrow_road_fee"`
	ShippingFee          float64     `json:"shipping_fee"`
	GrandTotal           float64     `json:"grand_total"`
	DeliveryAddress      string      `json:"delivery_address"`
	DeliveryLatitude     float64     `json:"delivery_latitude"`
	DeliveryLongitude    float64     `json:"delivery_longitude"`
	Notes                string      `json:"notes"`
	CancelReason         string      `json:"cancel_reason,omitempty"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
}

func GenerateOrderNo(t time.Time, seq int64) string {
	return fmt.Sprintf("BG-%s-%04d", t.Format("20060102"), seq%10000)
}

func NewOrder(
	tenantID uuid.UUID,
	branchID uuid.UUID,
	customerID uuid.UUID,
	customerType OrderType,
	orderNo string,
	deliveryAddress string,
	lat float64,
	lng float64,
	shippingFee float64,
	unloadingFee float64,
	narrowRoadFee float64,
	notes string,
) *Order {
	now := time.Now().UTC()
	return &Order{
		ID:                uuid.New(),
		TenantID:          tenantID,
		BranchID:          branchID,
		OrderNo:           orderNo,
		CustomerID:        customerID,
		CustomerType:      customerType,
		Status:            OrderStatusPending,
		Items:             make([]OrderItem, 0),
		TotalAmount:       0,
		TotalWeightKg:     0,
		TotalVolumeM3:     0,
		ShippingFee:       shippingFee,
		UnloadingFee:      unloadingFee,
		NarrowRoadFee:     narrowRoadFee,
		GrandTotal:        shippingFee + unloadingFee + narrowRoadFee,
		DeliveryAddress:   deliveryAddress,
		DeliveryLatitude:  lat,
		DeliveryLongitude: lng,
		Notes:             notes,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func (o *Order) AddItem(item OrderItem) {
	item.OrderID = o.ID
	o.Items = append(o.Items, item)
	o.RecalculateTotals()
}

func (o *Order) RecalculateTotals() {
	var totalAmount, totalWeight, totalVolume float64
	for _, it := range o.Items {
		totalAmount += it.Subtotal
		totalWeight += it.WeightKgSnapshot
		totalVolume += it.VolumeM3Snapshot
	}

	o.TotalAmount = totalAmount
	o.TotalWeightKg = totalWeight
	o.TotalVolumeM3 = totalVolume
	o.GrandTotal = o.TotalAmount + o.ShippingFee + o.UnloadingFee + o.NarrowRoadFee
	o.UpdatedAt = time.Now().UTC()
}

func (o *Order) CanTransitionTo(next OrderStatus) bool {
	switch o.Status {
	case OrderStatusPending:
		return next == OrderStatusAwaitingPayment || next == OrderStatusPaid || next == OrderStatusCancelled || next == OrderStatusRejectedStock
	case OrderStatusAwaitingPayment:
		return next == OrderStatusPaid || next == OrderStatusCancelled
	case OrderStatusPaid:
		return next == OrderStatusProcessing || next == OrderStatusCancelled
	case OrderStatusProcessing:
		return next == OrderStatusPartiallyDelivered || next == OrderStatusCompleted || next == OrderStatusCancelled
	case OrderStatusPartiallyDelivered:
		return next == OrderStatusCompleted || next == OrderStatusCancelled
	case OrderStatusCompleted, OrderStatusCancelled, OrderStatusRejectedStock:
		return false
	default:
		return false
	}
}

func (o *Order) UpdateStatus(newStatus OrderStatus) error {
	if !o.CanTransitionTo(newStatus) {
		return fmt.Errorf("%w: tidak dapat beralih dari %s ke %s", ErrInvalidOrderStatus, o.Status, newStatus)
	}
	o.Status = newStatus
	o.UpdatedAt = time.Now().UTC()
	return nil
}

func (o *Order) Cancel(reason string) error {
	if o.Status == OrderStatusCompleted || o.Status == OrderStatusCancelled {
		return ErrOrderCannotBeCancelled
	}
	o.Status = OrderStatusCancelled
	o.CancelReason = reason
	o.UpdatedAt = time.Now().UTC()
	return nil
}
