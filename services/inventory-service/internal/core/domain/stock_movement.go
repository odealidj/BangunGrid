package domain

import "time"

type MovementType string

const (
	MovementInboundPO           MovementType = "INBOUND_PO"
	MovementOutboundSO          MovementType = "OUTBOUND_SO"
	MovementTransferIn          MovementType = "TRANSFER_IN"
	MovementTransferOut         MovementType = "TRANSFER_OUT"
	MovementShrinkageAdjustment MovementType = "SHRINKAGE_ADJUSTMENT"
	MovementReturn              MovementType = "RETURN"
)

type StockMovement struct {
	ID           string       `json:"id"`
	TenantID     string       `json:"tenant_id"`
	WarehouseID  string       `json:"warehouse_id"`
	ProductID    string       `json:"product_id"`
	UOMCode      string       `json:"uom_code"`
	Quantity     float64      `json:"quantity"` // positif untuk masuk, negatif untuk keluar/susut
	MovementType MovementType `json:"movement_type"`
	ReferenceID  string       `json:"reference_id"` // PO Number, DO Number, Opname ID
	Notes        string       `json:"notes,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
}
