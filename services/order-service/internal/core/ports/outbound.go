package ports

import (
	"context"

	"github.com/google/uuid"
	"bangungrid/services/order-service/internal/core/domain"
)

type OrderRepositoryPort interface {
	Save(ctx context.Context, order *domain.Order) error
	FindByID(ctx context.Context, tenantID, orderID uuid.UUID) (*domain.Order, error)
	FindByOrderNo(ctx context.Context, tenantID uuid.UUID, orderNo string) (*domain.Order, error)
	List(ctx context.Context, filter ListOrdersFilter) ([]*domain.Order, int64, error)
	UpdateStatus(ctx context.Context, tenantID, orderID uuid.UUID, status domain.OrderStatus, reason string) error
	GetNextSequence(ctx context.Context, tenantID uuid.UUID) (int64, error)
}

type RFQRepositoryPort interface {
	Save(ctx context.Context, rfq *domain.B2BRFQ) error
	FindByID(ctx context.Context, tenantID, rfqID uuid.UUID) (*domain.B2BRFQ, error)
	List(ctx context.Context, tenantID uuid.UUID, customerID *uuid.UUID) ([]*domain.B2BRFQ, error)
	UpdateQuote(ctx context.Context, tenantID, rfqID uuid.UUID, status domain.RFQStatus, quotedAmount float64, quotationNotes string) error
}

type CallOffRepositoryPort interface {
	Save(ctx context.Context, callOff *domain.B2BCallOffDelivery) error
	FindByID(ctx context.Context, callOffID uuid.UUID) (*domain.B2BCallOffDelivery, error)
	ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.B2BCallOffDelivery, error)
	UpdateStatus(ctx context.Context, callOffID uuid.UUID, status domain.CallOffStatus) error
}

type OrderEventPublisherPort interface {
	PublishOrderCreated(ctx context.Context, order *domain.Order) error
	PublishOrderStatusChanged(ctx context.Context, order *domain.Order) error
	PublishOrderCancelled(ctx context.Context, order *domain.Order) error
}
