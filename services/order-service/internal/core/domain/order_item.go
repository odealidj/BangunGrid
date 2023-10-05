package domain

import (
	"github.com/google/uuid"
)

type OrderItem struct {
	ID                  uuid.UUID `json:"id"`
	OrderID             uuid.UUID `json:"order_id"`
	SKUID               uuid.UUID `json:"sku_id"`
	ProductNameSnapshot string    `json:"product_name_snapshot"`
	UOMCodeSnapshot     string    `json:"uom_code_snapshot"`
	Quantity            float64   `json:"quantity"`
	UnitPriceSnapshot   float64   `json:"unit_price_snapshot"`
	Subtotal            float64   `json:"subtotal"`
	WeightKgSnapshot    float64   `json:"weight_kg_snapshot"`
	VolumeM3Snapshot    float64   `json:"volume_m3_snapshot"`
}

func NewOrderItem(
	skuID uuid.UUID,
	name string,
	uom string,
	qty float64,
	unitPrice float64,
	weightKg float64,
	volumeM3 float64,
) (*OrderItem, error) {
	if qty <= 0 {
		return nil, ErrInvalidQuantity
	}
	if unitPrice < 0 {
		return nil, ErrInvalidPrice
	}

	subtotal := qty * unitPrice
	totalWeight := qty * weightKg
	totalVolume := qty * volumeM3

	return &OrderItem{
		ID:                  uuid.New(),
		SKUID:               skuID,
		ProductNameSnapshot: name,
		UOMCodeSnapshot:     uom,
		Quantity:            qty,
		UnitPriceSnapshot:   unitPrice,
		Subtotal:            subtotal,
		WeightKgSnapshot:    totalWeight,
		VolumeM3Snapshot:    totalVolume,
	}, nil
}
