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
	"bangungrid/services/order-service/internal/core/ports"
)

type PostgresOrderRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOrderRepository(pool *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{pool: pool}
}

func (r *PostgresOrderRepository) GetNextSequence(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var seq int64
	err := r.pool.QueryRow(ctx, "SELECT nextval('order_no_seq')").Scan(&seq)
	if err != nil {
		return 0, err
	}
	return seq, nil
}

func (r *PostgresOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	queryOrder := `
		INSERT INTO orders (
			id, tenant_id, branch_id, order_no, customer_id, customer_type,
			status, total_amount, total_weight_kg, total_volume_m3,
			unloading_fee, narrow_road_fee, shipping_fee, grand_total,
			delivery_address, delivery_latitude, delivery_longitude, notes,
			cancel_reason, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13, $14,
			$15, $16, $17, $18,
			$19, $20, $21
		)
		ON CONFLICT (tenant_id, order_no) DO UPDATE SET
			status = EXCLUDED.status,
			total_amount = EXCLUDED.total_amount,
			total_weight_kg = EXCLUDED.total_weight_kg,
			total_volume_m3 = EXCLUDED.total_volume_m3,
			unloading_fee = EXCLUDED.unloading_fee,
			narrow_road_fee = EXCLUDED.narrow_road_fee,
			shipping_fee = EXCLUDED.shipping_fee,
			grand_total = EXCLUDED.grand_total,
			cancel_reason = EXCLUDED.cancel_reason,
			updated_at = EXCLUDED.updated_at;
	`

	_, err = tx.Exec(ctx, queryOrder,
		order.ID, order.TenantID, order.BranchID, order.OrderNo, order.CustomerID, string(order.CustomerType),
		string(order.Status), order.TotalAmount, order.TotalWeightKg, order.TotalVolumeM3,
		order.UnloadingFee, order.NarrowRoadFee, order.ShippingFee, order.GrandTotal,
		order.DeliveryAddress, order.DeliveryLatitude, order.DeliveryLongitude, order.Notes,
		order.CancelReason, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal insert order: %w", err)
	}

	// Insert items
	queryItem := `
		INSERT INTO order_items (
			id, order_id, sku_id, product_name_snapshot, uom_code_snapshot,
			quantity, unit_price_snapshot, subtotal, weight_kg_snapshot, volume_m3_snapshot
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	for _, it := range order.Items {
		_, err = tx.Exec(ctx, queryItem,
			it.ID, order.ID, it.SKUID, it.ProductNameSnapshot, it.UOMCodeSnapshot,
			it.Quantity, it.UnitPriceSnapshot, it.Subtotal, it.WeightKgSnapshot, it.VolumeM3Snapshot,
		)
		if err != nil {
			return fmt.Errorf("gagal insert order_item: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresOrderRepository) FindByID(ctx context.Context, tenantID, orderID uuid.UUID) (*domain.Order, error) {
	query := `
		SELECT
			id, tenant_id, branch_id, order_no, customer_id, customer_type,
			status, total_amount, total_weight_kg, total_volume_m3,
			unloading_fee, narrow_road_fee, shipping_fee, grand_total,
			delivery_address, delivery_latitude, delivery_longitude, notes,
			COALESCE(cancel_reason, ''), created_at, updated_at
		FROM orders
		WHERE id = $1 AND tenant_id = $2
	`

	var o domain.Order
	var custType, status string
	err := r.pool.QueryRow(ctx, query, orderID, tenantID).Scan(
		&o.ID, &o.TenantID, &o.BranchID, &o.OrderNo, &o.CustomerID, &custType,
		&status, &o.TotalAmount, &o.TotalWeightKg, &o.TotalVolumeM3,
		&o.UnloadingFee, &o.NarrowRoadFee, &o.ShippingFee, &o.GrandTotal,
		&o.DeliveryAddress, &o.DeliveryLatitude, &o.DeliveryLongitude, &o.Notes,
		&o.CancelReason, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	o.CustomerType = domain.OrderType(custType)
	o.Status = domain.OrderStatus(status)

	// Fetch items
	items, err := r.fetchItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

func (r *PostgresOrderRepository) FindByOrderNo(ctx context.Context, tenantID uuid.UUID, orderNo string) (*domain.Order, error) {
	query := `
		SELECT
			id, tenant_id, branch_id, order_no, customer_id, customer_type,
			status, total_amount, total_weight_kg, total_volume_m3,
			unloading_fee, narrow_road_fee, shipping_fee, grand_total,
			delivery_address, delivery_latitude, delivery_longitude, notes,
			COALESCE(cancel_reason, ''), created_at, updated_at
		FROM orders
		WHERE order_no = $1 AND tenant_id = $2
	`

	var o domain.Order
	var custType, status string
	err := r.pool.QueryRow(ctx, query, orderNo, tenantID).Scan(
		&o.ID, &o.TenantID, &o.BranchID, &o.OrderNo, &o.CustomerID, &custType,
		&status, &o.TotalAmount, &o.TotalWeightKg, &o.TotalVolumeM3,
		&o.UnloadingFee, &o.NarrowRoadFee, &o.ShippingFee, &o.GrandTotal,
		&o.DeliveryAddress, &o.DeliveryLatitude, &o.DeliveryLongitude, &o.Notes,
		&o.CancelReason, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	o.CustomerType = domain.OrderType(custType)
	o.Status = domain.OrderStatus(status)

	items, err := r.fetchItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

func (r *PostgresOrderRepository) fetchItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	query := `
		SELECT
			id, order_id, sku_id, product_name_snapshot, uom_code_snapshot,
			quantity, unit_price_snapshot, subtotal, weight_kg_snapshot, volume_m3_snapshot
		FROM order_items
		WHERE order_id = $1
		ORDER BY id ASC
	`
	rows, err := r.pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var it domain.OrderItem
		if err := rows.Scan(
			&it.ID, &it.OrderID, &it.SKUID, &it.ProductNameSnapshot, &it.UOMCodeSnapshot,
			&it.Quantity, &it.UnitPriceSnapshot, &it.Subtotal, &it.WeightKgSnapshot, &it.VolumeM3Snapshot,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (r *PostgresOrderRepository) List(ctx context.Context, filter ports.ListOrdersFilter) ([]*domain.Order, int64, error) {
	baseQuery := `
		FROM orders
		WHERE tenant_id = $1
	`
	args := []any{filter.TenantID}
	idx := 2

	if filter.CustomerID != nil {
		baseQuery += fmt.Sprintf(" AND customer_id = $%d", idx)
		args = append(args, *filter.CustomerID)
		idx++
	}

	if filter.Status != nil {
		baseQuery += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, string(*filter.Status))
		idx++
	}

	var total int64
	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQuery := `
		SELECT
			id, tenant_id, branch_id, order_no, customer_id, customer_type,
			status, total_amount, total_weight_kg, total_volume_m3,
			unloading_fee, narrow_road_fee, shipping_fee, grand_total,
			delivery_address, delivery_latitude, delivery_longitude, notes,
			COALESCE(cancel_reason, ''), created_at, updated_at
	` + baseQuery + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		var custType, status string
		if err := rows.Scan(
			&o.ID, &o.TenantID, &o.BranchID, &o.OrderNo, &o.CustomerID, &custType,
			&status, &o.TotalAmount, &o.TotalWeightKg, &o.TotalVolumeM3,
			&o.UnloadingFee, &o.NarrowRoadFee, &o.ShippingFee, &o.GrandTotal,
			&o.DeliveryAddress, &o.DeliveryLatitude, &o.DeliveryLongitude, &o.Notes,
			&o.CancelReason, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		o.CustomerType = domain.OrderType(custType)
		o.Status = domain.OrderStatus(status)
		orders = append(orders, &o)
	}

	return orders, total, nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, tenantID, orderID uuid.UUID, status domain.OrderStatus, reason string) error {
	query := `
		UPDATE orders
		SET status = $1, cancel_reason = CASE WHEN $2 <> '' THEN $2 ELSE cancel_reason END, updated_at = $3
		WHERE id = $4 AND tenant_id = $5
	`
	tag, err := r.pool.Exec(ctx, query, string(status), reason, time.Now().UTC(), orderID, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}
