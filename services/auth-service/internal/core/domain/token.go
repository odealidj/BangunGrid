package domain

import "time"

type AuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"` // durasi dalam detik
	ExpiresAt    time.Time `json:"expires_at"`
}

type TokenClaims struct {
	UserID         string   `json:"sub"`
	TenantID       string   `json:"tenant_id"`
	OrganizationID string   `json:"organization_id,omitempty"`
	UserType       UserType `json:"user_type"`
	Role           string   `json:"role"`
	Email          string   `json:"email"`
}
