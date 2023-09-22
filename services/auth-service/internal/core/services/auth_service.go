package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"bangungrid/services/auth-service/internal/core/domain"
	"bangungrid/services/auth-service/internal/core/ports"
)

type AuthService struct {
	userRepo       ports.UserRepositoryPort
	orgRepo        ports.OrganizationRepositoryPort
	tokenService   ports.TokenServicePort
	passwordHasher ports.PasswordHasherPort
	eventPublisher ports.AuthEventPublisherPort
}

func NewAuthService(
	userRepo ports.UserRepositoryPort,
	orgRepo ports.OrganizationRepositoryPort,
	tokenService ports.TokenServicePort,
	hasher ports.PasswordHasherPort,
	publisher ports.AuthEventPublisherPort,
) ports.AuthUseCasePort {
	return &AuthService{
		userRepo:       userRepo,
		orgRepo:        orgRepo,
		tokenService:   tokenService,
		passwordHasher: hasher,
		eventPublisher: publisher,
	}
}

func (s *AuthService) Register(ctx context.Context, cmd ports.RegisterUserCommand) (*domain.User, *domain.AuthToken, error) {
	if len(cmd.Password) < 8 {
		return nil, nil, domain.ErrPasswordTooShort
	}

	cmd.Email = strings.TrimSpace(strings.ToLower(cmd.Email))
	cmd.Phone = strings.TrimSpace(cmd.Phone)

	// 1. Cek apakah user sudah terdaftar
	if cmd.Email != "" {
		existing, _ := s.userRepo.FindByEmail(ctx, cmd.TenantID, cmd.Email)
		if existing != nil {
			return nil, nil, domain.ErrUserAlreadyExists
		}
	}
	if cmd.Phone != "" {
		existing, _ := s.userRepo.FindByPhone(ctx, cmd.TenantID, cmd.Phone)
		if existing != nil {
			return nil, nil, domain.ErrUserAlreadyExists
		}
	}

	// 2. Hash Password
	hashedPassword, err := s.passwordHasher.Hash(cmd.Password)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now().UTC()
	var orgID *string

	// 3. Jika akun Kontraktor B2B, buat Organisasi perusahaan terlebih dahulu
	if cmd.UserType == domain.UserTypeB2BContractor && cmd.OrganizationName != "" {
		newOrgID := uuid.New().String()
		org := &domain.Organization{
			ID:                 newOrgID,
			TenantID:           cmd.TenantID,
			Name:               cmd.OrganizationName,
			NPWP:               cmd.NPWP,
			NIB:                cmd.NIB,
			LegalAddress:       cmd.LegalAddress,
			ContactEmail:       cmd.Email,
			ContactPhone:       cmd.Phone,
			VerificationStatus: domain.VerificationStatusPending,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err := s.orgRepo.Create(ctx, org); err != nil {
			return nil, nil, err
		}
		orgID = &newOrgID
	}

	// 4. Buat entitas User
	role := "CUSTOMER"
	if cmd.UserType == domain.UserTypeB2BContractor {
		role = "CONTRACTOR_ADMIN"
	}

	user := &domain.User{
		ID:             uuid.New().String(),
		TenantID:       cmd.TenantID,
		OrganizationID: orgID,
		Email:          cmd.Email,
		Phone:          cmd.Phone,
		PasswordHash:   hashedPassword,
		UserType:       cmd.UserType,
		Role:           role,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := user.Validate(); err != nil {
		return nil, nil, err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	// 5. Terbitkan Token JWT RS256
	var orgIDStr string
	if orgID != nil {
		orgIDStr = *orgID
	}

	claims := domain.TokenClaims{
		UserID:         user.ID,
		TenantID:       user.TenantID,
		OrganizationID: orgIDStr,
		UserType:       user.UserType,
		Role:           user.Role,
		Email:          user.Email,
	}

	tokens, err := s.tokenService.GenerateTokens(claims)
	if err != nil {
		return nil, nil, err
	}

	// 6. Publikasikan event asinkron ke NATS JetStream
	_ = s.eventPublisher.PublishUserRegistered(ctx, user)

	return user, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, cmd ports.LoginCommand) (*domain.User, *domain.AuthToken, error) {
	loginIdentifier := strings.TrimSpace(strings.ToLower(cmd.Login))
	var user *domain.User
	var err error

	// Cari berdasarkan email atau nomor telepon
	if strings.Contains(loginIdentifier, "@") {
		user, err = s.userRepo.FindByEmail(ctx, cmd.TenantID, loginIdentifier)
	} else {
		user, err = s.userRepo.FindByPhone(ctx, cmd.TenantID, loginIdentifier)
	}

	if err != nil || user == nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, nil, domain.ErrUserInactive
	}

	// Verifikasi kata sandi
	if !s.passwordHasher.Compare(user.PasswordHash, cmd.Password) {
		return nil, nil, domain.ErrInvalidCredentials
	}

	// Terbitkan token baru
	var orgIDStr string
	if user.OrganizationID != nil {
		orgIDStr = *user.OrganizationID
	}

	claims := domain.TokenClaims{
		UserID:         user.ID,
		TenantID:       user.TenantID,
		OrganizationID: orgIDStr,
		UserType:       user.UserType,
		Role:           user.Role,
		Email:          user.Email,
	}

	tokens, err := s.tokenService.GenerateTokens(claims)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, cmd ports.RefreshTokenCommand) (*domain.AuthToken, error) {
	claims, err := s.tokenService.ValidateRefreshToken(cmd.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil || user == nil || !user.IsActive {
		return nil, domain.ErrInvalidToken
	}

	var orgIDStr string
	if user.OrganizationID != nil {
		orgIDStr = *user.OrganizationID
	}

	newClaims := domain.TokenClaims{
		UserID:         user.ID,
		TenantID:       user.TenantID,
		OrganizationID: orgIDStr,
		UserType:       user.UserType,
		Role:           user.Role,
		Email:          user.Email,
	}

	return s.tokenService.GenerateTokens(newClaims)
}

func (s *AuthService) GetJWKS(ctx context.Context) (map[string]any, error) {
	return s.tokenService.GetJWKS()
}

func (s *AuthService) GetUserProfile(ctx context.Context, userID string) (*domain.User, *domain.Organization, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, nil, domain.ErrUserNotFound
	}

	var org *domain.Organization
	if user.OrganizationID != nil && *user.OrganizationID != "" {
		org, _ = s.orgRepo.FindByID(ctx, *user.OrganizationID)
	}

	return user, org, nil
}
