package domain

import (
	"strings"
	"time"
)

type ProductUOMConversion struct {
	ID                   string  `json:"id"`
	ProductID            string  `json:"product_id"`
	UOMCode              string  `json:"uom_code"`              // Contoh: 'SAK', 'PALET', 'M3', 'DUS', 'BATANG'
	ConversionMultiplier float64 `json:"conversion_multiplier"` // Pengali ke base_uom (contoh: 1 Palet = 40 Sak)
	WeightKg             float64 `json:"weight_kg"`             // Bobot per satuan untuk kontrol ODOL tonase
	VolumeM3             float64 `json:"volume_m3"`             // Volume per satuan untuk kubikasi kapasitas armada
	RetailPrice          float64 `json:"retail_price"`          // Harga eceran B2C
}

type ProductPriceTier struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	UOMCode     string  `json:"uom_code"`
	TierName    string  `json:"tier_name"`    // Contoh: 'B2B_PROJECT_TIER_1', 'B2B_DISTRIBUTOR'
	MinQuantity float64 `json:"min_quantity"` // Minimal kuantiti pembelian
	UnitPrice   float64 `json:"unit_price"`   // Harga diskon khusus tier
}

type Product struct {
	ID             string                 `json:"id"`
	TenantID       string                 `json:"tenant_id"`
	CategoryID     string                 `json:"category_id"`
	BrandName      string                 `json:"brand_name"`
	Name           string                 `json:"name"`
	Slug           string                 `json:"slug"`
	Description    string                 `json:"description,omitempty"`
	IsBulkMaterial bool                   `json:"is_bulk_material"` // True jika pasir, batu split, dsb
	BaseUoM        string                 `json:"base_uom"`         // Contoh: 'KG', 'SAK', 'BATANG', 'M3'
	Specifications map[string]any         `json:"specifications"`   // Karakteristik teknis (mutu beton, dimensi, sertifikasi SNI)
	UOMConversions []ProductUOMConversion `json:"uom_conversions,omitempty"`
	PriceTiers     []ProductPriceTier     `json:"price_tiers,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

func (p *Product) Validate() error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return ErrProductNameRequired
	}
	if p.TenantID == "" {
		return ErrTenantRequired
	}
	if p.BaseUoM == "" {
		return ErrBaseUOMRequired
	}
	if p.Slug == "" {
		p.Slug = slugify(p.Name)
	}
	return nil
}
