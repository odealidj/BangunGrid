'use client';

import React, { useState, useEffect } from 'react';
import { Warehouse as WarehouseIcon, MapPin, Database, ShieldAlert, CheckCircle2 } from 'lucide-react';
import { apiService } from '@/services/api';
import { Warehouse, ATPStock } from '@/types';

export default function WarehouseStockViewer() {
  const [warehouses, setWarehouses] = useState<Warehouse[]>([]);
  const [selectedWhId, setSelectedWhId] = useState<string>('');
  const [atpData, setAtpData] = useState<ATPStock | null>(null);

  useEffect(() => {
    async function loadWh() {
      const list = await apiService.getWarehouses();
      setWarehouses(list);
      if (list.length > 0) {
        setSelectedWhId(list[0].id);
      }
    }
    loadWh();
  }, []);

  useEffect(() => {
    async function loadATP() {
      if (!selectedWhId) return;
      const data = await apiService.getStockATP(selectedWhId, 'prod-001');
      setAtpData(data);
    }
    loadATP();
  }, [selectedWhId]);

  return (
    <section className="container" id="gudang-atp" style={{ margin: '4rem auto' }}>
      <div className="section-header">
        <span className="section-tag">AUDIT MANAJEMEN INVENTORI</span>
        <h2 className="section-title">Ketersediaan Stok Real-Time (ATP) Multi-Gudang</h2>
        <p className="section-subtitle">
          Alokasi kuota stok transparan dengan rumus <em>Available-To-Promise (ATP = Fisik - Reserved)</em> dan sistem
          kunci terdistribusi Redis SetNX untuk mencegah overselling ribuan sak semen saat checkout serentak.
        </p>
      </div>

      <div className="warehouse-grid">
        {warehouses.map((wh) => {
          const isSelected = wh.id === selectedWhId;

          return (
            <div
              key={wh.id}
              className="glass-card"
              style={{
                cursor: 'pointer',
                borderColor: isSelected ? 'var(--amber-primary)' : 'var(--border-card)',
                background: isSelected ? 'rgba(24, 36, 60, 0.95)' : 'var(--bg-card)'
              }}
              onClick={() => setSelectedWhId(wh.id)}
            >
              <div className="wh-card-header">
                <div>
                  <h3 style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.25rem' }}>{wh.name}</h3>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem', color: 'var(--text-dim)', fontSize: '0.8rem' }}>
                    <MapPin size={14} />
                    <span>{wh.city}</span>
                  </div>
                </div>
                <span className="wh-type-tag">{wh.type}</span>
              </div>

              <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginBottom: '1rem' }}>
                {wh.address}
              </p>

              {/* ATP Metrics */}
              <div className="stock-meter-box">
                <div>
                  <span className="meter-stat-label">Stok Fisik Tersimpan</span>
                  <div className="meter-stat-val" style={{ color: 'var(--text-main)' }}>
                    {isSelected && atpData ? atpData.physical_qty : 450} Sak
                  </div>
                </div>
                <div>
                  <span className="meter-stat-label">Terkunci / Reserved</span>
                  <div className="meter-stat-val" style={{ color: 'var(--amber-primary)' }}>
                    {isSelected && atpData ? atpData.reserved_qty : 70} Sak
                  </div>
                </div>
              </div>

              <div style={{ marginTop: '0.75rem', padding: '0.75rem', background: 'rgba(16, 185, 129, 0.1)', borderRadius: 'var(--radius-sm)', border: '1px solid rgba(16, 185, 129, 0.25)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--emerald-status)' }}>
                  Siap Dipesan (ATP):
                </span>
                <span style={{ fontSize: '1.1rem', fontWeight: 800, color: 'var(--emerald-status)' }}>
                  {isSelected && atpData ? atpData.available_qty : 380} Sak
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </section>
  );
}
