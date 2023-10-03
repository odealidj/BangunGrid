'use client';

import React, { useState, useEffect } from 'react';
import { Calculator, Scale, Layers, Plus, CheckCircle, RefreshCw } from 'lucide-react';
import { apiService } from '@/services/api';
import { BulkDensityMatrix, BulkCalculationResult, Product, UOMConversion } from '@/types';

interface BulkCalculatorProps {
  onAddToCart: (product: Product, uom: UOMConversion, quantity: number) => void;
}

export default function BulkCalculator({ onAddToCart }: BulkCalculatorProps) {
  const [matrices, setMatrices] = useState<BulkDensityMatrix[]>([]);
  const [selectedMaterial, setSelectedMaterial] = useState<string>('BULK-SAND-02');
  const [volumeM3, setVolumeM3] = useState<number>(8);
  const [calcResult, setCalcResult] = useState<BulkCalculationResult | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [addedNotice, setAddedNotice] = useState<boolean>(false);

  useEffect(() => {
    async function loadMatrices() {
      const data = await apiService.getBulkMatrices();
      setMatrices(data);
    }
    loadMatrices();
  }, []);

  useEffect(() => {
    async function runCalc() {
      if (!selectedMaterial || volumeM3 <= 0) return;
      setLoading(true);
      try {
        const result = await apiService.calculateBulk(selectedMaterial, volumeM3);
        setCalcResult(result);
      } catch (err) {
        console.error('Calculation error:', err);
      } finally {
        setLoading(false);
      }
    }

    const timer = setTimeout(() => {
      runCalc();
    }, 150);

    return () => clearTimeout(timer);
  }, [selectedMaterial, volumeM3]);

  const handleAddBulkToCart = () => {
    if (!calcResult) return;

    const currentMatrix = matrices.find(m => m.material_code === selectedMaterial);
    const materialName = currentMatrix ? currentMatrix.material_name : 'Material Curah Agregat';

    // Mock/Dynamic product wrap untuk agregat curah
    const bulkProduct: Product = {
      id: `bulk-${selectedMaterial}`,
      tenant_id: 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
      category_id: 'cat-curah',
      sku: selectedMaterial,
      name: `${materialName} (Standar SNI)`,
      description: `Agregat curah alam densitas ${calcResult.standard_density_kg_per_m3} Kg/m³ toleransi timbang ±${calcResult.tolerance_pct}%.`,
      base_uom: 'M3',
      is_bulk_material: true,
      bulk_material_code: selectedMaterial,
      uom_conversions: [
        {
          id: `uom-${selectedMaterial}`,
          product_id: `bulk-${selectedMaterial}`,
          uom_code: 'M3',
          uom_name: `Kubikasi (${calcResult.standard_density_kg_per_m3} Kg/m³)`,
          conversion_factor: 1,
          weight_kg: calcResult.standard_density_kg_per_m3,
          volume_m3: 1,
          retail_price: 295000,
          is_base_uom: true
        }
      ]
    };

    onAddToCart(bulkProduct, bulkProduct.uom_conversions[0], volumeM3);
    setAddedNotice(true);
    setTimeout(() => setAddedNotice(false), 2000);
  };

  const currentMat = matrices.find(m => m.material_code === selectedMaterial);

  return (
    <section className="container" id="kalkulator-curah" style={{ margin: '4rem auto' }}>
      <div className="section-header">
        <span className="section-tag">ENGINE MATERIAL CURAH SNI</span>
        <h2 className="section-title">Kalkulator Presisi Volume & Tonase</h2>
        <p className="section-subtitle">
          Konversi dua arah otomatis antara volume komersial (m³) dan tonase timbangan fisik jembatan timbang (Kg)
          berdasarkan tabel densitas resmi SNI untuk menghindari kecurangan muatan.
        </p>
      </div>

      <div className="glass-card calc-card">
        <div className="calc-grid">
          {/* Controls */}
          <div>
            <div className="form-group">
              <label className="form-label" htmlFor="material-select">
                <span>Pilih Jenis Agregat / Material Curah</span>
                {currentMat && (
                  <span style={{ color: 'var(--amber-primary)', fontSize: '0.8rem' }}>
                    Densitas SNI: {currentMat.standard_density_kg_per_m3} Kg/m³
                  </span>
                )}
              </label>
              <select
                id="material-select"
                className="form-select"
                value={selectedMaterial}
                onChange={(e) => setSelectedMaterial(e.target.value)}
              >
                {matrices.map((m) => (
                  <option key={m.material_code} value={m.material_code}>
                    {m.material_name} ({m.standard_density_kg_per_m3} Kg/m³ - Tol ±{m.normal_tolerance_pct}%)
                  </option>
                ))}
              </select>
            </div>

            <div className="form-group">
              <div className="form-label">
                <label htmlFor="volume-input">Volume Pesanan Kubikasi (m³)</label>
                <span style={{ color: 'var(--text-amber)', fontWeight: 700, fontSize: '1rem' }}>
                  {volumeM3} m³
                </span>
              </div>
              <input
                type="range"
                min="1"
                max="50"
                step="1"
                value={volumeM3}
                onChange={(e) => setVolumeM3(Number(e.target.value))}
                className="volume-slider"
                id="volume-slider"
              />
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.75rem', color: 'var(--text-dim)', marginTop: '0.25rem' }}>
                <span>1 m³ (Pickup)</span>
                <span>8 m³ (Dump Indeks 8)</span>
                <span>24 m³ (Tronton)</span>
                <span>50 m³ (Armada Borongan)</span>
              </div>
            </div>

            {/* Manual input box */}
            <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
              <input
                id="volume-input"
                type="number"
                min="0.5"
                step="0.5"
                value={volumeM3}
                onChange={(e) => setVolumeM3(Math.max(0.5, Number(e.target.value)))}
                className="form-input"
                style={{ width: '130px' }}
              />
              <span style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
                m³ setara dengan <strong>{(volumeM3 / 8).toFixed(1)} rit</strong> Dump Truck Indeks 8
              </span>
            </div>
          </div>

          {/* Results Display */}
          <div>
            <div className="calc-result-box">
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-muted)', fontSize: '0.85rem' }}>
                  <Scale size={16} />
                  <span>Estimasi Muatan Timbangan</span>
                </div>
                {loading && <RefreshCw size={14} className="animate-spin" color="var(--amber-primary)" />}
              </div>

              <div className="result-main">
                <div>
                  <span className="result-value">
                    {calcResult ? calcResult.expected_weight_ton.toLocaleString('id-ID') : '0'}
                  </span>
                  <span className="result-unit">Ton</span>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <span style={{ fontSize: '1.15rem', fontWeight: 700, color: 'var(--text-main)' }}>
                    {calcResult ? calcResult.expected_weight_kg.toLocaleString('id-ID') : '0'}
                  </span>
                  <span style={{ fontSize: '0.8rem', color: 'var(--text-dim)', marginLeft: '0.2rem' }}>Kg</span>
                </div>
              </div>

              {/* Tolerance Box */}
              <div className="tolerance-bar">
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span>Toleransi Jembatan Timbang (±{calcResult?.tolerance_pct || 2}%):</span>
                  <span style={{ color: 'var(--emerald-status)', fontWeight: 600 }}>Penyusutan Wajar</span>
                </div>
                <div className="tolerance-range">
                  <span>Min: {calcResult?.min_tolerance_kg.toLocaleString('id-ID')} Kg</span>
                  <span>Maks: {calcResult?.max_tolerance_kg.toLocaleString('id-ID')} Kg</span>
                </div>
              </div>

              <button
                onClick={handleAddBulkToCart}
                className="btn btn-primary"
                style={{ width: '100%', marginTop: '0.5rem' }}
                id="add-bulk-to-cart-btn"
              >
                {addedNotice ? (
                  <>
                    <CheckCircle size={18} />
                    <span>Ditambahkan ke Keranjang!</span>
                  </>
                ) : (
                  <>
                    <Plus size={18} />
                    <span>Pesan {volumeM3} m³ ({calcResult?.expected_weight_ton} Ton)</span>
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
