package domain

import "errors"

var (
	ErrOrderNotFound           = errors.New("pesanan tidak ditemukan")
	ErrInvalidOrderData        = errors.New("data pesanan tidak valid")
	ErrEmptyOrderItems         = errors.New("item pesanan tidak boleh kosong")
	ErrInvalidQuantity         = errors.New("kuantitas barang harus lebih dari 0")
	ErrInvalidPrice            = errors.New("harga satuan tidak valid")
	ErrInvalidOrderStatus      = errors.New("transisi status pesanan tidak diizinkan")
	ErrOrderAlreadyPaid        = errors.New("pesanan sudah dibayar")
	ErrOrderAlreadyCancelled   = errors.New("pesanan sudah dibatalkan")
	ErrOrderCannotBeCancelled  = errors.New("pesanan dalam status ini tidak dapat dibatalkan")
	ErrTenantMismatch          = errors.New("tenant_id tidak sesuai")
	ErrRFQNotFound             = errors.New("dokumen RFQ tidak ditemukan")
	ErrInvalidRFQStatus        = errors.New("status RFQ tidak valid")
	ErrCallOffNotFound         = errors.New("jadwal pengiriman parsial tidak ditemukan")
	ErrInvalidCallOffDate      = errors.New("tanggal pengiriman parsial harus di masa depan")
	ErrCallOffExceedsQuantity  = errors.New("kuantitas pengiriman parsial melebihi sisa pesanan")
)
