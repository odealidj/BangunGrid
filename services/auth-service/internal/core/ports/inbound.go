package ports

import (
	"context"

	"bangungrid/services/auth-service/internal/core/domain"
)

type RegisterUserCommand struct {
	TenantID         string
	Email            string
	Phone            string
	Password         string
	UserType         domain.UserType
	OrganizationName string // Khusus jika pendaftaran B2B
	NPWP             string
	NIB              string
	LegalAddress     string
}

type LoginCommand struct {
	TenantID string
	Login    string // Email atau Phone
	Password string
}

type RefreshTokenCommand struct {
	RefreshToken string
}

// AuthUseCasePort mendefinisikan antarmuka Use Case yang diekspos ke Driving Adapter
type AuthUseCasePort interface {
	Register(ctx context.Context, cmd RegisterUserCommand) (*domain.User, *domain.AuthToken, error)
	Login(ctx context.Context, cmd LoginCommand) (*domain.User, *domain.AuthToken, error)
	RefreshToken(ctx context.Context, cmd RefreshTokenCommand) (*domain.AuthToken, error)
	GetJWKS(ctx context.Context) (map[string]any, error)
	GetUserProfile(ctx context.Context, userID string) (*domain.User, *domain.Organization, error)
}
