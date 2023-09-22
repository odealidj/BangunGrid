package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"bangungrid/services/auth-service/internal/core/domain"
	"bangungrid/services/auth-service/internal/core/ports"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) ports.UserRepositoryPort {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, tenant_id, organization_id, email, phone, password_hash, user_type, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		u.ID, u.TenantID, u.OrganizationID, u.Email, u.Phone, u.PasswordHash,
		string(u.UserType), u.Role, u.IsActive, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, organization_id, email, phone, password_hash, user_type, role, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`
	var u domain.User
	var userTypeStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.TenantID, &u.OrganizationID, &u.Email, &u.Phone, &u.PasswordHash,
		&userTypeStr, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	u.UserType = domain.UserType(userTypeStr)
	return &u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, tenantID, email string) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, organization_id, email, phone, password_hash, user_type, role, is_active, created_at, updated_at
		FROM users WHERE tenant_id = $1 AND email = $2
	`
	var u domain.User
	var userTypeStr string
	err := r.pool.QueryRow(ctx, query, tenantID, email).Scan(
		&u.ID, &u.TenantID, &u.OrganizationID, &u.Email, &u.Phone, &u.PasswordHash,
		&userTypeStr, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	u.UserType = domain.UserType(userTypeStr)
	return &u, nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, tenantID, phone string) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, organization_id, email, phone, password_hash, user_type, role, is_active, created_at, updated_at
		FROM users WHERE tenant_id = $1 AND phone = $2
	`
	var u domain.User
	var userTypeStr string
	err := r.pool.QueryRow(ctx, query, tenantID, phone).Scan(
		&u.ID, &u.TenantID, &u.OrganizationID, &u.Email, &u.Phone, &u.PasswordHash,
		&userTypeStr, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	u.UserType = domain.UserType(userTypeStr)
	return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	query := `
		UPDATE users SET role = $1, is_active = $2, updated_at = $3
		WHERE id = $4
	`
	_, err := r.pool.Exec(ctx, query, u.Role, u.IsActive, u.UpdatedAt, u.ID)
	return err
}
