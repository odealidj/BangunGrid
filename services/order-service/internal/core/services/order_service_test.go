package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"bangungrid/services/order-service/internal/core/domain"
	"bangungrid/services/order-service/internal/core/ports"
	"bangungrid/services/order-service/internal/core/services"
)

// Mock Order Repository
type mockOrderRepo struct {
	orders   map[uuid.UUID]*domain.Order
	sequence int64
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{
		orders:   make(map[uuid.UUID]*domain.Order),
		sequence: 1,
	}
}

func (m *mockOrderRepo) Save(ctx context.Context, order *domain.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *mockOrderRepo) FindByID(ctx context.Context, tenantID, orderID uuid.UUID) (*domain.Order, error) {
	order, exists := m.orders[orderID]
	if !exists || order.TenantID != tenantID {
		return nil, domain.ErrOrderNotFound
	}
	return order, nil
}

func (m *mockOrderRepo) FindByOrderNo(ctx context.Context, tenantID uuid.UUID, orderNo string) (*domain.Order, error) {
	for _, o := range m.orders {
		if o.TenantID == tenantID && o.OrderNo == orderNo {
			return o, nil
		}
	}
	return nil, domain.ErrOrderNotFound
}

func (m *mockOrderRepo) List(ctx context.Context, filter ports.ListOrdersFilter) ([]*domain.Order, int64, error) {
	var result []*domain.Order
	for _, o := range m.orders {
		if o.TenantID == filter.TenantID {
			if filter.CustomerID != nil && o.CustomerID != *filter.CustomerID {
				continue
			}
			if filter.Status != nil && o.Status != *filter.Status {
				continue
			}
			result = append(result, o)
		}
	}
	return result, int64(len(result)), nil
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, tenantID, orderID uuid.UUID, status domain.OrderStatus, reason string) error {
	order, exists := m.orders[orderID]
	if !exists || order.TenantID != tenantID {
		return domain.ErrOrderNotFound
	}
	order.Status = status
	if reason != "" {
		order.CancelReason = reason
	}
	return nil
}

func (m *mockOrderRepo) GetNextSequence(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	m.sequence++
	return m.sequence, nil
}

// Mock RFQ Repository
type mockRFQRepo struct {
	rfqs map[uuid.UUID]*domain.B2BRFQ
}

func newMockRFQRepo() *mockRFQRepo {
	return &mockRFQRepo{rfqs: make(map[uuid.UUID]*domain.B2BRFQ)}
}

func (m *mockRFQRepo) Save(ctx context.Context, rfq *domain.B2BRFQ) error {
	m.rfqs[rfq.ID] = rfq
	return nil
}

func (m *mockRFQRepo) FindByID(ctx context.Context, tenantID, rfqID uuid.UUID) (*domain.B2BRFQ, error) {
	rfq, exists := m.rfqs[rfqID]
	if !exists || rfq.TenantID != tenantID {
		return nil, domain.ErrRFQNotFound
	}
	return rfq, nil
}

func (m *mockRFQRepo) List(ctx context.Context, tenantID uuid.UUID, customerID *uuid.UUID) ([]*domain.B2BRFQ, error) {
	var list []*domain.B2BRFQ
	for _, r := range m.rfqs {
		if r.TenantID == tenantID {
			if customerID != nil && r.CustomerID != *customerID {
				continue
			}
			list = append(list, r)
		}
	}
	return list, nil
}

func (m *mockRFQRepo) UpdateQuote(ctx context.Context, tenantID, rfqID uuid.UUID, status domain.RFQStatus, amount float64, notes string) error {
	rfq, exists := m.rfqs[rfqID]
	if !exists || rfq.TenantID != tenantID {
		return domain.ErrRFQNotFound
	}
	rfq.Status = status
	rfq.QuotedAmount = amount
	rfq.QuotationNotes = notes
	return nil
}

// Mock Call-Off Repository
type mockCallOffRepo struct {
	callOffs map[uuid.UUID]*domain.B2BCallOffDelivery
}

func newMockCallOffRepo() *mockCallOffRepo {
	return &mockCallOffRepo{callOffs: make(map[uuid.UUID]*domain.B2BCallOffDelivery)}
}

func (m *mockCallOffRepo) Save(ctx context.Context, callOff *domain.B2BCallOffDelivery) error {
	m.callOffs[callOff.ID] = callOff
	return nil
}

func (m *mockCallOffRepo) FindByID(ctx context.Context, callOffID uuid.UUID) (*domain.B2BCallOffDelivery, error) {
	c, exists := m.callOffs[callOffID]
	if !exists {
		return nil, domain.ErrCallOffNotFound
	}
	return c, nil
}

func (m *mockCallOffRepo) ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.B2BCallOffDelivery, error) {
	var list []*domain.B2BCallOffDelivery
	for _, c := range m.callOffs {
		if c.OrderID == orderID {
			list = append(list, c)
		}
	}
	return list, nil
}

func (m *mockCallOffRepo) UpdateStatus(ctx context.Context, deliveryID uuid.UUID, status domain.CallOffStatus) error {
	c, exists := m.callOffs[deliveryID]
	if !exists {
		return domain.ErrCallOffNotFound
	}
	c.Status = status
	return nil
}

// Tests
func TestOrderEstimate(t *testing.T) {
	svc := services.NewOrderService(newMockOrderRepo(), newMockRFQRepo(), newMockCallOffRepo(), nil)

	t.Run("Kalkulasi estimasi muatan & armada Pick-up", func(t *testing.T) {
		cmd := ports.EstimateOrderCommand{
			Items: []ports.CreateOrderItemCommand{
				{
					SKUID:               uuid.New(),
					ProductNameSnapshot: "Semen Portland Tiga Roda 50kg",
					UOMCodeSnapshot:     "SAK",
					Quantity:            20, // 20 x 50kg = 1.000 kg (1 Ton)
					UnitPriceSnapshot:   68000,
					WeightKg:            50,
					VolumeM3:            0.035,
				},
			},
			IncludeUnloading: true,
			NarrowRoadAccess: false,
		}

		est, err := svc.EstimateOrder(context.Background(), cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if est.TotalWeightKg != 1000 {
			t.Errorf("expected weight 1000kg, got %f", est.TotalWeightKg)
		}
		if est.RecommendedFleet != "PICKUP" {
			t.Errorf("expected fleet PICKUP, got %s", est.RecommendedFleet)
		}
		if est.EstimatedUnloadingFee != 50000 { // 20 sak x 1500 = 30000 -> min 50000
			t.Errorf("expected unloading fee 50000, got %f", est.EstimatedUnloadingFee)
		}
	})

	t.Run("Kalkulasi muatan besar (Dump Truck 6 Roda) + Akses Jalan Sempit", func(t *testing.T) {
		cmd := ports.EstimateOrderCommand{
			Items: []ports.CreateOrderItemCommand{
				{
					SKUID:               uuid.New(),
					ProductNameSnapshot: "Pasir Beton Hitam Lumajang (Curah)",
					UOMCodeSnapshot:     "M3",
					Quantity:            6, // 6 m3 x 1400 kg = 8.400 kg (8.4 Ton)
					UnitPriceSnapshot:   350000,
					WeightKg:            1400,
					VolumeM3:            1.0,
				},
			},
			IncludeUnloading: false,
			NarrowRoadAccess: true,
		}

		est, err := svc.EstimateOrder(context.Background(), cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if est.TotalWeightKg != 8400 {
			t.Errorf("expected weight 8400kg, got %f", est.TotalWeightKg)
		}
		if est.RecommendedFleet != "DUMP_6W" {
			t.Errorf("expected fleet DUMP_6W, got %s", est.RecommendedFleet)
		}
		// 8400 / 1500 = ceil(5.6) = 6 ritase x 150.000 = 900.000
		if est.EstimatedNarrowRoadFee != 900000 {
			t.Errorf("expected narrow road fee 900000, got %f", est.EstimatedNarrowRoadFee)
		}
	})

	t.Run("Gagal kalkulasi jika item kosong", func(t *testing.T) {
		_, err := svc.EstimateOrder(context.Background(), ports.EstimateOrderCommand{Items: nil})
		if err != domain.ErrEmptyOrderItems {
			t.Errorf("expected ErrEmptyOrderItems, got %v", err)
		}
	})
}

func TestCreateAndManageOrder(t *testing.T) {
	orderRepo := newMockOrderRepo()
	svc := services.NewOrderService(orderRepo, newMockRFQRepo(), newMockCallOffRepo(), nil)

	tenantID := uuid.New()
	branchID := uuid.New()
	customerID := uuid.New()

	var createdOrderID uuid.UUID

	t.Run("Berhasil membuat order B2C dengan validasi total", func(t *testing.T) {
		cmd := ports.CreateOrderCommand{
			TenantID:          tenantID,
			BranchID:          branchID,
			CustomerID:        customerID,
			CustomerType:      domain.CustomerTypeB2CRetail,
			DeliveryAddress:   "Jl. Konstruksi Raya No. 12, Surabaya",
			DeliveryLatitude:  -7.2575,
			DeliveryLongitude: 112.7521,
			IncludeUnloading:  true,
			NarrowRoadAccess:  false,
			ShippingFee:       150000,
			Notes:             "Kirim pagi hari sebelum jam 10",
			Items: []ports.CreateOrderItemCommand{
				{
					SKUID:               uuid.New(),
					ProductNameSnapshot: "Semen Gresik 40kg",
					UOMCodeSnapshot:     "SAK",
					Quantity:            50, // 50 x Rp 58.000 = Rp 2.900.000
					UnitPriceSnapshot:   58000,
					WeightKg:            40, // 2.000 kg
					VolumeM3:            0.03,
				},
			},
		}

		order, err := svc.CreateOrder(context.Background(), cmd)
		if err != nil {
			t.Fatalf("unexpected error creating order: %v", err)
		}

		createdOrderID = order.ID
		if order.Status != domain.OrderStatusPending {
			t.Errorf("expected status PENDING, got %s", order.Status)
		}
		if order.TotalAmount != 2900000 {
			t.Errorf("expected TotalAmount 2900000, got %f", order.TotalAmount)
		}
		if order.TotalWeightKg != 2000 {
			t.Errorf("expected TotalWeight 2000, got %f", order.TotalWeightKg)
		}
		// 50 sak x Rp 1.500 = Rp 75.000 bongkar
		if order.UnloadingFee != 75000 {
			t.Errorf("expected UnloadingFee 75000, got %f", order.UnloadingFee)
		}
		// Grand total = 2.900.000 + 150.000 (ongkir) + 75.000 (bongkar) = 3.125.000
		if order.GrandTotal != 3125000 {
			t.Errorf("expected GrandTotal 3125000, got %f", order.GrandTotal)
		}
	})

	t.Run("Siklus transisi status pesanan", func(t *testing.T) {
		// PENDING -> AWAITING_PAYMENT
		ord, err := svc.UpdateOrderStatus(context.Background(), tenantID, createdOrderID, domain.OrderStatusAwaitingPayment, "")
		if err != nil {
			t.Fatalf("failed transition to AWAITING_PAYMENT: %v", err)
		}
		if ord.Status != domain.OrderStatusAwaitingPayment {
			t.Errorf("expected status AWAITING_PAYMENT, got %s", ord.Status)
		}

		// AWAITING_PAYMENT -> PAID
		ord, err = svc.UpdateOrderStatus(context.Background(), tenantID, createdOrderID, domain.OrderStatusPaid, "")
		if err != nil {
			t.Fatalf("failed transition to PAID: %v", err)
		}
		if ord.Status != domain.OrderStatusPaid {
			t.Errorf("expected status PAID, got %s", ord.Status)
		}

		// PAID -> PROCESSING
		ord, err = svc.UpdateOrderStatus(context.Background(), tenantID, createdOrderID, domain.OrderStatusProcessing, "")
		if err != nil {
			t.Fatalf("failed transition to PROCESSING: %v", err)
		}
		if ord.Status != domain.OrderStatusProcessing {
			t.Errorf("expected status PROCESSING, got %s", ord.Status)
		}

		// PROCESSING -> COMPLETED
		ord, err = svc.UpdateOrderStatus(context.Background(), tenantID, createdOrderID, domain.OrderStatusCompleted, "")
		if err != nil {
			t.Fatalf("failed transition to COMPLETED: %v", err)
		}
		if ord.Status != domain.OrderStatusCompleted {
			t.Errorf("expected status COMPLETED, got %s", ord.Status)
		}

		// Tidak dapat membatalkan pesanan yang sudah COMPLETED
		err = svc.CancelOrder(context.Background(), tenantID, createdOrderID, "Salah pesan")
		if err != domain.ErrOrderCannotBeCancelled {
			t.Errorf("expected ErrOrderCannotBeCancelled, got %v", err)
		}
	})
}

func TestB2BRFQAndCallOff(t *testing.T) {
	rfqRepo := newMockRFQRepo()
	callOffRepo := newMockCallOffRepo()
	svc := services.NewOrderService(newMockOrderRepo(), rfqRepo, callOffRepo, nil)

	tenantID := uuid.New()
	customerID := uuid.New()

	var rfqID uuid.UUID

	t.Run("Submit RFQ Tender Proyek Kontraktor", func(t *testing.T) {
		cmd := ports.CreateRFQCommand{
			TenantID:        tenantID,
			CustomerID:      customerID,
			ProjectName:     "Pembangunan RSUD 5 Lantai Tahap 1",
			ProjectAddress:  "Jl. Merdeka Barat No. 88, Sidoarjo",
			BoQDocumentURL:  "https://storage.bangungrid.com/rfq/boq-rsud.pdf",
			EstimatedBudget: 500000000,
			Notes:           "Memerlukan spesifikasi semen tahan sulfat & split 1-2 gradasi rapat",
			Items: []ports.CreateRFQItemCommand{
				{
					MaterialName: "Semen Type V (Tahan Sulfat)",
					TargetUOM:    "TON",
					EstimatedQty: 50,
					TargetPrice:  1400000,
					Notes:        "Kemasan curah silo / jumbo bag",
				},
				{
					MaterialName: "Batu Split 1-2 SNI",
					TargetUOM:    "M3",
					EstimatedQty: 300,
					TargetPrice:  240000,
				},
			},
		}

		rfq, err := svc.CreateRFQ(context.Background(), cmd)
		if err != nil {
			t.Fatalf("unexpected error creating RFQ: %v", err)
		}

		rfqID = rfq.ID
		if rfq.Status != domain.RFQStatusSubmitted {
			t.Errorf("expected status SUBMITTED, got %s", rfq.Status)
		}
		if len(rfq.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(rfq.Items))
		}
	})

	t.Run("Rilis Quotation Penawaran Resmi oleh Sales", func(t *testing.T) {
		rfq, err := svc.UpdateRFQStatus(context.Background(), tenantID, rfqID, domain.RFQStatusQuoted, 485000000, "Harga volume khusus kontraktor tier-1 include PPN")
		if err != nil {
			t.Fatalf("unexpected error updating quote: %v", err)
		}

		if rfq.Status != domain.RFQStatusQuoted {
			t.Errorf("expected status QUOTED, got %s", rfq.Status)
		}
		if rfq.QuotedAmount != 485000000 {
			t.Errorf("expected QuotedAmount 485000000, got %f", rfq.QuotedAmount)
		}
	})

	t.Run("Jadwalkan Pengiriman Bertahap (Call-off Delivery)", func(t *testing.T) {
		orderID := uuid.New()
		delivCmd := ports.CreateCallOffCommand{
			OrderID:               orderID,
			SequenceNo:            1, // Ritase ke-1
			RequestedDeliveryDate: time.Now().Add(24 * time.Hour),
			Notes:                 "Harap koordinasi dengan mandor Pak Joko di lokasi",
			Items: []ports.CreateCallOffItemCommand{
				{
					SKUID:       uuid.New(),
					ProductName: "Semen Type V",
					Quantity:    10,
					UOM:         "TON",
				},
			},
		}

		callOff, err := svc.CreateCallOffDelivery(context.Background(), delivCmd)
		if err != nil {
			t.Fatalf("unexpected error creating call-off: %v", err)
		}

		if callOff.Status != domain.CallOffStatusScheduled {
			t.Errorf("expected status SCHEDULED, got %s", callOff.Status)
		}
		if callOff.SequenceNo != 1 {
			t.Errorf("expected SequenceNo 1, got %d", callOff.SequenceNo)
		}

		deliveries, err := svc.ListCallOffDeliveries(context.Background(), orderID)
		if err != nil {
			t.Fatalf("unexpected error listing call-offs: %v", err)
		}
		if len(deliveries) != 1 {
			t.Errorf("expected 1 call-off delivery, got %d", len(deliveries))
		}
	})
}
