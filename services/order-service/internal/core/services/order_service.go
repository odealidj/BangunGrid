package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"bangungrid/services/order-service/internal/core/domain"
	"bangungrid/services/order-service/internal/core/ports"
)

type OrderService struct {
	orderRepo   ports.OrderRepositoryPort
	rfqRepo     ports.RFQRepositoryPort
	callOffRepo ports.CallOffRepositoryPort
	publisher   ports.OrderEventPublisherPort
}

func NewOrderService(
	orderRepo ports.OrderRepositoryPort,
	rfqRepo ports.RFQRepositoryPort,
	callOffRepo ports.CallOffRepositoryPort,
	publisher ports.OrderEventPublisherPort,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		rfqRepo:     rfqRepo,
		callOffRepo: callOffRepo,
		publisher:   publisher,
	}
}

// CalculateFees menghitung biaya logistik khusus material bangunan (kuli bongkar & langsir)
func CalculateFees(items []ports.CreateOrderItemCommand, includeUnloading, narrowRoadAccess bool) (unloadingFee float64, narrowRoadFee float64, totalWeight float64, totalVolume float64, subtotal float64) {
	for _, it := range items {
		itemSubtotal := it.Quantity * it.UnitPriceSnapshot
		subtotal += itemSubtotal

		itemWeight := it.Quantity * it.WeightKg
		totalWeight += itemWeight

		itemVolume := it.Quantity * it.VolumeM3
		totalVolume += itemVolume

		if includeUnloading {
			// Perhitungan ongkos kuli bongkar material berat:
			// Semen sak: Rp 1.500 / sak; Palet: Rp 25.000 / palet; Curah / tonase: Rp 35.000 / ton (Rp 35/kg)
			switch it.UOMCodeSnapshot {
			case "SAK":
				unloadingFee += it.Quantity * 1500
			case "PALET":
				unloadingFee += it.Quantity * 25000
			default:
				unloadingFee += itemWeight * 35 // Rp 35 / Kg = Rp 35.000 / Ton
			}
		}
	}

	if includeUnloading && unloadingFee < 50000 && unloadingFee > 0 {
		unloadingFee = 50000 // Biaya minimal jasa bongkar tim kuli
	}

	if narrowRoadAccess {
		// Biaya langsir armada kecil (Pick-up bak 1.5 ton) jika jalan sempit
		ritaseCount := math.Ceil(totalWeight / 1500.0)
		if ritaseCount < 1 {
			ritaseCount = 1
		}
		narrowRoadFee = ritaseCount * 150000 // Rp 150.000 per ritase langsir
	}

	return unloadingFee, narrowRoadFee, totalWeight, totalVolume, subtotal
}

func DetermineRecommendedFleet(totalWeightKg, totalVolumeM3 float64) string {
	if totalWeightKg <= 1500 && totalVolumeM3 <= 1.5 {
		return "PICKUP"
	} else if totalWeightKg <= 5000 && totalVolumeM3 <= 4.0 {
		return "ENGKEL_4W"
	} else if totalWeightKg <= 12000 && totalVolumeM3 <= 8.0 {
		return "DUMP_6W"
	}
	return "TRONTON_10W"
}

func (s *OrderService) EstimateOrder(ctx context.Context, cmd ports.EstimateOrderCommand) (*ports.OrderEstimate, error) {
	if len(cmd.Items) == 0 {
		return nil, domain.ErrEmptyOrderItems
	}

	unloadingFee, narrowFee, totalWeight, totalVolume, subtotal := CalculateFees(cmd.Items, cmd.IncludeUnloading, cmd.NarrowRoadAccess)
	fleet := DetermineRecommendedFleet(totalWeight, totalVolume)

	return &ports.OrderEstimate{
		TotalWeightKg:          totalWeight,
		TotalVolumeM3:          totalVolume,
		EstimatedItemsSubtotal: subtotal,
		RecommendedFleet:       fleet,
		EstimatedUnloadingFee:  unloadingFee,
		EstimatedNarrowRoadFee: narrowFee,
		EstimatedGrandTotal:    subtotal + unloadingFee + narrowFee,
	}, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, cmd ports.CreateOrderCommand) (*domain.Order, error) {
	if len(cmd.Items) == 0 {
		return nil, domain.ErrEmptyOrderItems
	}
	if cmd.TenantID == uuid.Nil || cmd.CustomerID == uuid.Nil {
		return nil, domain.ErrInvalidOrderData
	}

	seq, err := s.orderRepo.GetNextSequence(ctx, cmd.TenantID)
	if err != nil {
		seq = time.Now().Unix() % 10000
	}
	orderNo := domain.GenerateOrderNo(time.Now(), seq)

	unloadingFee, narrowRoadFee, _, _, _ := CalculateFees(cmd.Items, cmd.IncludeUnloading, cmd.NarrowRoadAccess)

	order := domain.NewOrder(
		cmd.TenantID,
		cmd.BranchID,
		cmd.CustomerID,
		cmd.CustomerType,
		orderNo,
		cmd.DeliveryAddress,
		cmd.DeliveryLatitude,
		cmd.DeliveryLongitude,
		cmd.ShippingFee,
		unloadingFee,
		narrowRoadFee,
		cmd.Notes,
	)

	for _, it := range cmd.Items {
		item, err := domain.NewOrderItem(
			it.SKUID,
			it.ProductNameSnapshot,
			it.UOMCodeSnapshot,
			it.Quantity,
			it.UnitPriceSnapshot,
			it.WeightKg,
			it.VolumeM3,
		)
		if err != nil {
			return nil, err
		}
		order.AddItem(*item)
	}

	if err := s.orderRepo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("gagal menyimpan pesanan: %w", err)
	}

	if s.publisher != nil {
		_ = s.publisher.PublishOrderCreated(ctx, order)
	}

	return order, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, tenantID, orderID uuid.UUID) (*domain.Order, error) {
	return s.orderRepo.FindByID(ctx, tenantID, orderID)
}

func (s *OrderService) GetOrderByOrderNo(ctx context.Context, tenantID uuid.UUID, orderNo string) (*domain.Order, error) {
	return s.orderRepo.FindByOrderNo(ctx, tenantID, orderNo)
}

func (s *OrderService) ListOrders(ctx context.Context, filter ports.ListOrdersFilter) ([]*domain.Order, int64, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	return s.orderRepo.List(ctx, filter)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, tenantID, orderID uuid.UUID, status domain.OrderStatus, reason string) (*domain.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, tenantID, orderID)
	if err != nil {
		return nil, err
	}

	if err := order.UpdateStatus(status); err != nil {
		return nil, err
	}

	if err := s.orderRepo.UpdateStatus(ctx, tenantID, orderID, status, reason); err != nil {
		return nil, err
	}

	if s.publisher != nil {
		_ = s.publisher.PublishOrderStatusChanged(ctx, order)
	}

	return order, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, tenantID, orderID uuid.UUID, reason string) error {
	order, err := s.orderRepo.FindByID(ctx, tenantID, orderID)
	if err != nil {
		return err
	}

	if err := order.Cancel(reason); err != nil {
		return err
	}

	if err := s.orderRepo.UpdateStatus(ctx, tenantID, orderID, domain.OrderStatusCancelled, reason); err != nil {
		return err
	}

	if s.publisher != nil {
		_ = s.publisher.PublishOrderCancelled(ctx, order)
	}

	return nil
}

func (s *OrderService) CreateRFQ(ctx context.Context, cmd ports.CreateRFQCommand) (*domain.B2BRFQ, error) {
	if cmd.TenantID == uuid.Nil || cmd.CustomerID == uuid.Nil || cmd.ProjectName == "" {
		return nil, domain.ErrInvalidOrderData
	}

	rfq := domain.NewB2BRFQ(
		cmd.TenantID,
		cmd.CustomerID,
		cmd.ProjectName,
		cmd.ProjectAddress,
		cmd.BoQDocumentURL,
		cmd.EstimatedBudget,
		cmd.Notes,
	)

	for _, it := range cmd.Items {
		rfq.AddItem(it.MaterialName, it.TargetUOM, it.EstimatedQty, it.TargetPrice, it.Notes)
	}

	if err := s.rfqRepo.Save(ctx, rfq); err != nil {
		return nil, fmt.Errorf("gagal menyimpan RFQ: %w", err)
	}

	return rfq, nil
}

func (s *OrderService) ListRFQs(ctx context.Context, tenantID uuid.UUID, customerID *uuid.UUID) ([]*domain.B2BRFQ, error) {
	return s.rfqRepo.List(ctx, tenantID, customerID)
}

func (s *OrderService) GetRFQByID(ctx context.Context, tenantID, rfqID uuid.UUID) (*domain.B2BRFQ, error) {
	return s.rfqRepo.FindByID(ctx, tenantID, rfqID)
}

func (s *OrderService) UpdateRFQStatus(ctx context.Context, tenantID, rfqID uuid.UUID, status domain.RFQStatus, quotedAmount float64, quotationNotes string) (*domain.B2BRFQ, error) {
	rfq, err := s.rfqRepo.FindByID(ctx, tenantID, rfqID)
	if err != nil {
		return nil, err
	}

	rfq.Status = status
	if status == domain.RFQStatusQuoted {
		rfq.UpdateQuote(quotedAmount, quotationNotes)
	}

	if err := s.rfqRepo.UpdateQuote(ctx, tenantID, rfqID, status, quotedAmount, quotationNotes); err != nil {
		return nil, err
	}

	return rfq, nil
}

func (s *OrderService) CreateCallOffDelivery(ctx context.Context, cmd ports.CreateCallOffCommand) (*domain.B2BCallOffDelivery, error) {
	if cmd.OrderID == uuid.Nil {
		return nil, domain.ErrInvalidOrderData
	}

	callOff := domain.NewCallOffDelivery(
		cmd.OrderID,
		cmd.SequenceNo,
		cmd.RequestedDeliveryDate,
		cmd.Notes,
	)

	for _, it := range cmd.Items {
		callOff.AddItem(it.SKUID, it.ProductName, it.Quantity, it.UOM)
	}

	if err := s.callOffRepo.Save(ctx, callOff); err != nil {
		return nil, fmt.Errorf("gagal menjadwalkan pengiriman bertahap: %w", err)
	}

	return callOff, nil
}

func (s *OrderService) ListCallOffDeliveries(ctx context.Context, orderID uuid.UUID) ([]*domain.B2BCallOffDelivery, error) {
	return s.callOffRepo.ListByOrderID(ctx, orderID)
}

func (s *OrderService) UpdateCallOffStatus(ctx context.Context, deliveryID uuid.UUID, status domain.CallOffStatus) error {
	return s.callOffRepo.UpdateStatus(ctx, deliveryID, status)
}
