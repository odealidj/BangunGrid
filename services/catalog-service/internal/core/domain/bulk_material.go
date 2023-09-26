package domain

import (
	"fmt"
	"math"
)

type BulkDensityMatrix struct {
	ID                       string  `json:"id"`
	MaterialCode             string  `json:"material_code"`             // Contoh: 'BULK-SAND-01' (Pasir Pasang)
	MaterialName             string  `json:"material_name"`             // Contoh: 'Pasir Pasang Halus'
	StandardDensityKgPerM3   float64 `json:"standard_density_kg_per_m3"` // Contoh: 1400.0 (1 m3 = 1.4 Ton)
	NormalTolerancePct       float64 `json:"normal_tolerance_pct"`       // Contoh: 2.0 (toleransi susut/air 2.0%)
}

func (b *BulkDensityMatrix) Validate() error {
	if b.MaterialCode == "" || b.MaterialName == "" {
		return ErrBulkMaterialNotFound
	}
	if b.StandardDensityKgPerM3 <= 0 {
		return ErrInvalidDensity
	}
	if b.NormalTolerancePct <= 0 {
		b.NormalTolerancePct = 2.0
	}
	return nil
}

// VolumeCalculationResult memuat hasil kalkulasi konversi material curah antara volume (m3) dan tonase (Kg)
type VolumeCalculationResult struct {
	MaterialCode       string  `json:"material_code"`
	VolumeM3           float64 `json:"volume_m3"`
	ExpectedWeightKg   float64 `json:"expected_weight_kg"`
	ExpectedWeightTon  float64 `json:"expected_weight_ton"`
	MinToleranceKg     float64 `json:"min_tolerance_kg"`
	MaxToleranceKg     float64 `json:"max_tolerance_kg"`
	StandardDensity    float64 `json:"standard_density_kg_per_m3"`
	TolerancePct       float64 `json:"tolerance_pct"`
}

// CalculateWeightFromVolume menghitung ekspektasi bobot tonase dari volume komersial m3
func (b *BulkDensityMatrix) CalculateWeightFromVolume(volumeM3 float64) VolumeCalculationResult {
	expectedKg := volumeM3 * b.StandardDensityKgPerM3
	toleranceMargin := expectedKg * (b.NormalTolerancePct / 100.0)

	return VolumeCalculationResult{
		MaterialCode:      b.MaterialCode,
		VolumeM3:          volumeM3,
		ExpectedWeightKg:  math.Round(expectedKg*100) / 100,
		ExpectedWeightTon: math.Round((expectedKg/1000.0)*100) / 100,
		MinToleranceKg:    math.Round((expectedKg-toleranceMargin)*100) / 100,
		MaxToleranceKg:    math.Round((expectedKg+toleranceMargin)*100) / 100,
		StandardDensity:   b.StandardDensityKgPerM3,
		TolerancePct:      b.NormalTolerancePct,
	}
}

// IsWeightWithinTolerance memeriksa apakah hasil timbangan nyata (weighbridge) masih berada dalam batas toleransi wajar
func (b *BulkDensityMatrix) IsWeightWithinTolerance(volumeM3, actualWeightKg float64) (bool, float64, string) {
	expectedKg := volumeM3 * b.StandardDensityKgPerM3
	diffKg := actualWeightKg - expectedKg
	diffPct := (diffKg / expectedKg) * 100.0

	isNormal := math.Abs(diffPct) <= b.NormalTolerancePct
	var statusMsg string
	if isNormal {
		statusMsg = "NORMAL: Bobot timbangan sesuai batas toleransi susut wajar"
	} else if diffPct < 0 {
		statusMsg = fmt.Sprintf("ANOMALI PENYUSUTAN: Muatan kurang %.2f%% (melebihi batas toleransi %.1f%%)", math.Abs(diffPct), b.NormalTolerancePct)
	} else {
		statusMsg = fmt.Sprintf("ANOMALI KELEBIHAN: Muatan melebihi %.2f%% (potensi kadar air tinggi / lumpur)", diffPct)
	}

	return isNormal, diffPct, statusMsg
}
