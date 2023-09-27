//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	"bangungrid/pkg/response"
	httpdto "bangungrid/services/catalog-service/internal/adapters/handlers/http"
	"bangungrid/services/catalog-service/internal/core/domain"
)

func getBaseURL() string {
	url := os.Getenv("CATALOG_SERVICE_URL")
	if url == "" {
		return "http://localhost:8086"
	}
	return url
}

func TestCatalogIntegration_Healthz(t *testing.T) {
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

func TestCatalogIntegration_BulkMaterialOperations(t *testing.T) {
	baseURL := getBaseURL()

	// 1. Ambil daftar matriks massa jenis material curah
	t.Run("ListBulkMatrices", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/v1/catalog/bulk/matrices")
		if err != nil {
			t.Fatalf("failed to get bulk matrices: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var res response.Response[[]domain.BulkDensityMatrix]
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(res.Data) == 0 {
			t.Errorf("expected seeded bulk density matrices, got 0")
		}
	})

	// 2. Kalkulasi konversi volume (m3) ke berat (Kg/Ton)
	t.Run("CalculateVolume", func(t *testing.T) {
		calcReq := httpdto.BulkVolumeCalculateRequest{
			MaterialCode: "BULK-SAND-01",
			VolumeM3:     7.0, // 1 Dump truck 7 m3
		}
		body, _ := json.Marshal(calcReq)
		resp, err := http.Post(baseURL+"/api/v1/catalog/bulk/calculate", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("failed to calculate volume: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var res response.Response[domain.VolumeCalculationResult]
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Pasir pasang: 1400 kg/m3 * 7 m3 = 9800 kg (9.8 Ton)
		if res.Data.ExpectedWeightKg != 9800 {
			t.Errorf("expected 9800 kg, got %.2f", res.Data.ExpectedWeightKg)
		}
		if res.Data.ExpectedWeightTon != 9.8 {
			t.Errorf("expected 9.8 ton, got %.2f", res.Data.ExpectedWeightTon)
		}
	})

	// 3. Verifikasi toleransi timbangan jembatan (Weighbridge)
	t.Run("ValidateWeighbridgeTolerance", func(t *testing.T) {
		// Toleransi wajar (actual 9750 kg vs expected 9800 kg -> selisih -0.51%)
		verifyReq := httpdto.WeighbridgeVerifyRequest{
			MaterialCode:   "BULK-SAND-01",
			VolumeM3:       7.0,
			ActualWeightKg: 9750.0,
		}
		body, _ := json.Marshal(verifyReq)
		resp, err := http.Post(baseURL+"/api/v1/catalog/bulk/verify-weighbridge", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("failed to validate weighbridge: %v", err)
		}
		defer resp.Body.Close()

		var res response.Response[map[string]any]
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if isNormal, ok := res.Data["is_normal"].(bool); !ok || !isNormal {
			t.Errorf("expected weighbridge weight to be normal, got: %v", res.Data)
		}

		// Anomali susut berlebih (actual 9000 kg vs expected 9800 kg -> selisih > 2%)
		verifyReqAnomaly := httpdto.WeighbridgeVerifyRequest{
			MaterialCode:   "BULK-SAND-01",
			VolumeM3:       7.0,
			ActualWeightKg: 9000.0,
		}
		bodyAnomaly, _ := json.Marshal(verifyReqAnomaly)
		respAnomaly, err := http.Post(baseURL+"/api/v1/catalog/bulk/verify-weighbridge", "application/json", bytes.NewReader(bodyAnomaly))
		if err != nil {
			t.Fatalf("failed to post anomaly weighbridge: %v", err)
		}
		defer respAnomaly.Body.Close()

		var resAnomaly response.Response[map[string]any]
		if err := json.NewDecoder(respAnomaly.Body).Decode(&resAnomaly); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if isNormal, ok := resAnomaly.Data["is_normal"].(bool); !ok || isNormal {
			t.Errorf("expected anomaly detection flag to be false (anomalous), got: %v", resAnomaly.Data)
		}
	})
}

func TestCatalogIntegration_CategoryAndProductLifecycle(t *testing.T) {
	baseURL := getBaseURL()
	tenantID := uuid.New().String()

	// 1. Buat Kategori Baru
	catReq := httpdto.CreateCategoryRequest{
		Name: "Semen & Perekat " + uuid.New().String()[:6],
	}
	bodyCat, _ := json.Marshal(catReq)

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/catalog/categories", bytes.NewReader(bodyCat))
	req.Header.Set("X-Tenant-ID", tenantID)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	respCat, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	defer respCat.Body.Close()

	if respCat.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 for category, got %d", respCat.StatusCode)
	}

	var resCat response.Response[domain.Category]
	if err := json.NewDecoder(respCat.Body).Decode(&resCat); err != nil {
		t.Fatalf("failed to decode category response: %v", err)
	}

	categoryID := resCat.Data.ID
	if categoryID == "" {
		t.Fatalf("expected valid category ID")
	}

	// 2. Buat Produk Lengkap dengan Multi-UoM & Tier Harga
	prodReq := httpdto.CreateProductRequest{
		CategoryID:     categoryID,
		BrandName:      "Semen Gresik",
		Name:           fmt.Sprintf("Semen PCC 40kg (%s)", uuid.New().String()[:6]),
		Description:    "Semen Portland Composite untuk plesteran, acian, dan cor",
		IsBulkMaterial: false,
		BaseUoM:        "SAK",
		Specifications: map[string]any{
			"tipe": "PCC",
			"sni":  "SNI 7064:2014",
		},
		Conversions: []httpdto.UOMConversionInput{
			{
				UOMCode:              "SAK",
				ConversionMultiplier: 1.0,
				WeightKg:             40.0,
				VolumeM3:             0.03,
				RetailPrice:          65000,
			},
			{
				UOMCode:              "PALET",
				ConversionMultiplier: 40.0, // 1 Palet = 40 Sak
				WeightKg:             1600.0,
				VolumeM3:             1.2,
				RetailPrice:          2520000,
			},
		},
		PriceTiers: []httpdto.PriceTierInput{
			{
				UOMCode:     "SAK",
				TierName:    "CONTRACTOR_VOLUME_1",
				MinQuantity: 100,
				UnitPrice:   61000,
			},
		},
	}
	bodyProd, _ := json.Marshal(prodReq)

	reqProd, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/catalog/products", bytes.NewReader(bodyProd))
	reqProd.Header.Set("X-Tenant-ID", tenantID)
	reqProd.Header.Set("Content-Type", "application/json")

	respProd, err := client.Do(reqProd)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}
	defer respProd.Body.Close()

	if respProd.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 for product creation, got %d", respProd.StatusCode)
	}

	var resProd response.Response[domain.Product]
	if err := json.NewDecoder(respProd.Body).Decode(&resProd); err != nil {
		t.Fatalf("failed to decode product response: %v", err)
	}

	productID := resProd.Data.ID
	if productID == "" {
		t.Fatalf("expected valid product ID")
	}

	// 3. Ambil Detail Produk berdasarkan ID
	reqGet, _ := http.NewRequest(http.MethodGet, baseURL+"/api/v1/catalog/products/"+productID, nil)
	reqGet.Header.Set("X-Tenant-ID", tenantID)

	respGet, err := client.Do(reqGet)
	if err != nil {
		t.Fatalf("failed to fetch product by id: %v", err)
	}
	defer respGet.Body.Close()

	if respGet.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 fetching product, got %d", respGet.StatusCode)
	}

	var resGetProd response.Response[domain.Product]
	if err := json.NewDecoder(respGet.Body).Decode(&resGetProd); err != nil {
		t.Fatalf("failed to decode product detail: %v", err)
	}

	if len(resGetProd.Data.UOMConversions) != 2 {
		t.Errorf("expected 2 UOM conversions, got %d", len(resGetProd.Data.UOMConversions))
	}
	if len(resGetProd.Data.PriceTiers) != 1 {
		t.Errorf("expected 1 price tier, got %d", len(resGetProd.Data.PriceTiers))
	}

	// 4. List Produk dengan Filter
	reqList, _ := http.NewRequest(http.MethodGet, baseURL+"/api/v1/catalog/products?page=1&limit=10", nil)
	reqList.Header.Set("X-Tenant-ID", tenantID)

	respList, err := client.Do(reqList)
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	defer respList.Body.Close()

	if respList.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 listing products, got %d", respList.StatusCode)
	}

	var resList response.Response[[]domain.Product]
	if err := json.NewDecoder(respList.Body).Decode(&resList); err != nil {
		t.Fatalf("failed to decode product list: %v", err)
	}

	if len(resList.Data) == 0 {
		t.Errorf("expected at least 1 product listed, got 0")
	}
}
