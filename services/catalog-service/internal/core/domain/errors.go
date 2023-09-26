package domain

import "errors"

var (
	ErrTenantRequired          = errors.New("tenant_id wajib diisi")
	ErrCategoryNameRequired    = errors.New("nama kategori wajib diisi")
	ErrCategoryNotFound        = errors.New("kategori tidak ditemukan")
	ErrProductNameRequired     = errors.New("nama produk wajib diisi")
	ErrBaseUOMRequired         = errors.New("base UoM wajib diisi")
	ErrProductNotFound         = errors.New("produk tidak ditemukan")
	ErrInvalidMultiplier       = errors.New("conversion multiplier harus lebih besar dari 0")
	ErrBulkMaterialNotFound    = errors.New("matriks massa jenis material curah tidak ditemukan")
	ErrInvalidDensity          = errors.New("densitas material curah harus lebih besar dari 0")
)
