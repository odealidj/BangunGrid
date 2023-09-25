package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"bangungrid/services/auth-service/internal/core/domain"
	"bangungrid/services/auth-service/internal/core/ports"
	"bangungrid/services/auth-service/internal/core/services"
)

// ==========================================
// Mocks for Ports
// ==========================================

type mockUserRepo struct {
	usersByID    map[string]*domain.User
	usersByEmail map[string]*domain.User
	usersByPhone map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByID:    make(map[string]*domain.User),
		usersByEmail: make(map[string]*domain.User),
		usersByPhone: make(map[string]*domain.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	m.usersByID[u.ID] = u
	if u.Email != "" {
		m.usersByEmail[u.TenantID+":"+u.Email] = u
	}
	if u.Phone != "" {
		m.usersByPhone[u.TenantID+":"+u.Phone] = u
	}
	return nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	u, ok := m.usersByID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, tenantID, email string) (*domain.User, error) {
	u, ok := m.usersByEmail[tenantID+":"+email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) FindByPhone(ctx context.Context, tenantID, phone string) (*domain.User, error) {
	u, ok := m.usersByPhone[tenantID+":"+phone]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) Update(ctx context.Context, u *domain.User) error {
	m.usersByID[u.ID] = u
	return nil
}

type mockOrgRepo struct {
	orgs map[string]*domain.Organization
}

func newMockOrgRepo() *mockOrgRepo {
	return &mockOrgRepo{orgs: make(map[string]*domain.Organization)}
}

func (m *mockOrgRepo) Create(ctx context.Context, org *domain.Organization) error {
	m.orgs[org.ID] = org
	return nil
}

func (m *mockOrgRepo) FindByID(ctx context.Context, id string) (*domain.Organization, error) {
	org, ok := m.orgs[id]
	if !ok {
		return nil, errors.New("organization not found")
	}
	return org, nil
}

type mockTokenService struct{}

func (m *mockTokenService) GenerateTokens(claims domain.TokenClaims) (*domain.AuthToken, error) {
	return &domain.AuthToken{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    86400,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}, nil
}

func (m *mockTokenService) ValidateRefreshToken(tokenString string) (*domain.TokenClaims, error) {
	if tokenString == "valid-refresh-token" {
		return &domain.TokenClaims{
			UserID:   "user-123",
			TenantID: "tenant-123",
			Role:     "CUSTOMER",
			Email:    "budi@example.com",
		}, nil
	}
	return nil, domain.ErrInvalidToken
}

func (m *mockTokenService) GetJWKS() (map[string]any, error) {
	return map[string]any{"keys": []any{}}, nil
}

type mockPasswordHasher struct{}

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	return "hashed-" + password, nil
}

func (m *mockPasswordHasher) Compare(hashedPassword, password string) bool {
	return hashedPassword == "hashed-"+password
}

type mockEventPublisher struct {
	publishedUsers []*domain.User
}

func (m *mockEventPublisher) PublishUserRegistered(ctx context.Context, user *domain.User) error {
	m.publishedUsers = append(m.publishedUsers, user)
	return nil
}

// ==========================================
// Test Cases
// ==========================================

func setupAuthService() (*services.AuthService, *mockUserRepo, *mockOrgRepo, *mockEventPublisher) {
	userRepo := newMockUserRepo()
	orgRepo := newMockOrgRepo()
	tokenSvc := &mockTokenService{}
	hasher := &mockPasswordHasher{}
	publisher := &mockEventPublisher{}

	svc := services.NewAuthService(userRepo, orgRepo, tokenSvc, hasher, publisher).(*services.AuthService)
	return svc, userRepo, orgRepo, publisher
}

func TestAuthService_Register_B2C_Success(t *testing.T) {
	svc, userRepo, _, publisher := setupAuthService()
	ctx := context.Background()

	cmd := ports.RegisterUserCommand{
		TenantID: "tenant-001",
		Email:    "customer@gmail.com",
		Phone:    "081234567890",
		Password: "SecurePassword123!",
		UserType: domain.UserTypeB2CRetail,
	}

	user, tokens, err := svc.Register(ctx, cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if user.ID == "" {
		t.Errorf("expected non-empty user ID")
	}
	if user.Email != "customer@gmail.com" {
		t.Errorf("expected email customer@gmail.com, got %s", user.Email)
	}
	if user.Role != "CUSTOMER" {
		t.Errorf("expected role CUSTOMER, got %s", user.Role)
	}
	if tokens.AccessToken != "mock-access-token" {
		t.Errorf("expected valid access token")
	}

	// Verify persistence in repository
	saved, err := userRepo.FindByID(ctx, user.ID)
	if err != nil || saved == nil {
		t.Errorf("expected user to be saved in repository")
	}

	// Verify event published
	if len(publisher.publishedUsers) != 1 {
		t.Errorf("expected 1 published event, got %d", len(publisher.publishedUsers))
	}
}

func TestAuthService_Register_B2B_Success(t *testing.T) {
	svc, userRepo, orgRepo, publisher := setupAuthService()
	ctx := context.Background()

	cmd := ports.RegisterUserCommand{
		TenantID:         "tenant-001",
		Email:            "kontraktor@adhimix.co.id",
		Phone:            "081987654321",
		Password:         "SecurePassword123!",
		UserType:         domain.UserTypeB2BContractor,
		OrganizationName: "PT Adhimix Konstruksi Indonesia",
		NPWP:             "01.234.567.8-901.000",
	}

	user, tokens, err := svc.Register(ctx, cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if user.OrganizationID == nil {
		t.Fatalf("expected organization_id to be populated for B2B user")
	}
	if user.Role != "CONTRACTOR_ADMIN" {
		t.Errorf("expected role CONTRACTOR_ADMIN, got %s", user.Role)
	}
	if tokens.AccessToken == "" {
		t.Errorf("expected non-empty tokens")
	}

	// Verify organization saved
	org, err := orgRepo.FindByID(ctx, *user.OrganizationID)
	if err != nil || org == nil {
		t.Errorf("expected organization to be saved")
	}
	if org.Name != "PT Adhimix Konstruksi Indonesia" {
		t.Errorf("expected org name match, got %s", org.Name)
	}

	// Verify user saved
	saved, _ := userRepo.FindByID(ctx, user.ID)
	if saved == nil {
		t.Errorf("expected user saved")
	}
	if len(publisher.publishedUsers) != 1 {
		t.Errorf("expected 1 event published")
	}
}

func TestAuthService_Register_PasswordTooShort(t *testing.T) {
	svc, _, _, _ := setupAuthService()
	ctx := context.Background()

	cmd := ports.RegisterUserCommand{
		TenantID: "tenant-001",
		Email:    "short@gmail.com",
		Password: "123",
	}

	_, _, err := svc.Register(ctx, cmd)
	if !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Fatalf("expected ErrPasswordTooShort, got: %v", err)
	}
}

func TestAuthService_Register_AlreadyExists(t *testing.T) {
	svc, userRepo, _, _ := setupAuthService()
	ctx := context.Background()

	existingUser := &domain.User{
		ID:           "existing-1",
		TenantID:     "tenant-001",
		Email:        "duplicate@gmail.com",
		Phone:        "0811111111",
		PasswordHash: "hashed-pass",
		UserType:     domain.UserTypeB2CRetail,
		Role:         "CUSTOMER",
		IsActive:     true,
	}
	_ = userRepo.Create(ctx, existingUser)

	cmd := ports.RegisterUserCommand{
		TenantID: "tenant-001",
		Email:    "duplicate@gmail.com",
		Password: "SecurePassword123!",
	}

	_, _, err := svc.Register(ctx, cmd)
	if !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got: %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	svc, userRepo, _, _ := setupAuthService()
	ctx := context.Background()

	u := &domain.User{
		ID:           "user-001",
		TenantID:     "tenant-001",
		Email:        "login.test@gmail.com",
		Phone:        "0812344321",
		PasswordHash: "hashed-Password123!",
		UserType:     domain.UserTypeB2CRetail,
		Role:         "CUSTOMER",
		IsActive:     true,
	}
	_ = userRepo.Create(ctx, u)

	// Login via Email
	user, tokens, err := svc.Login(ctx, ports.LoginCommand{
		TenantID: "tenant-001",
		Login:    "login.test@gmail.com",
		Password: "Password123!",
	})
	if err != nil {
		t.Fatalf("expected login success via email, got: %v", err)
	}
	if user.ID != "user-001" {
		t.Errorf("expected user ID user-001, got %s", user.ID)
	}
	if tokens.AccessToken != "mock-access-token" {
		t.Errorf("expected valid tokens")
	}

	// Login via Phone
	userPhone, _, err := svc.Login(ctx, ports.LoginCommand{
		TenantID: "tenant-001",
		Login:    "0812344321",
		Password: "Password123!",
	})
	if err != nil {
		t.Fatalf("expected login success via phone, got: %v", err)
	}
	if userPhone.ID != "user-001" {
		t.Errorf("expected user ID user-001, got %s", userPhone.ID)
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	svc, userRepo, _, _ := setupAuthService()
	ctx := context.Background()

	u := &domain.User{
		ID:           "user-002",
		TenantID:     "tenant-001",
		Email:        "wrongpass@gmail.com",
		PasswordHash: "hashed-CorrectPass123!",
		IsActive:     true,
	}
	_ = userRepo.Create(ctx, u)

	// Wrong password
	_, _, err := svc.Login(ctx, ports.LoginCommand{
		TenantID: "tenant-001",
		Login:    "wrongpass@gmail.com",
		Password: "IncorrectPassword!",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials on wrong password, got: %v", err)
	}

	// Unknown email
	_, _, err = svc.Login(ctx, ports.LoginCommand{
		TenantID: "tenant-001",
		Login:    "unknown@gmail.com",
		Password: "AnyPassword123!",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials on unknown user, got: %v", err)
	}
}

func TestAuthService_Login_UserInactive(t *testing.T) {
	svc, userRepo, _, _ := setupAuthService()
	ctx := context.Background()

	u := &domain.User{
		ID:           "user-inactive",
		TenantID:     "tenant-001",
		Email:        "inactive@gmail.com",
		PasswordHash: "hashed-Password123!",
		IsActive:     false,
	}
	_ = userRepo.Create(ctx, u)

	_, _, err := svc.Login(ctx, ports.LoginCommand{
		TenantID: "tenant-001",
		Login:    "inactive@gmail.com",
		Password: "Password123!",
	})
	if !errors.Is(err, domain.ErrUserInactive) {
		t.Fatalf("expected ErrUserInactive, got: %v", err)
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	svc, userRepo, _, _ := setupAuthService()
	ctx := context.Background()

	// Seed user aktif untuk user-123
	_ = userRepo.Create(ctx, &domain.User{
		ID:       "user-123",
		TenantID: "tenant-123",
		Email:    "budi@example.com",
		IsActive: true,
	})

	// Valid refresh token
	tokens, err := svc.RefreshToken(ctx, ports.RefreshTokenCommand{
		RefreshToken: "valid-refresh-token",
	})
	if err != nil {
		t.Fatalf("expected success on valid refresh token, got: %v", err)
	}
	if tokens.AccessToken != "mock-access-token" {
		t.Errorf("expected valid access token")
	}

	// Invalid refresh token
	_, err = svc.RefreshToken(ctx, ports.RefreshTokenCommand{
		RefreshToken: "invalid-expired-token",
	})
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got: %v", err)
	}
}

func TestAuthService_GetUserProfile(t *testing.T) {
	svc, userRepo, orgRepo, _ := setupAuthService()
	ctx := context.Background()

	orgID := "org-999"
	org := &domain.Organization{
		ID:       orgID,
		TenantID: "tenant-001",
		Name:     "PT Bangun Persada",
	}
	_ = orgRepo.Create(ctx, org)

	u := &domain.User{
		ID:             "user-profile-test",
		TenantID:       "tenant-001",
		OrganizationID: &orgID,
		Email:          "mandor@bangunpersada.co.id",
		UserType:       domain.UserTypeB2BContractor,
		Role:           "CONTRACTOR_ADMIN",
		IsActive:       true,
	}
	_ = userRepo.Create(ctx, u)

	profileUser, profileOrg, err := svc.GetUserProfile(ctx, "user-profile-test")
	if err != nil {
		t.Fatalf("expected no error fetching profile, got: %v", err)
	}
	if profileUser.Email != u.Email {
		t.Errorf("expected email match, got %s", profileUser.Email)
	}
	if profileOrg == nil || profileOrg.Name != "PT Bangun Persada" {
		t.Errorf("expected organization match, got %v", profileOrg)
	}
}
