package domain

import (
	"math"
	"time"
)

type StockItem struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	WarehouseID    string    `json:"warehouse_id"`
	ProductID      string    `json:"product_id"`
	UOMCode        string    `json:"uom_code"` // Base UoM (e.g. SAK, KG, M3, BTG)
	PhysicalQty    float64   `json:"physical_qty"`
	ReservedQty    float64   `json:"reserved_qty"`
	AvailableQty   float64   `json:"available_qty"` // PhysicalQty - ReservedQty
	SafetyStock    float64   `json:"safety_stock"`
	MinReorderQty  float64   `json:"min_reorder_qty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (s *StockItem) Recalculate() {
	s.AvailableQty = math.Round((s.PhysicalQty-s.ReservedQty)*1000) / 1000
	if s.AvailableQty < 0 {
		s.AvailableQty = 0
	}
}

func (s *StockItem) CanReserve(qty float64) bool {
	s.Recalculate()
	return s.AvailableQty >= qty
}

func (s *StockItem) Reserve(qty float64) error {
	if qty <= 0 {
		return ErrInvalidQuantity
	}
	s.Recalculate()
	if s.AvailableQty < qty {
		return ErrInsufficientStock
	}
	s.ReservedQty = math.Round((s.ReservedQty+qty)*1000) / 1000
	s.Recalculate()
	s.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *StockItem) ReleaseReservation(qty float64) error {
	if qty <= 0 {
		return ErrInvalidQuantity
	}
	if s.ReservedQty < qty {
		s.ReservedQty = 0
	} else {
		s.ReservedQty = math.Round((s.ReservedQty-qty)*1000) / 1000
	}
	s.Recalculate()
	s.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *StockItem) DeductCommitted(qty float64) error {
	if qty <= 0 {
		return ErrInvalidQuantity
	}
	s.PhysicalQty = math.Round((s.PhysicalQty-qty)*1000) / 1000
	if s.PhysicalQty < 0 {
		s.PhysicalQty = 0
	}
	if s.ReservedQty >= qty {
		s.ReservedQty = math.Round((s.ReservedQty-qty)*1000) / 1000
	} else {
		s.ReservedQty = 0
	}
	s.Recalculate()
	s.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *StockItem) AddPhysical(qty float64) error {
	if qty <= 0 {
		return ErrInvalidQuantity
	}
	s.PhysicalQty = math.Round((s.PhysicalQty+qty)*1000) / 1000
	s.Recalculate()
	s.UpdatedAt = time.Now().UTC()
	return nil
}
