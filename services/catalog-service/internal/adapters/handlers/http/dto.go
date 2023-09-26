package http

type CreateCategoryRequest struct {
	ParentID *string `json:"parent_id,omitempty"`
	Name     string  `json:"name"`
}

type UOMConversionInput struct {
	UOMCode              string  `json:"uom_code"`
	ConversionMultiplier float64 `json:"conversion_multiplier"`
	WeightKg             float64 `json:"weight_kg"`
	VolumeM3             float64 `json:"volume_m3"`
	RetailPrice          float64 `json:"retail_price"`
}

type PriceTierInput struct {
	UOMCode     string  `json:"uom_code"`
	TierName    string  `json:"tier_name"`
	MinQuantity float64 `json:"min_quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type CreateProductRequest struct {
	CategoryID     string                 `json:"category_id"`
	BrandName      string                 `json:"brand_name"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	IsBulkMaterial bool                   `json:"is_bulk_material"`
	BaseUoM        string                 `json:"base_uom"`
	Specifications map[string]any         `json:"specifications,omitempty"`
	Conversions    []UOMConversionInput   `json:"conversions,omitempty"`
	PriceTiers     []PriceTierInput       `json:"price_tiers,omitempty"`
}

type AddUOMConversionRequest struct {
	UOMCode              string  `json:"uom_code"`
	ConversionMultiplier float64 `json:"conversion_multiplier"`
	WeightKg             float64 `json:"weight_kg"`
	VolumeM3             float64 `json:"volume_m3"`
	RetailPrice          float64 `json:"retail_price"`
}

type BulkVolumeCalculateRequest struct {
	MaterialCode string  `json:"material_code"`
	VolumeM3     float64 `json:"volume_m3"`
}

type WeighbridgeVerifyRequest struct {
	MaterialCode   string  `json:"material_code"`
	VolumeM3       float64 `json:"volume_m3"`
	ActualWeightKg float64 `json:"actual_weight_kg"`
}
