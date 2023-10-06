//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"bangungrid/pkg/response"
	httpdto "bangungrid/services/order-service/internal/adapters/handlers/http"
	"bangungrid/services/order-service/internal/core/domain"
	"bangungrid/services/order-service/internal/core/ports"
)

func getBaseURL() string {
	url := os.Getenv("ORDER_SERVICE_URL")
	if url == "" {
		return "http://localhost:8084"
	}
	return url
}

func TestOrderIntegration_Healthz(t *testing.T) {
	baseURL := getBaseURL()
	resp, err := http.Get(baseURL + "/healthz")
	if err != nil {
		t.Fatalf("failed to call /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var res response.Response[map[string]any]
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !res.Success {
		t.Errorf("expected success true, got false")
	}
	if res.Data["status"] != "UP" {
		t.Errorf("expected status UP, got %v", res.Data["status"])
	}
}

func TestOrderIntegration_EstimateOrder(t *testing.T) {
	baseURL := getBaseURL()
	client := &http.Client{Timeout: 10 * time.Second}

	reqBody := httpdto.EstimateOrderRequest{
		Items: []httpdto.OrderItemRequest{
			{
				SKUID:       uuid.New().String(),
				ProductName: "Semen Tiga Roda 50kg",
				UOMCode:     "SAK",
				Quantity:    40, // 40 x 50kg = 2.000 kg (2 Ton) -> Rekomendasi ENGKEL_4W
				UnitPrice:   68000,
				WeightKg:    50,
				VolumeM3:    0.035,
			},
		},
		IncludeUnloading: true,
		NarrowRoadAccess: false,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/orders/estimate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to estimate order: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var res response.Response[ports.OrderEstimate]
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.Data.TotalWeightKg != 2000 {
		t.Errorf("expected weight 2000kg, got %f", res.Data.TotalWeightKg)
	}
	if res.Data.RecommendedFleet != "ENGKEL_4W" {
		t.Errorf("expected fleet ENGKEL_4W, got %s", res.Data.RecommendedFleet)
	}
	// 40 sak x Rp 1.500 = Rp 60.000
	if res.Data.EstimatedUnloadingFee != 60000 {
		t.Errorf("expected unloading fee 60000, got %f", res.Data.EstimatedUnloadingFee)
	}
}

func TestOrderIntegration_CreateAndLifecycle(t *testing.T) {
	baseURL := getBaseURL()
	client := &http.Client{Timeout: 10 * time.Second}
	tenantID := uuid.New().String()
	customerID := uuid.New().String()
	branchID := uuid.New().String()

	// 1. Buat Pesanan Baru
	createReq := httpdto.CreateOrderRequest{
		BranchID:          branchID,
		CustomerID:        customerID,
		CustomerType:      "B2C_RETAIL",
		DeliveryAddress:   "Proyek Ruko Harmoni Blok B-14, BSD City",
		DeliveryLatitude:  -6.3012,
		DeliveryLongitude: 106.6521,
		IncludeUnloading:  true,
		NarrowRoadAccess:  false,
		ShippingFee:       200000,
		Notes:             "Bongkar di depan pagar proyek",
		Items: []httpdto.OrderItemRequest{
			{
				SKUID:       uuid.New().String(),
				ProductName: "Bata Ringan Hebel 10cm Standard SNI",
				UOMCode:     "PALET",
				Quantity:    5, // 5 x Rp 650.000 = Rp 3.250.000
				UnitPrice:   650000,
				WeightKg:    600, // 3.000 kg
				VolumeM3:    1.2,
			},
		},
	}
	body, _ := json.Marshal(createReq)

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("X-Tenant-ID", tenantID)
	req.Header.Set("X-User-ID", customerID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to create order: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var resOrder response.Response[domain.Order]
	if err := json.NewDecoder(resp.Body).Decode(&resOrder); err != nil {
		t.Fatalf("failed to decode order: %v", err)
	}
	orderID := resOrder.Data.ID.String()

	if resOrder.Data.Status != domain.OrderStatusPending {
		t.Errorf("expected status PENDING, got %s", resOrder.Data.Status)
	}
	if resOrder.Data.TotalAmount != 3250000 {
		t.Errorf("expected TotalAmount 3250000, got %f", resOrder.Data.TotalAmount)
	}

	// 2. Ambil Detail Pesanan by ID
	reqGet, _ := http.NewRequest(http.MethodGet, baseURL+"/api/v1/orders/"+orderID, nil)
	reqGet.Header.Set("X-Tenant-ID", tenantID)

	respGet, err := client.Do(reqGet)
	if err != nil {
		t.Fatalf("failed to get order: %v", err)
	}
	defer respGet.Body.Close()

	if respGet.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", respGet.StatusCode)
	}

	// 3. Update Status ke AWAITING_PAYMENT
	updateReq := httpdto.UpdateOrderStatusRequest{
		Status: string(domain.OrderStatusAwaitingPayment),
		Reason: "Stok teralokasi di WMS",
	}
	bodyUpd, _ := json.Marshal(updateReq)
	reqUpd, _ := http.NewRequest(http.MethodPatch, baseURL+"/api/v1/orders/"+orderID+"/status", bytes.NewReader(bodyUpd))
	reqUpd.Header.Set("X-Tenant-ID", tenantID)
	reqUpd.Header.Set("Content-Type", "application/json")

	respUpd, err := client.Do(reqUpd)
	if err != nil {
		t.Fatalf("failed to update order status: %v", err)
	}
	defer respUpd.Body.Close()

	if respUpd.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", respUpd.StatusCode)
	}
}

func TestOrderIntegration_B2BRFQAndCallOff(t *testing.T) {
	baseURL := getBaseURL()
	client := &http.Client{Timeout: 10 * time.Second}
	tenantID := uuid.New().String()
	customerID := uuid.New().String()

	// 1. Submit RFQ Tender Proyek
	rfqReq := httpdto.CreateRFQRequest{
		ProjectName:     "Pembangunan Jembatan Tol Ruas 4",
		ProjectAddress:  "Jl. Arteri Primer KM 45, Karawang",
		BoQDocumentURL:  "https://storage.bangungrid.com/rfq/boq-jembatan.pdf",
		EstimatedBudget: 1200000000,
		Notes:           "Memerlukan suplai rutin 300 ton semen curah dan 1000 m3 split",
		Items: []httpdto.RFQItemRequest{
			{
				MaterialName: "Semen Portland Tipe 1 Curah",
				TargetUOM:    "TON",
				EstimatedQty: 300,
				TargetPrice:  1100000,
			},
		},
	}
	bodyRFQ, _ := json.Marshal(rfqReq)

	reqRFQ, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/orders/rfq", bytes.NewReader(bodyRFQ))
	reqRFQ.Header.Set("X-Tenant-ID", tenantID)
	reqRFQ.Header.Set("X-User-ID", customerID)
	reqRFQ.Header.Set("Content-Type", "application/json")

	respRFQ, err := client.Do(reqRFQ)
	if err != nil {
		t.Fatalf("failed to submit RFQ: %v", err)
	}
	defer respRFQ.Body.Close()

	if respRFQ.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 for RFQ, got %d", respRFQ.StatusCode)
	}

	var resRFQ response.Response[domain.B2BRFQ]
	if err := json.NewDecoder(respRFQ.Body).Decode(&resRFQ); err != nil {
		t.Fatalf("failed to decode RFQ: %v", err)
	}
	rfqID := resRFQ.Data.ID.String()

	// 2. Update Quote (Penawaran Resmi Sales)
	quoteReq := httpdto.UpdateRFQQuoteRequest{
		Status:         "QUOTED",
		QuotedAmount:   1150000000,
		QuotationNotes: "Harga include delivery dump truck tronton",
	}
	bodyQuote, _ := json.Marshal(quoteReq)
	reqQuote, _ := http.NewRequest(http.MethodPatch, baseURL+"/api/v1/orders/rfq/"+rfqID, bytes.NewReader(bodyQuote))
	reqQuote.Header.Set("X-Tenant-ID", tenantID)
	reqQuote.Header.Set("Content-Type", "application/json")

	respQuote, err := client.Do(reqQuote)
	if err != nil {
		t.Fatalf("failed to update quote: %v", err)
	}
	defer respQuote.Body.Close()

	if respQuote.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for update quote, got %d", respQuote.StatusCode)
	}
}

