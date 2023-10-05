package domain

import (
	"time"

	"github.com/google/uuid"
)

type CallOffStatus string

const (
	CallOffStatusScheduled  CallOffStatus = "SCHEDULED"
	CallOffStatusDispatched CallOffStatus = "DISPATCHED"
	CallOffStatusCompleted  CallOffStatus = "COMPLETED"
	CallOffStatusCancelled  CallOffStatus = "CANCELLED"
)

type CallOffItem struct {
	ID          uuid.UUID `json:"id"`
	CallOffID   uuid.UUID `json:"call_off_id"`
	SKUID       uuid.UUID `json:"sku_id"`
	ProductName string    `json:"product_name"`
	Quantity    float64   `json:"quantity"`
	UOM         string    `json:"uom"`
}

type B2BCallOffDelivery struct {
	ID                    uuid.UUID     `json:"id"`
	OrderID               uuid.UUID     `json:"order_id"`
	SequenceNo            int           `json:"sequence_no"`
	RequestedDeliveryDate time.Time     `json:"requested_delivery_date"`
	Status                CallOffStatus `json:"status"`
	Notes                 string        `json:"notes,omitempty"`
	Items                 []CallOffItem `json:"items"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`
}

func NewCallOffDelivery(
	orderID uuid.UUID,
	sequenceNo int,
	deliveryDate time.Time,
	notes string,
) *B2BCallOffDelivery {
	now := time.Now().UTC()
	return &B2BCallOffDelivery{
		ID:                    uuid.New(),
		OrderID:               orderID,
		SequenceNo:            sequenceNo,
		RequestedDeliveryDate: deliveryDate,
		Status:                CallOffStatusScheduled,
		Notes:                 notes,
		Items:                 make([]CallOffItem, 0),
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

func (c *B2BCallOffDelivery) AddItem(skuID uuid.UUID, productName string, qty float64, uom string) {
	c.Items = append(c.Items, CallOffItem{
		ID:          uuid.New(),
		CallOffID:   c.ID,
		SKUID:       skuID,
		ProductName: productName,
		Quantity:    qty,
		UOM:         uom,
	})
	c.UpdatedAt = time.Now().UTC()
}
