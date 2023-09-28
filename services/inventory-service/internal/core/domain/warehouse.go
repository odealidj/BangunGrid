package domain

import (
	"strings"
	"time"
)

type WarehouseType string

const (
	WarehouseTypeCentral   WarehouseType = "CENTRAL_WAREHOUSE" // Gudang Utama Distribusi
	WarehouseTypeStockyard WarehouseType = "STOCKYARD"          // Lapangan Terbuka (Pasir, Batu Split, Bata, Besi)
	WarehouseTypeSilo      WarehouseType = "SILO"               // Silo Curah Semen / Fly Ash
	WarehouseTypeBranch    WarehouseType = "BRANCH_STORE"       // Toko Cabang / Outlet Retail
)

type Warehouse struct {
	ID         string        `json:"id"`
	TenantID   string        `json:"tenant_id"`
	Code       string        `json:"code"`
	Name       string        `json:"name"`
	Type       WarehouseType `json:"type"`
	Address    string        `json:"address"`
	City       string        `json:"city"`
	PostalCode string        `json:"postal_code"`
	Latitude   float64       `json:"latitude"`
	Longitude  float64       `json:"longitude"`
	IsActive   bool          `json:"is_active"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

func (w *Warehouse) Validate() error {
	w.Code = strings.TrimSpace(strings.ToUpper(w.Code))
	w.Name = strings.TrimSpace(w.Name)
	if w.TenantID == "" {
		return ErrTenantRequired
	}
	if w.Code == "" || w.Name == "" {
		return ErrWarehouseCodeRequired
	}
	if w.Type == "" {
		w.Type = WarehouseTypeCentral
	}
	return nil
}
