//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"bangungrid/pkg/response"
	httpdto "bangungrid/services/inventory-service/internal/adapters/handlers/http"
	"bangungrid/services/inventory-service/internal/core/domain"
)

func getBaseURL() string {
	url := os.Getenv("INVENTORY_SERVICE_URL")
	if url == "" {
		return "http://localhost:8087"
	}
	return url
}

func TestInventoryIntegration_Healthz(t *testing.T) {
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

func TestInventoryIntegration_FullLifecycle(t *testing.T) {
	baseURL := getBaseURL()
	tenantID := uuid.New().String()
	client := &http.Client{Timeout: 10 * time.Second}

	// 1. Buat Gudang / Stockyard Baru
	whCode := "WH-" + uuid.New().String()[:8]
	whReq := httpdto.CreateWarehouseRequest{
		Code:       whCode,
		Name:       "Gudang Utama Distribusi Barat",
		Type:       domain.WarehouseTypeCentral,
		Address:    "Jl. Logistik Semen Raya No. 10",
		City:       "Tangerang",
		PostalCode: "15111",
		Latitude:   -6.1783,
		Longitude:  106.6319,
	}
	bodyWH, _ := json.Marshal(whReq)

	reqWH, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/inventory/warehouses", bytes.NewReader(bodyWH))
	reqWH.Header.Set("X-Tenant-ID", tenantID)
	reqWH.Header.Set("Content-Type", "application/json")

	respWH, err := client.Do(reqWH)
	if err != nil {
		t.Fatalf("failed to create warehouse: %v", err)
	}
	defer respWH.Body.Close()

	if respWH.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 for warehouse, got %d", respWH.StatusCode)
	}

	var resWH response.Response[domain.Warehouse]
	if err := json.NewDecoder(respWH.Body).Decode(&resWH); err != nil {
		t.Fatalf("failed to decode warehouse: %v", err)
	}
	warehouseID := resWH.Data.ID

	// 2. Tambah Stok Masuk (Inbound PO) - misal 1000 Sak Semen
	productID := uuid.New().String()
	stockReq := httpdto.AddStockRequest{
		WarehouseID: warehouseID,
		ProductID:   productID,
		UOMCode:     "SAK",
		Quantity:    1000.0,
		ReferenceID: "PO-INBOUND-001",
		Notes:       "Penerimaan armada tronton pabrik",
	}
	bodyStock, _ := json.Marshal(stockReq)

	reqStock, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/inventory/stock", bytes.NewReader(bodyStock))
	reqStock.Header.Set("X-Tenant-ID", tenantID)
	reqStock.Header.Set("Content-Type", "application/json")

	respStock, err := client.Do(reqStock)
	if err != nil {
		t.Fatalf("failed to add stock: %v", err)
	}
	defer respStock.Body.Close()

	if respStock.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 adding stock, got %d", respStock.StatusCode)
	}

	var resStock response.Response[domain.StockItem]
	if err := json.NewDecoder(respStock.Body).Decode(&resStock); err != nil {
		t.Fatalf("failed to decode add stock: %v", err)
	}
	if resStock.Data.PhysicalQty != 1000.0 || resStock.Data.AvailableQty != 1000.0 {
		t.Errorf("expected 1000 physical/available, got P:%.0f A:%.0f", resStock.Data.PhysicalQty, resStock.Data.AvailableQty)
	}

	// 3. Reservasi Stok (Soft-Locking untuk Pesanan Kontraktor 300 Sak)
	orderID := "SO-PROJECT-KONTRAKTOR-99"
	reserveReq := httpdto.ReserveStockRequest{
		WarehouseID: warehouseID,
		ProductID:   productID,
		OrderID:     orderID,
		Quantity:    300.0,
		UOMCode:     "SAK",
		TTLSeconds:  600,
	}
	bodyReserve, _ := json.Marshal(reserveReq)

	reqReserve, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/inventory/reservations", bytes.NewReader(bodyReserve))
	reqReserve.Header.Set("X-Tenant-ID", tenantID)
	reqReserve.Header.Set("Content-Type", "application/json")

	respReserve, err := client.Do(reqReserve)
	if err != nil {
		t.Fatalf("failed to reserve stock: %v", err)
	}
	defer respReserve.Body.Close()

	if respReserve.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 reserving stock, got %d", respReserve.StatusCode)
	}

	var resReserve response.Response[domain.StockReservation]
	if err := json.NewDecoder(respReserve.Body).Decode(&resReserve); err != nil {
		t.Fatalf("failed to decode reservation: %v", err)
	}
	reservationID := resReserve.Data.ID

	// Cek status stok pasca reservasi: Physical = 1000, Reserved = 300, Available = 700
	reqGetStock, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/inventory/stock/%s/%s", baseURL, warehouseID, productID), nil)
	reqGetStock.Header.Set("X-Tenant-ID", tenantID)

	respGetStock, err := client.Do(reqGetStock)
	if err != nil {
		t.Fatalf("failed to get stock: %v", err)
	}
	defer respGetStock.Body.Close()

	var resGetStock response.Response[domain.StockItem]
	_ = json.NewDecoder(respGetStock.Body).Decode(&resGetStock)
	if resGetStock.Data.ReservedQty != 300.0 || resGetStock.Data.AvailableQty != 700.0 {
		t.Errorf("expected 300 reserved and 700 available, got R:%.0f A:%.0f", resGetStock.Data.ReservedQty, resGetStock.Data.AvailableQty)
	}

	// 4. Komit Reservasi (Surat Jalan DO dicetak & barang dimuat -> Potong Stok Fisik)
	reqCommit, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/inventory/reservations/%s/commit", baseURL, reservationID), nil)
	reqCommit.Header.Set("X-Tenant-ID", tenantID)

	respCommit, err := client.Do(reqCommit)
	if err != nil {
		t.Fatalf("failed to commit reservation: %v", err)
	}
	defer respCommit.Body.Close()

	if respCommit.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 committing reservation, got %d", respCommit.StatusCode)
	}

	// 5. Penyesuaian Susut (Shrinkage Adjustment) - misal 50 sak rusak/sobek saat muat
	shrinkReq := httpdto.AdjustShrinkageRequest{
		WarehouseID:  warehouseID,
		ProductID:    productID,
		UOMCode:      "SAK",
		ShrinkageQty: 50.0,
		Reason:       "Kemasan sobek saat bongkar muat forklift",
		ReferenceID:  "SHRINKAGE-AUDIT-001",
	}
	bodyShrink, _ := json.Marshal(shrinkReq)

	reqShrink, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/inventory/stock/shrinkage", bytes.NewReader(bodyShrink))
	reqShrink.Header.Set("X-Tenant-ID", tenantID)
	reqShrink.Header.Set("Content-Type", "application/json")

	respShrink, err := client.Do(reqShrink)
	if err != nil {
		t.Fatalf("failed to adjust shrinkage: %v", err)
	}
	defer respShrink.Body.Close()

	if respShrink.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 adjusting shrinkage, got %d", respShrink.StatusCode)
	}

	// 6. Verifikasi Buku Besar Audit (Movement Ledger)
	reqMv, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/inventory/movements?warehouse_id=%s&product_id=%s", baseURL, warehouseID, productID), nil)
	reqMv.Header.Set("X-Tenant-ID", tenantID)

	respMv, err := client.Do(reqMv)
	if err != nil {
		t.Fatalf("failed to list movements: %v", err)
	}
	defer respMv.Body.Close()

	if respMv.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 listing movements, got %d", respMv.StatusCode)
	}

	var resMv response.Response[[]domain.StockMovement]
	_ = json.NewDecoder(respMv.Body).Decode(&resMv)

	// Harus ada 3 transaksi movement: Inbound PO (+1000), Outbound SO (-300), Shrinkage (-50)
	if len(resMv.Data) != 3 {
		t.Errorf("expected 3 movements in audit ledger, got %d", len(resMv.Data))
	}
}
