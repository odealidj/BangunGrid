package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"bangungrid/services/order-service/internal/core/domain"
)

type PostgresRFQRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRFQRepository(pool *pgxpool.Pool) *PostgresRFQRepository {
	return &PostgresRFQRepository{pool: pool}
}

func (r *PostgresRFQRepository) Save(ctx context.Context, rfq *domain.B2BRFQ) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	queryRFQ := `
		INSERT INTO b2b_rfqs (
			id, tenant_id, customer_id, project_name, project_address,
			boq_document_url, estimated_budget, status, notes,
			quoted_amount, quotation_notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = tx.Exec(ctx, queryRFQ,
		rfq.ID, rfq.TenantID, rfq.CustomerID, rfq.ProjectName, rfq.ProjectAddress,
		rfq.BoQDocumentURL, rfq.EstimatedBudget, string(rfq.Status), rfq.Notes,
		rfq.QuotedAmount, rfq.QuotationNotes, rfq.CreatedAt, rfq.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal insert b2b_rfq: %w", err)
	}

	queryItem := `
		INSERT INTO b2b_rfq_items (
			id, rfq_id, material_name, target_uom, estimated_qty, target_price, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for _, it := range rfq.Items {
		_, err = tx.Exec(ctx, queryItem,
			it.ID, rfq.ID, it.MaterialName, it.TargetUOM, it.EstimatedQty, it.TargetPrice, it.Notes,
		)
		if err != nil {
			return fmt.Errorf("gagal insert b2b_rfq_item: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresRFQRepository) FindByID(ctx context.Context, tenantID, rfqID uuid.UUID) (*domain.B2BRFQ, error) {
	query := `
		SELECT
			id, tenant_id, customer_id, project_name, project_address,
			COALESCE(boq_document_url, ''), estimated_budget, status, COALESCE(notes, ''),
			quoted_amount, COALESCE(quotation_notes, ''), created_at, updated_at
		FROM b2b_rfqs
		WHERE id = $1 AND tenant_id = $2
	`

	var rfq domain.B2BRFQ
	var status string
	err := r.pool.QueryRow(ctx, query, rfqID, tenantID).Scan(
		&rfq.ID, &rfq.TenantID, &rfq.CustomerID, &rfq.ProjectName, &rfq.ProjectAddress,
		&rfq.BoQDocumentURL, &rfq.EstimatedBudget, &status, &rfq.Notes,
		&rfq.QuotedAmount, &rfq.QuotationNotes, &rfq.CreatedAt, &rfq.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRFQNotFound
		}
		return nil, err
	}
	rfq.Status = domain.RFQStatus(status)

	// Fetch items
	itemQuery := `
		SELECT id, rfq_id, material_name, target_uom, estimated_qty, target_price, COALESCE(notes, '')
		FROM b2b_rfq_items
		WHERE rfq_id = $1
		ORDER BY id ASC
	`
	rows, err := r.pool.Query(ctx, itemQuery, rfq.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var it domain.RFQItem
		if err := rows.Scan(
			&it.ID, &it.RFQID, &it.MaterialName, &it.TargetUOM, &it.EstimatedQty, &it.TargetPrice, &it.Notes,
		); err != nil {
			return nil, err
		}
		rfq.Items = append(rfq.Items, it)
	}

	return &rfq, nil
}

func (r *PostgresRFQRepository) List(ctx context.Context, tenantID uuid.UUID, customerID *uuid.UUID) ([]*domain.B2BRFQ, error) {
	query := `
		SELECT
			id, tenant_id, customer_id, project_name, project_address,
			COALESCE(boq_document_url, ''), estimated_budget, status, COALESCE(notes, ''),
			quoted_amount, COALESCE(quotation_notes, ''), created_at, updated_at
		FROM b2b_rfqs
		WHERE tenant_id = $1
	`
	args := []any{tenantID}

	if customerID != nil {
		query += " AND customer_id = $2"
		args = append(args, *customerID)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.B2BRFQ
	for rows.Next() {
		var rfq domain.B2BRFQ
		var status string
		if err := rows.Scan(
			&rfq.ID, &rfq.TenantID, &rfq.CustomerID, &rfq.ProjectName, &rfq.ProjectAddress,
			&rfq.BoQDocumentURL, &rfq.EstimatedBudget, &status, &rfq.Notes,
			&rfq.QuotedAmount, &rfq.QuotationNotes, &rfq.CreatedAt, &rfq.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rfq.Status = domain.RFQStatus(status)
		list = append(list, &rfq)
	}

	return list, nil
}

func (r *PostgresRFQRepository) UpdateQuote(ctx context.Context, tenantID, rfqID uuid.UUID, status domain.RFQStatus, amount float64, notes string) error {
	query := `
		UPDATE b2b_rfqs
		SET status = $1, quoted_amount = $2, quotation_notes = $3, updated_at = $4
		WHERE id = $5 AND tenant_id = $6
	`
	tag, err := r.pool.Exec(ctx, query, string(status), amount, notes, time.Now().UTC(), rfqID, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrRFQNotFound
	}
	return nil
}
