package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"bangungrid/services/auth-service/internal/core/domain"
	"bangungrid/services/auth-service/internal/core/ports"
)

type OrganizationRepository struct {
	pool *pgxpool.Pool
}

func NewOrganizationRepository(pool *pgxpool.Pool) ports.OrganizationRepositoryPort {
	return &OrganizationRepository{pool: pool}
}

func (r *OrganizationRepository) Create(ctx context.Context, org *domain.Organization) error {
	query := `
		INSERT INTO organizations (id, tenant_id, name, npwp, nib, legal_address, contact_email, contact_phone, verification_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		org.ID, org.TenantID, org.Name, org.NPWP, org.NIB, org.LegalAddress,
		org.ContactEmail, org.ContactPhone, string(org.VerificationStatus), org.CreatedAt, org.UpdatedAt,
	)
	return err
}

func (r *OrganizationRepository) FindByID(ctx context.Context, id string) (*domain.Organization, error) {
	query := `
		SELECT id, tenant_id, name, npwp, nib, legal_address, contact_email, contact_phone, verification_status, created_at, updated_at
		FROM organizations WHERE id = $1
	`
	var org domain.Organization
	var statusStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&org.ID, &org.TenantID, &org.Name, &org.NPWP, &org.NIB, &org.LegalAddress,
		&org.ContactEmail, &org.ContactPhone, &statusStr, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, err
	}
	org.VerificationStatus = domain.VerificationStatus(statusStr)
	return &org, nil
}
