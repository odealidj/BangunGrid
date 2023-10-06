package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"bangungrid/pkg/response"
	"bangungrid/services/order-service/internal/core/domain"
	"bangungrid/services/order-service/internal/core/ports"
)

type OrderHandler struct {
	useCase ports.OrderUseCasePort
}

func NewOrderHandler(useCase ports.OrderUseCasePort) *OrderHandler {
	return &OrderHandler{useCase: useCase}
}

func getTenantID(r *http.Request) uuid.UUID {
	val := r.Header.Get("X-Tenant-ID")
	if id, err := uuid.Parse(val); err == nil && id != uuid.Nil {
		return id
	}
	// Default dummy tenant jika dijalankan tanpa gateway
	return uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
}

func getUserID(r *http.Request) uuid.UUID {
	val := r.Header.Get("X-User-ID")
	if id, err := uuid.Parse(val); err == nil && id != uuid.Nil {
		return id
	}
	return uuid.MustParse("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b22")
}

// EstimateOrder menghitung estimasi bobot, volume, armada, biaya kuli, dan langsir
func (h *OrderHandler) EstimateOrder(w http.ResponseWriter, r *http.Request) {
	var req EstimateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	cmdItems := make([]ports.CreateOrderItemCommand, len(req.Items))
	for i, it := range req.Items {
		skuID, _ := uuid.Parse(it.SKUID)
		cmdItems[i] = ports.CreateOrderItemCommand{
			SKUID:               skuID,
			ProductNameSnapshot: it.ProductName,
			UOMCodeSnapshot:     it.UOMCode,
			Quantity:            it.Quantity,
			UnitPriceSnapshot:   it.UnitPrice,
			WeightKg:            it.WeightKg,
			VolumeM3:            it.VolumeM3,
		}
	}

	cmd := ports.EstimateOrderCommand{
		Items:            cmdItems,
		IncludeUnloading: req.IncludeUnloading,
		NarrowRoadAccess: req.NarrowRoadAccess,
	}

	est, err := h.useCase.EstimateOrder(r.Context(), cmd)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyOrderItems) {
			response.JSONError(w, r, http.StatusBadRequest, err.Error(), "EMPTY_ORDER_ITEMS", nil)
			return
		}
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "CALCULATION_ERROR", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Estimasi pesanan dan muatan berhasil dihitung", est)
}

// CreateOrder membuat pesanan baru
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	tenantID := getTenantID(r)
	customerID := getUserID(r)
	if req.CustomerID != "" {
		if cid, err := uuid.Parse(req.CustomerID); err == nil {
			customerID = cid
		}
	}

	branchID := uuid.Nil
	if req.BranchID != "" {
		if bid, err := uuid.Parse(req.BranchID); err == nil {
			branchID = bid
		}
	}
	if branchID == uuid.Nil {
		branchID = uuid.MustParse("c0eebc99-9c0b-4ef8-bb6d-6bb9bd380c33")
	}

	custType := domain.CustomerTypeB2CRetail
	if req.CustomerType == string(domain.CustomerTypeB2BContractor) {
		custType = domain.CustomerTypeB2BContractor
	}

	cmdItems := make([]ports.CreateOrderItemCommand, len(req.Items))
	for i, it := range req.Items {
		skuID, _ := uuid.Parse(it.SKUID)
		cmdItems[i] = ports.CreateOrderItemCommand{
			SKUID:               skuID,
			ProductNameSnapshot: it.ProductName,
			UOMCodeSnapshot:     it.UOMCode,
			Quantity:            it.Quantity,
			UnitPriceSnapshot:   it.UnitPrice,
			WeightKg:            it.WeightKg,
			VolumeM3:            it.VolumeM3,
		}
	}

	cmd := ports.CreateOrderCommand{
		TenantID:          tenantID,
		BranchID:          branchID,
		CustomerID:        customerID,
		CustomerType:      custType,
		DeliveryAddress:   req.DeliveryAddress,
		DeliveryLatitude:  req.DeliveryLatitude,
		DeliveryLongitude: req.DeliveryLongitude,
		IncludeUnloading:  req.IncludeUnloading,
		NarrowRoadAccess:  req.NarrowRoadAccess,
		ShippingFee:       req.ShippingFee,
		Notes:             req.Notes,
		Items:             cmdItems,
	}

	order, err := h.useCase.CreateOrder(r.Context(), cmd)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyOrderItems) || errors.Is(err, domain.ErrInvalidOrderData) || errors.Is(err, domain.ErrInvalidQuantity) || errors.Is(err, domain.ErrInvalidPrice) {
			response.JSONError(w, r, http.StatusBadRequest, err.Error(), "INVALID_ORDER_DATA", nil)
			return
		}
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "CREATE_ORDER_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusCreated, "Pesanan berhasil dibuat (menunggu konfirmasi inventori)", order)
}

// GetOrderByID mengambil detail pesanan berdasarkan ID
func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "ID pesanan tidak valid", "INVALID_ID", nil)
		return
	}

	order, err := h.useCase.GetOrderByID(r.Context(), tenantID, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			response.JSONError(w, r, http.StatusNotFound, "Pesanan tidak ditemukan", "ORDER_NOT_FOUND", nil)
			return
		}
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Detail pesanan ditemukan", order)
}

// GetOrderByOrderNo mengambil detail pesanan berdasarkan Nomor Pesanan
func (h *OrderHandler) GetOrderByOrderNo(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	orderNo := chi.URLParam(r, "orderNo")

	order, err := h.useCase.GetOrderByOrderNo(r.Context(), tenantID, orderNo)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			response.JSONError(w, r, http.StatusNotFound, "Pesanan tidak ditemukan", "ORDER_NOT_FOUND", nil)
			return
		}
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Detail pesanan ditemukan", order)
}

// ListOrders mengambil daftar pesanan
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	var customerID *uuid.UUID
	if cidStr := r.URL.Query().Get("customer_id"); cidStr != "" {
		if cid, err := uuid.Parse(cidStr); err == nil {
			customerID = &cid
		}
	}

	var status *domain.OrderStatus
	if stStr := r.URL.Query().Get("status"); stStr != "" {
		st := domain.OrderStatus(stStr)
		status = &st
	}

	filter := ports.ListOrdersFilter{
		TenantID:   tenantID,
		CustomerID: customerID,
		Status:     status,
		Limit:      limit,
		Offset:     offset,
	}

	orders, total, err := h.useCase.ListOrders(r.Context(), filter)
	if err != nil {
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "FETCH_ORDERS_FAILED", nil)
		return
	}

	pagination := response.Pagination{
		Limit:      limit,
		TotalItems: total,
	}

	response.JSONPaginated(w, r, http.StatusOK, "Daftar pesanan berhasil diambil", orders, pagination)
}

// UpdateOrderStatus memperbarui status pesanan
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "ID pesanan tidak valid", "INVALID_ID", nil)
		return
	}

	var req UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	order, err := h.useCase.UpdateOrderStatus(r.Context(), tenantID, orderID, domain.OrderStatus(req.Status), req.Reason)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			response.JSONError(w, r, http.StatusNotFound, "Pesanan tidak ditemukan", "ORDER_NOT_FOUND", nil)
			return
		}
		if errors.Is(err, domain.ErrInvalidOrderStatus) {
			response.JSONError(w, r, http.StatusBadRequest, err.Error(), "INVALID_ORDER_STATUS_TRANSITION", nil)
			return
		}
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "UPDATE_STATUS_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Status pesanan berhasil diperbarui", order)
}

// CancelOrder membatalkan pesanan
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "ID pesanan tidak valid", "INVALID_ID", nil)
		return
	}

	var req CancelOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Reason = "Dibatalkan oleh pengguna"
	}

	if err := h.useCase.CancelOrder(r.Context(), tenantID, orderID, req.Reason); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			response.JSONError(w, r, http.StatusNotFound, "Pesanan tidak ditemukan", "ORDER_NOT_FOUND", nil)
			return
		}
		if errors.Is(err, domain.ErrOrderCannotBeCancelled) {
			response.JSONError(w, r, http.StatusBadRequest, err.Error(), "ORDER_CANNOT_BE_CANCELLED", nil)
			return
		}
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "CANCEL_ORDER_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Pesanan berhasil dibatalkan", map[string]string{
		"order_id": orderIDStr,
		"status":   string(domain.OrderStatusCancelled),
	})
}

// ==========================================
// B2B RFQ (Request for Quotation) Handlers
// ==========================================

func (h *OrderHandler) CreateRFQ(w http.ResponseWriter, r *http.Request) {
	var req CreateRFQRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	tenantID := getTenantID(r)
	customerID := getUserID(r)

	items := make([]ports.CreateRFQItemCommand, len(req.Items))
	for i, it := range req.Items {
		items[i] = ports.CreateRFQItemCommand{
			MaterialName: it.MaterialName,
			TargetUOM:    it.TargetUOM,
			EstimatedQty: it.EstimatedQty,
			TargetPrice:  it.TargetPrice,
			Notes:        it.Notes,
		}
	}

	cmd := ports.CreateRFQCommand{
		TenantID:        tenantID,
		CustomerID:      customerID,
		ProjectName:     req.ProjectName,
		ProjectAddress:  req.ProjectAddress,
		BoQDocumentURL:  req.BoQDocumentURL,
		EstimatedBudget: req.EstimatedBudget,
		Notes:           req.Notes,
		Items:           items,
	}

	rfq, err := h.useCase.CreateRFQ(r.Context(), cmd)
	if err != nil {
		response.JSONError(w, r, http.StatusBadRequest, err.Error(), "CREATE_RFQ_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusCreated, "Pengajuan RFQ Tender Kontraktor berhasil disubmit", rfq)
}

func (h *OrderHandler) ListRFQs(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	var customerID *uuid.UUID
	if cidStr := r.URL.Query().Get("customer_id"); cidStr != "" {
		if cid, err := uuid.Parse(cidStr); err == nil {
			customerID = &cid
		}
	}

	rfqs, err := h.useCase.ListRFQs(r.Context(), tenantID, customerID)
	if err != nil {
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "FETCH_RFQS_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Daftar RFQ berhasil diambil", rfqs)
}

func (h *OrderHandler) GetRFQByID(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	rfqIDStr := chi.URLParam(r, "id")
	rfqID, err := uuid.Parse(rfqIDStr)
	if err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "ID RFQ tidak valid", "INVALID_ID", nil)
		return
	}

	rfq, err := h.useCase.GetRFQByID(r.Context(), tenantID, rfqID)
	if err != nil {
		if errors.Is(err, domain.ErrRFQNotFound) {
			response.JSONError(w, r, http.StatusNotFound, "RFQ tidak ditemukan", "RFQ_NOT_FOUND", nil)
			return
		}
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Detail RFQ ditemukan", rfq)
}

func (h *OrderHandler) UpdateRFQStatus(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	rfqIDStr := chi.URLParam(r, "id")
	rfqID, err := uuid.Parse(rfqIDStr)
	if err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "ID RFQ tidak valid", "INVALID_ID", nil)
		return
	}

	var req UpdateRFQQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	rfq, err := h.useCase.UpdateRFQStatus(r.Context(), tenantID, rfqID, domain.RFQStatus(req.Status), req.QuotedAmount, req.QuotationNotes)
	if err != nil {
		if errors.Is(err, domain.ErrRFQNotFound) {
			response.JSONError(w, r, http.StatusNotFound, "RFQ tidak ditemukan", "RFQ_NOT_FOUND", nil)
			return
		}
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "UPDATE_RFQ_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Status penawaran RFQ berhasil diperbarui", rfq)
}

// ==========================================
// B2B Call-Off (Pengiriman Parsial) Handlers
// ==========================================

func (h *OrderHandler) CreateCallOffDelivery(w http.ResponseWriter, r *http.Request) {
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "ID pesanan tidak valid", "INVALID_ID", nil)
		return
	}

	var req CreateCallOffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	deliveryDate, err := time.Parse("2006-01-02", req.RequestedDeliveryDate)
	if err != nil {
		deliveryDate, err = time.Parse(time.RFC3339, req.RequestedDeliveryDate)
		if err != nil {
			deliveryDate = time.Now().Add(24 * time.Hour)
		}
	}

	items := make([]ports.CreateCallOffItemCommand, len(req.Items))
	for i, it := range req.Items {
		skuID, _ := uuid.Parse(it.SKUID)
		items[i] = ports.CreateCallOffItemCommand{
			SKUID:       skuID,
			ProductName: it.ProductName,
			Quantity:    it.Quantity,
			UOM:         it.UOM,
		}
	}

	cmd := ports.CreateCallOffCommand{
		OrderID:               orderID,
		SequenceNo:            req.SequenceNo,
		RequestedDeliveryDate: deliveryDate,
		Notes:                 req.Notes,
		Items:                 items,
	}

	callOff, err := h.useCase.CreateCallOffDelivery(r.Context(), cmd)
	if err != nil {
		response.JSONError(w, r, http.StatusBadRequest, err.Error(), "CREATE_CALL_OFF_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusCreated, "Jadwal pengiriman bertahap (ritase) berhasil dibuat", callOff)
}

func (h *OrderHandler) ListCallOffDeliveries(w http.ResponseWriter, r *http.Request) {
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "ID pesanan tidak valid", "INVALID_ID", nil)
		return
	}

	deliveries, err := h.useCase.ListCallOffDeliveries(r.Context(), orderID)
	if err != nil {
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "FETCH_CALL_OFFS_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Daftar jadwal pengiriman parsial berhasil diambil", deliveries)
}

func (h *OrderHandler) UpdateCallOffStatus(w http.ResponseWriter, r *http.Request) {
	delivIDStr := chi.URLParam(r, "deliveryId")
	delivID, err := uuid.Parse(delivIDStr)
	if err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "ID pengiriman parsial tidak valid", "INVALID_ID", nil)
		return
	}

	var req UpdateCallOffStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	if err := h.useCase.UpdateCallOffStatus(r.Context(), delivID, domain.CallOffStatus(req.Status)); err != nil {
		if errors.Is(err, domain.ErrCallOffNotFound) {
			response.JSONError(w, r, http.StatusNotFound, "Jadwal pengiriman parsial tidak ditemukan", "CALL_OFF_NOT_FOUND", nil)
			return
		}
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "UPDATE_CALL_OFF_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Status pengiriman parsial berhasil diperbarui", map[string]string{
		"delivery_id": delivIDStr,
		"status":      req.Status,
	})
}
