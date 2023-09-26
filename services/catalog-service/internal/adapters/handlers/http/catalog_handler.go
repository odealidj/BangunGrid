package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"bangungrid/pkg/response"
	"bangungrid/services/catalog-service/internal/core/domain"
	"bangungrid/services/catalog-service/internal/core/ports"
)

type CatalogHandler struct {
	useCase ports.CatalogUseCasePort
}

func NewCatalogHandler(useCase ports.CatalogUseCasePort) *CatalogHandler {
	return &CatalogHandler{useCase: useCase}
}

func getTenantID(r *http.Request) string {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		return "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" // Default tenant
	}
	return tenantID
}

// ==========================================
// Kategori Handlers
// ==========================================

func (h *CatalogHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	cmd := ports.CreateCategoryCommand{
		TenantID: getTenantID(r),
		ParentID: req.ParentID,
		Name:     req.Name,
	}

	cat, err := h.useCase.CreateCategory(r.Context(), cmd)
	if err != nil {
		response.JSONError(w, r, http.StatusUnprocessableEntity, err.Error(), "CREATE_CATEGORY_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusCreated, "Kategori berhasil dibuat", cat)
}

func (h *CatalogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	categories, err := h.useCase.ListCategories(r.Context(), tenantID)
	if err != nil {
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "FETCH_CATEGORIES_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Daftar kategori berhasil dimuat", categories)
}

// ==========================================
// Produk Handlers
// ==========================================

func (h *CatalogHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	var convCmds []ports.CreateUOMConversionCommand
	for _, c := range req.Conversions {
		convCmds = append(convCmds, ports.CreateUOMConversionCommand{
			UOMCode:              c.UOMCode,
			ConversionMultiplier: c.ConversionMultiplier,
			WeightKg:             c.WeightKg,
			VolumeM3:             c.VolumeM3,
			RetailPrice:          c.RetailPrice,
		})
	}

	var tierCmds []ports.CreatePriceTierCommand
	for _, t := range req.PriceTiers {
		tierCmds = append(tierCmds, ports.CreatePriceTierCommand{
			UOMCode:     t.UOMCode,
			TierName:    t.TierName,
			MinQuantity: t.MinQuantity,
			UnitPrice:   t.UnitPrice,
		})
	}

	cmd := ports.CreateProductCommand{
		TenantID:       getTenantID(r),
		CategoryID:     req.CategoryID,
		BrandName:      req.BrandName,
		Name:           req.Name,
		Description:    req.Description,
		IsBulkMaterial: req.IsBulkMaterial,
		BaseUoM:        req.BaseUoM,
		Specifications: req.Specifications,
		Conversions:    convCmds,
		PriceTiers:     tierCmds,
	}

	prod, err := h.useCase.CreateProduct(r.Context(), cmd)
	if err != nil {
		response.JSONError(w, r, http.StatusUnprocessableEntity, err.Error(), "CREATE_PRODUCT_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusCreated, "Produk berhasil dibuat", prod)
}

func (h *CatalogHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	prodID := chi.URLParam(r, "id")

	prod, err := h.useCase.GetProductByID(r.Context(), tenantID, prodID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrProductNotFound) {
			status = http.StatusNotFound
		}
		response.JSONError(w, r, status, err.Error(), "PRODUCT_NOT_FOUND", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Detail produk berhasil dimuat", prod)
}

func (h *CatalogHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	categoryID := r.URL.Query().Get("category_id")
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	var isBulkOnly *bool
	if isBulkStr := r.URL.Query().Get("is_bulk"); isBulkStr != "" {
		isBulk := isBulkStr == "true" || isBulkStr == "1"
		isBulkOnly = &isBulk
	}

	query := ports.ListProductsQuery{
		TenantID:   tenantID,
		CategoryID: categoryID,
		Search:     search,
		IsBulkOnly: isBulkOnly,
		Page:       page,
		Limit:      limit,
	}

	products, totalItems, err := h.useCase.ListProducts(r.Context(), query)
	if err != nil {
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "LIST_PRODUCTS_FAILED", nil)
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	totalPages := int((totalItems + int64(query.Limit) - 1) / int64(query.Limit))

	pagination := response.Pagination{
		Page:        query.Page,
		Limit:       query.Limit,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasNextPage: query.Page < totalPages,
		HasPrevPage: query.Page > 1,
	}

	response.JSONPaginated(w, r, http.StatusOK, "Daftar produk berhasil dimuat", products, pagination)
}

func (h *CatalogHandler) AddUOMConversion(w http.ResponseWriter, r *http.Request) {
	prodID := chi.URLParam(r, "id")
	var req AddUOMConversionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	cmd := ports.CreateUOMConversionCommand{
		UOMCode:              req.UOMCode,
		ConversionMultiplier: req.ConversionMultiplier,
		WeightKg:             req.WeightKg,
		VolumeM3:             req.VolumeM3,
		RetailPrice:          req.RetailPrice,
	}

	conv, err := h.useCase.AddUOMConversion(r.Context(), prodID, cmd)
	if err != nil {
		response.JSONError(w, r, http.StatusUnprocessableEntity, err.Error(), "ADD_UOM_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusCreated, "Konversi UoM berhasil ditambahkan", conv)
}

// ==========================================
// Material Curah Handlers
// ==========================================

func (h *CatalogHandler) ListBulkDensity(w http.ResponseWriter, r *http.Request) {
	matrices, err := h.useCase.ListBulkDensityMatrices(r.Context())
	if err != nil {
		response.JSONError(w, r, http.StatusInternalServerError, err.Error(), "FETCH_BULK_DENSITY_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Matriks massa jenis material curah berhasil dimuat", matrices)
}

func (h *CatalogHandler) CalculateBulkVolume(w http.ResponseWriter, r *http.Request) {
	var req BulkVolumeCalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	res, err := h.useCase.CalculateBulkMaterialVolume(r.Context(), req.MaterialCode, req.VolumeM3)
	if err != nil {
		response.JSONError(w, r, http.StatusNotFound, err.Error(), "CALCULATION_FAILED", nil)
		return
	}

	response.JSONSuccess(w, r, http.StatusOK, "Perhitungan estimasi bobot material curah berhasil", res)
}

func (h *CatalogHandler) VerifyWeighbridge(w http.ResponseWriter, r *http.Request) {
	var req WeighbridgeVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSONError(w, r, http.StatusBadRequest, "Format JSON tidak valid", "MALFORMED_JSON", nil)
		return
	}

	isNormal, diffPct, msg, err := h.useCase.ValidateWeighbridgeTolerance(r.Context(), req.MaterialCode, req.VolumeM3, req.ActualWeightKg)
	if err != nil {
		response.JSONError(w, r, http.StatusNotFound, err.Error(), "VERIFICATION_FAILED", nil)
		return
	}

	resData := map[string]any{
		"material_code":    req.MaterialCode,
		"volume_m3":        req.VolumeM3,
		"actual_weight_kg": req.ActualWeightKg,
		"is_normal":        isNormal,
		"diff_percentage":  diffPct,
		"evaluation":       msg,
	}

	response.JSONSuccess(w, r, http.StatusOK, "Verifikasi jembatan timbang berhasil", resData)
}
