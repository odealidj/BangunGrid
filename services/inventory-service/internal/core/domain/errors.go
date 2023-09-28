package domain

import "errors"

var (
	ErrWarehouseNotFound            = errors.New("gudang / stockyard tidak ditemukan")
	ErrStockNotFound                = errors.New("data stok produk di gudang tidak ditemukan")
	ErrInsufficientStock            = errors.New("stok tersedia (available stock) tidak mencukupi untuk reservasi")
	ErrInvalidQuantity              = errors.New("kuantitas stok harus lebih besar dari 0")
	ErrReservationNotFound          = errors.New("reservasi stok tidak ditemukan")
	ErrReservationAlreadyProcessed  = errors.New("reservasi stok sudah selesai (committed) atau dibatalkan")
	ErrDistributedLockFailed        = errors.New("gagal memperoleh distributed lock stok, coba beberapa saat lagi")
	ErrTenantRequired               = errors.New("tenant ID wajib diisi")
	ErrWarehouseCodeRequired        = errors.New("kode dan nama gudang wajib diisi")
)
