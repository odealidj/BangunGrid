package ports

import (
	"context"

	"bangungrid/services/auth-service/internal/core/domain"
)

// UserRepositoryPort untuk akses basis data pengguna
type UserRepositoryPort interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, tenantID, email string) (*domain.User, error)
	FindByPhone(ctx context.Context, tenantID, phone string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

// OrganizationRepositoryPort untuk akses basis data organisasi B2B
type OrganizationRepositoryPort interface {
	Create(ctx context.Context, org *domain.Organization) error
	FindByID(ctx context.Context, id string) (*domain.Organization, error)
}

// TokenServicePort untuk penandatanganan kriptografis JWT RS256 & JWKS
type TokenServicePort interface {
	GenerateTokens(claims domain.TokenClaims) (*domain.AuthToken, error)
	ValidateRefreshToken(tokenString string) (*domain.TokenClaims, error)
	GetJWKS() (map[string]any, error)
}

// PasswordHasherPort untuk keamanan kata sandi (Bcrypt / Argon2)
type PasswordHasherPort interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) bool
}

// AuthEventPublisherPort untuk publikasi event ke NATS JetStream
type AuthEventPublisherPort interface {
	PublishUserRegistered(ctx context.Context, user *domain.User) error
}
