package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"bangungrid/pkg/response"
	"bangungrid/services/inventory-service/internal/core/ports"
)

type InventoryHandler struct {
	useCase ports.InventoryUseCasePort
}

func NewInventoryHandler(useCase ports.InventoryUseCasePort) *InventoryHandler {
	return &InventoryHandler{useCase: useCase}
}

func getTenantID(r *http.Request) string {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		return "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	}
	return tenantID
}

// ==========================================
// Warehouse Handlers
// ==========================================

func (h *InventoryHandler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	var req CreateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	cmd := ports.CreateWarehouseCommand{
		TenantID:   getTenantID(r),
		Code:       req.Code,
		Name:       req.Name,
		Type:       req.Type,
		Address:    req.Address,
		City:       req.City,
		PostalCode: req.PostalCode,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
	}

	wh, err := h.useCase.CreateWarehouse(r.Context(), cmd)
	if err != nil {
		response.JSONError(w, r, http.StatusUnprocessableEntity, err.Error(), "CREATE_WAREHOUSE_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusCreated, "Gudang / Stockyard berhasil didaftarkan", wh)
}

func (h *InventoryHandler) ListWarehouses(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	list, err := h.useCase.ListWarehouses(r.Context(), tenantID)
	if err != nil {
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "LIST_WAREHOUSES_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Daftar gudang berhasil dimuat", list)
}

func (h *InventoryHandler) GetWarehouseByID(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	id := chi.URLParam(r, "id")

	wh, err := h.useCase.GetWarehouseByID(r.Context(), tenantID, id)
	if err != nil {
		response.JSONError(w, r, http.StatusNotFound, err.Error(), "WAREHOUSE_NOT_FOUND", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Detail gudang berhasil dimuat", wh)
}

// ==========================================
// Stock & ATP Handlers
// ==========================================

func (h *InventoryHandler) AddStock(w http.ResponseWriter, r *http.Request) {
	var req AddStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	cmd := ports.AddStockCommand{
		TenantID:    getTenantID(r),
		WarehouseID: req.WarehouseID,
		ProductID:   req.ProductID,
		UOMCode:     req.UOMCode,
		Quantity:    req.Quantity,
		ReferenceID: req.ReferenceID,
		Notes:       req.Notes,
	}

	stock, err := h.useCase.AddStock(r.Context(), cmd)
	if err != nil {
		response.JSONError(w, r, http.StatusUnprocessableEntity, err.Error(), "ADD_STOCK_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Stok berhasil ditambahkan", stock)
}

func (h *InventoryHandler) AdjustShrinkage(w http.ResponseWriter, r *http.Request) {
	var req AdjustShrinkageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	cmd := ports.AdjustShrinkageCommand{
		TenantID:     getTenantID(r),
		WarehouseID:  req.WarehouseID,
		ProductID:    req.ProductID,
		UOMCode:      req.UOMCode,
		ShrinkageQty: req.ShrinkageQty,
		Reason:       req.Reason,
		ReferenceID:  req.ReferenceID,
	}

	stock, err := h.useCase.AdjustShrinkage(r.Context(), cmd)
	if err != nil {
		response.JSONError(w, r, http.StatusUnprocessableEntity, err.Error(), "ADJUST_SHRINKAGE_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Penyesuaian susut stok berhasil dicatat", stock)
}

func (h *InventoryHandler) GetStock(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	whID := chi.URLParam(r, "warehouseId")
	prodID := chi.URLParam(r, "productId")

	stock, err := h.useCase.GetStock(r.Context(), tenantID, whID, prodID)
	if err != nil {
		response.JSONError(w, r, http.StatusNotFound, err.Error(), "STOCK_NOT_FOUND", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Informasi stok & ATP berhasil dimuat", stock)
}

func (h *InventoryHandler) ListStockByWarehouse(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	whID := chi.URLParam(r, "warehouseId")

	stocks, err := h.useCase.ListStockByWarehouse(r.Context(), tenantID, whID)
	if err != nil {
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "LIST_STOCK_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Daftar stok gudang berhasil dimuat", stocks)
}

// ==========================================
// Stock Reservation Handlers
// ==========================================

func (h *InventoryHandler) ReserveStock(w http.ResponseWriter, r *http.Request) {
	var req ReserveStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	ttl := 15 * time.Minute
	if req.TTLSeconds > 0 {
		ttl = time.Duration(req.TTLSeconds) * time.Second
	}

	cmd := ports.ReserveStockCommand{
		TenantID:    getTenantID(r),
		WarehouseID: req.WarehouseID,
		ProductID:   req.ProductID,
		OrderID:     req.OrderID,
		Quantity:    req.Quantity,
		UOMCode:     req.UOMCode,
		TTL:         ttl,
	}

	res, err := h.useCase.ReserveStock(r.Context(), cmd)
	if err != nil {
		response.JSONError(w, r, http.StatusConflict, err.Error(), "RESERVATION_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusCreated, "Stok berhasil direservasi (soft-locked)", res)
}

func (h *InventoryHandler) CommitReservation(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	resID := chi.URLParam(r, "id")

	if err := h.useCase.CommitReservation(r.Context(), tenantID, resID); err != nil {
		response.JSONError(w, r, http.StatusUnprocessableEntity, err.Error(), "COMMIT_RESERVATION_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Reservasi stok berhasil dikomit (potong fisik)", map[string]string{
		"reservation_id": resID,
		"status":         "COMMITTED",
	})
}

func (h *InventoryHandler) ReleaseReservation(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	resID := chi.URLParam(r, "id")

	if err := h.useCase.ReleaseReservation(r.Context(), tenantID, resID); err != nil {
		response.JSONError(w, r, http.StatusUnprocessableEntity, err.Error(), "RELEASE_RESERVATION_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Reservasi stok berhasil dibatalkan dan dirilis", map[string]string{
		"reservation_id": resID,
		"status":         "CANCELLED",
	})
}

// ==========================================
// Audit Ledger Movements Handlers
// ==========================================

func (h *InventoryHandler) ListMovements(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	whID := r.URL.Query().Get("warehouse_id")
	prodID := r.URL.Query().Get("product_id")

	movements, err := h.useCase.ListMovements(r.Context(), tenantID, whID, prodID)
	if err != nil {
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "LIST_MOVEMENTS_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Buku besar perpindahan stok (movement ledger) berhasil dimuat", movements)
}
