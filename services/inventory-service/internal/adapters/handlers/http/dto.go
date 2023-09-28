package http

import "bangungrid/services/inventory-service/internal/core/domain"

type CreateWarehouseRequest struct {
	Code       string               `json:"code"`
	Name       string               `json:"name"`
	Type       domain.WarehouseType `json:"type"`
	Address    string               `json:"address,omitempty"`
	City       string               `json:"city,omitempty"`
	PostalCode string               `json:"postal_code,omitempty"`
	Latitude   float64              `json:"latitude,omitempty"`
	Longitude  float64              `json:"longitude,omitempty"`
}

type AddStockRequest struct {
	WarehouseID string  `json:"warehouse_id"`
	ProductID   string  `json:"product_id"`
	UOMCode     string  `json:"uom_code"`
	Quantity    float64 `json:"quantity"`
	ReferenceID string  `json:"reference_id,omitempty"`
	Notes       string  `json:"notes,omitempty"`
}

type AdjustShrinkageRequest struct {
	WarehouseID  string  `json:"warehouse_id"`
	ProductID    string  `json:"product_id"`
	UOMCode      string  `json:"uom_code"`
	ShrinkageQty float64 `json:"shrinkage_qty"`
	Reason       string  `json:"reason"`
	ReferenceID  string  `json:"reference_id,omitempty"`
}

type ReserveStockRequest struct {
	WarehouseID string  `json:"warehouse_id"`
	ProductID   string  `json:"product_id"`
	OrderID     string  `json:"order_id"`
	Quantity    float64 `json:"quantity"`
	UOMCode     string  `json:"uom_code"`
	TTLSeconds  int64   `json:"ttl_seconds,omitempty"` // default 900 (15 min)
}
