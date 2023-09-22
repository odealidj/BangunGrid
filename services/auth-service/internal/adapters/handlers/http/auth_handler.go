package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"bangungrid/pkg/middleware"
	"bangungrid/pkg/response"
	"bangungrid/services/auth-service/internal/core/domain"
	"bangungrid/services/auth-service/internal/core/ports"
)

type AuthHandler struct {
	useCase ports.AuthUseCasePort
}

func NewAuthHandler(useCase ports.AuthUseCasePort) *AuthHandler {
	return &AuthHandler{useCase: useCase}
}

// Register menangani pendaftaran B2C dan B2B
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	if req.TenantID == "" {
		req.TenantID = "00000000-0000-0000-0000-000000000001" // Default master tenant
	}

	cmd := ports.RegisterUserCommand{
		TenantID:         req.TenantID,
		Email:            req.Email,
		Phone:            req.Phone,
		Password:         req.Password,
		UserType:         req.UserType,
		OrganizationName: req.OrganizationName,
		NPWP:             req.NPWP,
		NIB:              req.NIB,
		LegalAddress:     req.LegalAddress,
	}

	user, tokens, err := h.useCase.Register(r.Context(), cmd)
	if err != nil {
		status := http.StatusUnprocessableEntity
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			status = http.StatusConflict
		}
		response.JSONError(w, r, status, err.Error(), "REGISTRATION_FAILED", nil)
		return
	}

	resData := AuthResponseData{
		User:   user,
		Tokens: tokens,
	}

	response.JSONSuccess(w, r, http.StatusCreated, "Pendaftaran berhasil", resData)
}

// Login menangani autentikasi akun
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	if req.TenantID == "" {
		req.TenantID = "00000000-0000-0000-0000-000000000001"
	}

	loginIdentifier := req.Login
	if loginIdentifier == "" {
		if req.Email != "" {
			loginIdentifier = req.Email
		} else if req.Phone != "" {
			loginIdentifier = req.Phone
		}
	}

	cmd := ports.LoginCommand{
		TenantID: req.TenantID,
		Login:    loginIdentifier,
		Password: req.Password,
	}

	user, tokens, err := h.useCase.Login(r.Context(), cmd)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, domain.ErrUserInactive) {
			status = http.StatusForbidden
		}
		response.JSONError(w, r, status, err.Error(), "LOGIN_FAILED", nil)
		return
	}

	resData := AuthResponseData{
		User:   user,
		Tokens: tokens,
	}

	response.JSONSuccess(w, r, http.StatusOK, "Login berhasil", resData)
}

// RefreshToken memperbarui access token menggunakan refresh token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	cmd := ports.RefreshTokenCommand{
		RefreshToken: req.RefreshToken,
	}

	tokens, err := h.useCase.RefreshToken(r.Context(), cmd)
	if err != nil {
		response.JSONError(w, r, http.StatusUnauthorized, err.Error(), "INVALID_REFRESH_TOKEN", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Token berhasil diperbarui", tokens)
}

// GetJWKS menyajikan kunci publik RSA untuk KrakenD API Gateway
func (h *AuthHandler) GetJWKS(w http.ResponseWriter, r *http.Request) {
	jwks, err := h.useCase.GetJWKS(r.Context())
	if err != nil {
		response.JSONError(w, r, http.StatusInternalServerError, "Gagal mengambil JWKS", "INTERNAL_ERROR", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(jwks)
}

// GetProfile mengambil profil pengguna saat ini
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.JSONError(w, r, http.StatusUnauthorized, "Identitas pengguna tidak ditemukan pada request header", "UNAUTHORIZED", nil)
		return
	}

	user, org, err := h.useCase.GetUserProfile(r.Context(), userID)
	if err != nil {
		response.JSONError(w, r, http.StatusNotFound, err.Error(), "USER_NOT_FOUND", nil)
		return
	}

	data := map[string]any{
		"user": user,
	}
	if org != nil {
		data["organization"] = org
	}

	response.JSONSuccess(w, r, http.StatusOK, "Data profil berhasil diambil", data)
}
