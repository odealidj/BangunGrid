package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"bangungrid/services/order-service/internal/core/domain"
)

type OrderCreatedEvent struct {
	EventID       uuid.UUID   `json:"event_id"`
	EventType     string      `json:"event_type"`
	Timestamp     time.Time   `json:"timestamp"`
	OrderID       uuid.UUID   `json:"order_id"`
	TenantID      uuid.UUID   `json:"tenant_id"`
	BranchID      uuid.UUID   `json:"branch_id"`
	OrderNo       string      `json:"order_no"`
	CustomerID    uuid.UUID   `json:"customer_id"`
	CustomerType  string      `json:"customer_type"`
	Status        string      `json:"status"`
	TotalAmount   float64     `json:"total_amount"`
	TotalWeightKg float64     `json:"total_weight_kg"`
	TotalVolumeM3 float64     `json:"total_volume_m3"`
	GrandTotal    float64     `json:"grand_total"`
	Items         []OrderItem `json:"items"`
}

type OrderItem struct {
	SKUID       uuid.UUID `json:"sku_id"`
	ProductName string    `json:"product_name"`
	UOM         string    `json:"uom"`
	Quantity    float64   `json:"quantity"`
	Subtotal    float64   `json:"subtotal"`
}

type OrderStatusChangedEvent struct {
	EventID   uuid.UUID `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	OrderID   uuid.UUID `json:"order_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	OrderNo   string    `json:"order_no"`
	OldStatus string    `json:"old_status,omitempty"`
	NewStatus string    `json:"new_status"`
	Reason    string    `json:"reason,omitempty"`
}

type NatsOrderPublisher struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func NewNatsOrderPublisher(nc *nats.Conn) *NatsOrderPublisher {
	if nc == nil {
		return &NatsOrderPublisher{}
	}
	js, err := nc.JetStream()
	if err != nil {
		log.Printf("[WARN] NATS JetStream tidak dapat diinisialisasi: %v, fallback ke Core NATS", err)
	}
	return &NatsOrderPublisher{nc: nc, js: js}
}

func (p *NatsOrderPublisher) PublishOrderCreated(ctx context.Context, order *domain.Order) error {
	if p.nc == nil {
		return nil
	}

	items := make([]OrderItem, len(order.Items))
	for i, it := range order.Items {
		items[i] = OrderItem{
			SKUID:       it.SKUID,
			ProductName: it.ProductNameSnapshot,
			UOM:         it.UOMCodeSnapshot,
			Quantity:    it.Quantity,
			Subtotal:    it.Subtotal,
		}
	}

	evt := OrderCreatedEvent{
		EventID:       uuid.New(),
		EventType:     "ORDER_CREATED",
		Timestamp:     time.Now().UTC(),
		OrderID:       order.ID,
		TenantID:      order.TenantID,
		BranchID:      order.BranchID,
		OrderNo:       order.OrderNo,
		CustomerID:    order.CustomerID,
		CustomerType:  string(order.CustomerType),
		Status:        string(order.Status),
		TotalAmount:   order.TotalAmount,
		TotalWeightKg: order.TotalWeightKg,
		TotalVolumeM3: order.TotalVolumeM3,
		GrandTotal:    order.GrandTotal,
		Items:         items,
	}

	data, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("gagal serialisasi event order created: %w", err)
	}

	subject := "bangungrid.orders.created"
	if p.js != nil {
		_, err = p.js.Publish(subject, data)
	} else {
		err = p.nc.Publish(subject, data)
	}
	if err != nil {
		log.Printf("[WARN] Gagal publish event %s ke NATS: %v", subject, err)
	}
	return nil
}

func (p *NatsOrderPublisher) PublishOrderStatusChanged(ctx context.Context, order *domain.Order) error {
	if p.nc == nil {
		return nil
	}

	evt := OrderStatusChangedEvent{
		EventID:   uuid.New(),
		EventType: "ORDER_STATUS_CHANGED",
		Timestamp: time.Now().UTC(),
		OrderID:   order.ID,
		TenantID:  order.TenantID,
		OrderNo:   order.OrderNo,
		NewStatus: string(order.Status),
	}

	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	subject := "bangungrid.orders.status_updated"
	if p.js != nil {
		_, _ = p.js.Publish(subject, data)
	} else {
		_ = p.nc.Publish(subject, data)
	}
	return nil
}

func (p *NatsOrderPublisher) PublishOrderCancelled(ctx context.Context, order *domain.Order) error {
	if p.nc == nil {
		return nil
	}

	evt := OrderStatusChangedEvent{
		EventID:   uuid.New(),
		EventType: "ORDER_CANCELLED",
		Timestamp: time.Now().UTC(),
		OrderID:   order.ID,
		TenantID:  order.TenantID,
		OrderNo:   order.OrderNo,
		NewStatus: string(domain.OrderStatusCancelled),
		Reason:    order.CancelReason,
	}

	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	subject := "bangungrid.orders.cancelled"
	if p.js != nil {
		_, _ = p.js.Publish(subject, data)
	} else {
		_ = p.nc.Publish(subject, data)
	}
	return nil
}
