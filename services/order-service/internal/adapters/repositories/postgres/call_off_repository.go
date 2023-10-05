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

type PostgresCallOffRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresCallOffRepository(pool *pgxpool.Pool) *PostgresCallOffRepository {
	return &PostgresCallOffRepository{pool: pool}
}

func (r *PostgresCallOffRepository) Save(ctx context.Context, callOff *domain.B2BCallOffDelivery) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	query := `
		INSERT INTO b2b_call_off_deliveries (
			id, order_id, sequence_no, requested_delivery_date, status, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = tx.Exec(ctx, query,
		callOff.ID, callOff.OrderID, callOff.SequenceNo, callOff.RequestedDeliveryDate,
		string(callOff.Status), callOff.Notes, callOff.CreatedAt, callOff.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal insert b2b_call_off_delivery: %w", err)
	}

	queryItem := `
		INSERT INTO b2b_call_off_items (
			id, call_off_id, sku_id, product_name, quantity, uom
		) VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, it := range callOff.Items {
		_, err = tx.Exec(ctx, queryItem,
			it.ID, callOff.ID, it.SKUID, it.ProductName, it.Quantity, it.UOM,
		)
		if err != nil {
			return fmt.Errorf("gagal insert b2b_call_off_item: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresCallOffRepository) FindByID(ctx context.Context, callOffID uuid.UUID) (*domain.B2BCallOffDelivery, error) {
	query := `
		SELECT id, order_id, sequence_no, requested_delivery_date, status, COALESCE(notes, ''), created_at, updated_at
		FROM b2b_call_off_deliveries
		WHERE id = $1
	`
	var c domain.B2BCallOffDelivery
	var status string
	err := r.pool.QueryRow(ctx, query, callOffID).Scan(
		&c.ID, &c.OrderID, &c.SequenceNo, &c.RequestedDeliveryDate,
		&status, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCallOffNotFound
		}
		return nil, err
	}
	c.Status = domain.CallOffStatus(status)

	// Fetch items
	itemQuery := `
		SELECT id, call_off_id, sku_id, product_name, quantity, uom
		FROM b2b_call_off_items
		WHERE call_off_id = $1
		ORDER BY id ASC
	`
	rows, err := r.pool.Query(ctx, itemQuery, c.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var it domain.CallOffItem
		if err := rows.Scan(&it.ID, &it.CallOffID, &it.SKUID, &it.ProductName, &it.Quantity, &it.UOM); err != nil {
			return nil, err
		}
		c.Items = append(c.Items, it)
	}

	return &c, nil
}

func (r *PostgresCallOffRepository) ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.B2BCallOffDelivery, error) {
	query := `
		SELECT id, order_id, sequence_no, requested_delivery_date, status, COALESCE(notes, ''), created_at, updated_at
		FROM b2b_call_off_deliveries
		WHERE order_id = $1
		ORDER BY sequence_no ASC
	`
	rows, err := r.pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.B2BCallOffDelivery
	for rows.Next() {
		var c domain.B2BCallOffDelivery
		var status string
		if err := rows.Scan(
			&c.ID, &c.OrderID, &c.SequenceNo, &c.RequestedDeliveryDate,
			&status, &c.Notes, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		c.Status = domain.CallOffStatus(status)
		list = append(list, &c)
	}

	return list, nil
}

func (r *PostgresCallOffRepository) UpdateStatus(ctx context.Context, deliveryID uuid.UUID, status domain.CallOffStatus) error {
	query := `
		UPDATE b2b_call_off_deliveries
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	tag, err := r.pool.Exec(ctx, query, string(status), time.Now().UTC(), deliveryID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCallOffNotFound
	}
	return nil
}
