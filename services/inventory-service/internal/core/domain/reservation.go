package domain

import "time"

type ReservationStatus string

const (
	ReservationStatusActive    ReservationStatus = "ACTIVE"    // Stok terpesan (soft-lock)
	ReservationStatusCommitted ReservationStatus = "COMMITTED" // Pesanan dibayar/dikirim, stok fisik resmi dipotong
	ReservationStatusCancelled ReservationStatus = "CANCELLED" // Pesanan kadaluarsa/dibatalkan, reservasi dirilis
)

type StockReservation struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenant_id"`
	WarehouseID string            `json:"warehouse_id"`
	ProductID   string            `json:"product_id"`
	OrderID     string            `json:"order_id"`
	Quantity    float64           `json:"quantity"`
	UOMCode     string            `json:"uom_code"`
	Status      ReservationStatus `json:"status"`
	ExpiresAt   time.Time         `json:"expires_at"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
