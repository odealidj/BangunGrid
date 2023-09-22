package domain

import (
	"strings"
	"time"
)

type UserType string

const (
	UserTypeB2CRetail     UserType = "B2C_RETAIL"
	UserTypeB2BContractor UserType = "B2B_CONTRACTOR"
	UserTypeInternalStaff UserType = "INTERNAL_STAFF"
)

type User struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	OrganizationID *string   `json:"organization_id,omitempty"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	PasswordHash   string    `json:"-"`
	UserType       UserType  `json:"user_type"`
	Role           string    `json:"role"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (u *User) Validate() error {
	u.Email = strings.TrimSpace(strings.ToLower(u.Email))
	u.Phone = strings.TrimSpace(u.Phone)

	if u.Email == "" && u.Phone == "" {
		return ErrEmailOrPhoneRequired
	}
	if u.TenantID == "" {
		return ErrTenantRequired
	}
	if u.UserType == "" {
		u.UserType = UserTypeB2CRetail
	}
	if u.Role == "" {
		u.Role = "CUSTOMER"
	}
	return nil
}
