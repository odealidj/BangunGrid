package http

import "bangungrid/services/auth-service/internal/core/domain"

type RegisterRequest struct {
	TenantID         string          `json:"tenant_id"`
	Email            string          `json:"email"`
	Phone            string          `json:"phone"`
	Password         string          `json:"password"`
	UserType         domain.UserType `json:"user_type"`
	OrganizationName string          `json:"organization_name,omitempty"`
	NPWP             string          `json:"npwp,omitempty"`
	NIB              string          `json:"nib,omitempty"`
	LegalAddress     string          `json:"legal_address,omitempty"`
}

type LoginRequest struct {
	TenantID string `json:"tenant_id"`
	Login    string `json:"login"` // Email atau No HP
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponseData struct {
	User         *domain.User      `json:"user"`
	Organization *domain.Organization `json:"organization,omitempty"`
	Tokens       *domain.AuthToken `json:"tokens"`
}
