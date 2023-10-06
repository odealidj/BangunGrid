package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"bangungrid/pkg/response"
	handlers "bangungrid/services/order-service/internal/adapters/handlers/http"
	"bangungrid/services/order-service/internal/core/domain"
	"bangungrid/services/order-service/internal/core/ports"
	"bangungrid/services/order-service/internal/core/services"
)

type mockOrderRepo struct {
	orders map[uuid.UUID]*domain.Order
}

func (m *mockOrderRepo) Save(ctx context.Context, order *domain.Order) error {
	m.orders[order.ID] = order
	return nil
}
func (m *mockOrderRepo) FindByID(ctx context.Context, tenantID, orderID uuid.UUID) (*domain.Order, error) {
	o, exists := m.orders[orderID]
	if !exists || o.TenantID != tenantID {
		return nil, domain.ErrOrderNotFound
	}
	return o, nil
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
	var res []*domain.Order
	for _, o := range m.orders {
		if o.TenantID == filter.TenantID {
			res = append(res, o)
		}
	}
	return res, int64(len(res)), nil
}
func (m *mockOrderRepo) UpdateStatus(ctx context.Context, tenantID, orderID uuid.UUID, status domain.OrderStatus, reason string) error {
	o, exists := m.orders[orderID]
	if !exists || o.TenantID != tenantID {
		return domain.ErrOrderNotFound
	}
	o.Status = status
	if reason != "" {
		o.CancelReason = reason
	}
	return nil
}
func (m *mockOrderRepo) GetNextSequence(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return 101, nil
}

type mockRFQRepo struct {
	rfqs map[uuid.UUID]*domain.B2BRFQ
}

func (m *mockRFQRepo) Save(ctx context.Context, rfq *domain.B2BRFQ) error {
	m.rfqs[rfq.ID] = rfq
	return nil
}
func (m *mockRFQRepo) FindByID(ctx context.Context, tenantID, rfqID uuid.UUID) (*domain.B2BRFQ, error) {
	r, exists := m.rfqs[rfqID]
	if !exists || r.TenantID != tenantID {
		return nil, domain.ErrRFQNotFound
	}
	return r, nil
}
func (m *mockRFQRepo) List(ctx context.Context, tenantID uuid.UUID, customerID *uuid.UUID) ([]*domain.B2BRFQ, error) {
	var res []*domain.B2BRFQ
	for _, r := range m.rfqs {
		if r.TenantID == tenantID {
			res = append(res, r)
		}
	}
	return res, nil
}
func (m *mockRFQRepo) UpdateQuote(ctx context.Context, tenantID, rfqID uuid.UUID, status domain.RFQStatus, amount float64, notes string) error {
	r, exists := m.rfqs[rfqID]
	if !exists || r.TenantID != tenantID {
		return domain.ErrRFQNotFound
	}
	r.Status = status
	r.QuotedAmount = amount
	r.QuotationNotes = notes
	return nil
}

type mockCallOffRepo struct {
	callOffs map[uuid.UUID]*domain.B2BCallOffDelivery
}

func (m *mockCallOffRepo) Save(ctx context.Context, c *domain.B2BCallOffDelivery) error {
	m.callOffs[c.ID] = c
	return nil
}
func (m *mockCallOffRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.B2BCallOffDelivery, error) {
	c, exists := m.callOffs[id]
	if !exists {
		return nil, domain.ErrCallOffNotFound
	}
	return c, nil
}
func (m *mockCallOffRepo) ListByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.B2BCallOffDelivery, error) {
	var res []*domain.B2BCallOffDelivery
	for _, c := range m.callOffs {
		if c.OrderID == orderID {
			res = append(res, c)
		}
	}
	return res, nil
}
func (m *mockCallOffRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CallOffStatus) error {
	c, exists := m.callOffs[id]
	if !exists {
		return domain.ErrCallOffNotFound
	}
	c.Status = status
	return nil
}

func setupTestRouter() (http.Handler, uuid.UUID) {
	orderRepo := &mockOrderRepo{orders: make(map[uuid.UUID]*domain.Order)}
	rfqRepo := &mockRFQRepo{rfqs: make(map[uuid.UUID]*domain.B2BRFQ)}
	callOffRepo := &mockCallOffRepo{callOffs: make(map[uuid.UUID]*domain.B2BCallOffDelivery)}

	svc := services.NewOrderService(orderRepo, rfqRepo, callOffRepo, nil)
	handler := handlers.NewOrderHandler(svc)
	router := handlers.NewRouter(handler)

	tenantID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	return router, tenantID
}

func TestHTTPHealthz(t *testing.T) {
	router, _ := setupTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestHTTPEstimateOrder(t *testing.T) {
	router, _ := setupTestRouter()

	reqBody := handlers.EstimateOrderRequest{
		Items: []handlers.OrderItemRequest{
			{
				SKUID:       uuid.New().String(),
				ProductName: "Semen Portland 50kg",
				UOMCode:     "SAK",
				Quantity:    10,
				UnitPrice:   65000,
				WeightKg:    50,
				VolumeM3:    0.035,
			},
		},
		IncludeUnloading: true,
		NarrowRoadAccess: false,
	}
	data, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/estimate", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var res response.Response[ports.OrderEstimate]
	_ = json.NewDecoder(w.Body).Decode(&res)
	if res.Data.TotalWeightKg != 500 {
		t.Errorf("expected weight 500, got %f", res.Data.TotalWeightKg)
	}
}

func TestHTTPCreateAndGetOrder(t *testing.T) {
	router, tenantID := setupTestRouter()

	createReq := handlers.CreateOrderRequest{
		BranchID:          uuid.New().String(),
		CustomerID:        uuid.New().String(),
		CustomerType:      "B2C_RETAIL",
		DeliveryAddress:   "Jl. Daan Mogot KM 11",
		DeliveryLatitude:  -6.15,
		DeliveryLongitude: 106.75,
		IncludeUnloading:  false,
		NarrowRoadAccess:  false,
		ShippingFee:       100000,
		Items: []handlers.OrderItemRequest{
			{
				SKUID:       uuid.New().String(),
				ProductName: "Pasir Pasang",
				UOMCode:     "M3",
				Quantity:    2,
				UnitPrice:   300000,
				WeightKg:    1400,
				VolumeM3:    1.0,
			},
		},
	}
	data, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(data))
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var res response.Response[domain.Order]
	_ = json.NewDecoder(w.Body).Decode(&res)
	orderID := res.Data.ID.String()

	// GET order
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/orders/"+orderID, nil)
	reqGet.Header.Set("X-Tenant-ID", tenantID.String())
	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", wGet.Code)
	}
}

func TestHTTPB2BRFQAndCallOff(t *testing.T) {
	router, tenantID := setupTestRouter()

	// RFQ
	rfqReq := handlers.CreateRFQRequest{
		ProjectName:     "Proyek Pergudangan Cikupa",
		ProjectAddress:  "Kawasan Industri Cikupa",
		EstimatedBudget: 800000000,
		Items: []handlers.RFQItemRequest{
			{
				MaterialName: "Batu Split 2-3",
				TargetUOM:    "M3",
				EstimatedQty: 500,
			},
		},
	}
	rfqData, _ := json.Marshal(rfqReq)

	reqRFQ := httptest.NewRequest(http.MethodPost, "/api/v1/orders/rfq", bytes.NewReader(rfqData))
	reqRFQ.Header.Set("X-Tenant-ID", tenantID.String())
	reqRFQ.Header.Set("Content-Type", "application/json")
	wRFQ := httptest.NewRecorder()
	router.ServeHTTP(wRFQ, reqRFQ)

	if wRFQ.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for RFQ, got %d", wRFQ.Code)
	}

	// Call-Off
	orderID := uuid.New().String()
	callOffReq := handlers.CreateCallOffRequest{
		SequenceNo:            1,
		RequestedDeliveryDate: time.Now().Add(48 * time.Hour).Format("2006-01-02"),
		Notes:                 "Ritase pertama 50 sak",
		Items: []handlers.CallOffItemRequest{
			{
				SKUID:       uuid.New().String(),
				ProductName: "Semen Gresik",
				Quantity:    50,
				UOM:         "SAK",
			},
		},
	}
	callOffData, _ := json.Marshal(callOffReq)

	reqCO := httptest.NewRequest(http.MethodPost, "/api/v1/orders/"+orderID+"/call-offs", bytes.NewReader(callOffData))
	reqCO.Header.Set("X-Tenant-ID", tenantID.String())
	reqCO.Header.Set("Content-Type", "application/json")
	wCO := httptest.NewRecorder()
	router.ServeHTTP(wCO, reqCO)

	if wCO.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for call-off, got %d", wCO.Code)
	}
}

