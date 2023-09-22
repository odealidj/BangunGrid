package domain

import "time"

type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "PENDING"
	VerificationStatusApproved VerificationStatus = "APPROVED"
	VerificationStatusRejected VerificationStatus = "REJECTED"
)

type Organization struct {
	ID                 string             `json:"id"`
	TenantID           string             `json:"tenant_id"`
	Name               string             `json:"name"`
	NPWP               string             `json:"npwp"`
	NIB                string             `json:"nib"`
	LegalAddress       string             `json:"legal_address"`
	ContactEmail       string             `json:"contact_email"`
	ContactPhone       string             `json:"contact_phone"`
	VerificationStatus VerificationStatus `json:"verification_status"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type Branch struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}
