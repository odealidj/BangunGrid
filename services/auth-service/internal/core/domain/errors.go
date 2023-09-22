package domain

import "errors"

var (
	ErrUserNotFound          = errors.New("pengguna tidak ditemukan")
	ErrUserAlreadyExists     = errors.New("email atau nomor telepon sudah terdaftar")
	ErrInvalidCredentials    = errors.New("kombinasi email/telepon dan kata sandi salah")
	ErrUserInactive          = errors.New("akun pengguna telah dinonaktifkan")
	ErrEmailOrPhoneRequired  = errors.New("email atau nomor telepon wajib diisi")
	ErrPasswordTooShort      = errors.New("kata sandi minimal 8 karakter")
	ErrTenantRequired        = errors.New("tenant id wajib diisi")
	ErrInvalidToken          = errors.New("token otentikasi tidak valid atau kedaluwarsa")
	ErrOrganizationNotFound  = errors.New("organisasi B2B tidak ditemukan")
)
