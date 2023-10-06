package nats

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"bangungrid/services/order-service/internal/core/domain"
	"bangungrid/services/order-service/internal/core/ports"
)

type StockReservedEvent struct {
	OrderID       uuid.UUID `json:"order_id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	ReservationID uuid.UUID `json:"reservation_id"`
}

type StockFailedEvent struct {
	OrderID  uuid.UUID `json:"order_id"`
	TenantID uuid.UUID `json:"tenant_id"`
	Reason   string    `json:"reason"`
}

type InventoryEventConsumer struct {
	nc       *nats.Conn
	orderSvc ports.OrderUseCasePort
}

func NewInventoryEventConsumer(nc *nats.Conn, orderSvc ports.OrderUseCasePort) *InventoryEventConsumer {
	return &InventoryEventConsumer{
		nc:       nc,
		orderSvc: orderSvc,
	}
}

func (c *InventoryEventConsumer) Start(ctx context.Context) error {
	if c.nc == nil {
		log.Println("[CONSUMER] NATS connection tidak aktif, consumer Saga diabaikan")
		return nil
	}

	// 1. Subscribe to Stock Reserved -> Transition to AWAITING_PAYMENT
	_, err := c.nc.Subscribe("bangungrid.inventory.stock.reserved", func(msg *nats.Msg) {
		var evt StockReservedEvent
		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			log.Printf("[CONSUMER] Gagal unmarshal StockReservedEvent: %v", err)
			return
		}

		log.Printf("[SAGA] Menerima konfirmasi reservasi stok untuk Order %s", evt.OrderID)
		_, err := c.orderSvc.UpdateOrderStatus(ctx, evt.TenantID, evt.OrderID, domain.OrderStatusAwaitingPayment, "Stok berhasil direservasi")
		if err != nil {
			log.Printf("[SAGA] Gagal memperbarui status order %s ke AWAITING_PAYMENT: %v", evt.OrderID, err)
		}
	})
	if err != nil {
		log.Printf("[WARN] Gagal subscribe ke bangungrid.inventory.stock.reserved: %v", err)
	}

	// 2. Subscribe to Stock Failed -> Transition to REJECTED_OUT_OF_STOCK
	_, err = c.nc.Subscribe("bangungrid.inventory.stock.failed", func(msg *nats.Msg) {
		var evt StockFailedEvent
		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			log.Printf("[CONSUMER] Gagal unmarshal StockFailedEvent: %v", err)
			return
		}

		log.Printf("[SAGA] Stok gagal/habis untuk Order %s. Membatalkan pesanan...", evt.OrderID)
		_, err := c.orderSvc.UpdateOrderStatus(ctx, evt.TenantID, evt.OrderID, domain.OrderStatusRejectedStock, evt.Reason)
		if err != nil {
			log.Printf("[SAGA] Gagal memperbarui status order %s ke REJECTED_OUT_OF_STOCK: %v", evt.OrderID, err)
		}
	})
	if err != nil {
		log.Printf("[WARN] Gagal subscribe ke bangungrid.inventory.stock.failed: %v", err)
	}

	log.Println("[CONSUMER] NATS Saga Event Consumer untuk inventory berhasil diaktifkan")
	return nil
}

