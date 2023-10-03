'use client';

import React from 'react';
import { Truck, ShieldCheck, Scale, Cpu, Search } from 'lucide-react';

interface HeroBannerProps {
  onSearch: (query: string) => void;
  searchQuery: string;
}

export default function HeroBanner({ onSearch, searchQuery }: HeroBannerProps) {
  return (
    <section className="hero-section" id="beranda">
      <div className="container hero-grid">
        {/* Left Column: Headline & Value Prop */}
        <div>
          <div className="hero-badge">
            <Cpu size={14} />
            <span>Multi-Tenancy & Edge API Gateway Ready</span>
          </div>

          <h1 className="hero-title">
            Pengadaan Agregat Curah & Logistik <span>Armada Berat</span> Proyek
          </h1>

          <p className="hero-description">
            Platform rantai pasok material konstruksi berstandar SNI langsung dari tambang quarry & terminal semen.
            Didukung kalkulator volume-tonase presisi jembatan timbang dan armada Pick-up hingga Tronton 25 Ton.
          </p>

          {/* Search Box */}
          <div style={{ maxWidth: '520px', marginBottom: '2rem', position: 'relative' }}>
            <div style={{ display: 'flex', gap: '0.5rem', background: 'var(--bg-input)', padding: '0.4rem', borderRadius: 'var(--radius-md)', border: '1px solid var(--border-card)' }}>
              <div style={{ display: 'flex', alignItems: 'center', paddingLeft: '0.75rem', color: 'var(--text-dim)' }}>
                <Search size={18} />
              </div>
              <input
                type="text"
                placeholder="Cari semen PCC, pasir cor Galunggung, batu split 1-2, besi ulir..."
                value={searchQuery}
                onChange={(e) => onSearch(e.target.value)}
                style={{
                  background: 'transparent',
                  border: 'none',
                  color: 'var(--text-main)',
                  outline: 'none',
                  width: '100%',
                  fontSize: '0.9rem'
                }}
                id="search-input"
              />
            </div>
          </div>

          {/* Key Metrics */}
          <div className="hero-features">
            <div className="hero-feat-item">
              <span className="hero-feat-value" style={{ color: 'var(--amber-primary)' }}>100%</span>
              <span className="hero-feat-label">Standar SNI & Toleransi ±2%</span>
            </div>
            <div className="hero-feat-item">
              <span className="hero-feat-value" style={{ color: 'var(--blue-tech)' }}>Multi-UoM</span>
              <span className="hero-feat-label">Sak, Palet, m³ & Tonase</span>
            </div>
            <div className="hero-feat-item">
              <span className="hero-feat-value" style={{ color: 'var(--emerald-status)' }}>Live ATP</span>
              <span className="hero-feat-label">Alokasi Stok Multi-Gudang</span>
            </div>
          </div>
        </div>

        {/* Right Column: Hero Visual Showcase */}
        <div>
          <div className="glass-card" style={{ position: 'relative', overflow: 'hidden' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.25rem', borderBottom: '1px solid var(--border-subtle)', paddingBottom: '0.75rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Truck size={20} color="var(--amber-primary)" />
                <span style={{ fontWeight: 700, fontSize: '0.95rem' }}>Simulasi Pengiriman Armada Berat</span>
              </div>
              <span style={{ fontSize: '0.75rem', color: 'var(--emerald-status)', fontWeight: 600 }}>PostGIS Geolocation</span>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.75rem', background: 'rgba(0,0,0,0.3)', borderRadius: 'var(--radius-sm)' }}>
                <div>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)', display: 'block' }}>Armada Direkomendasikan</span>
                  <span style={{ fontWeight: 700, color: 'var(--text-main)' }}>Dump Truck Indeks 8 (12 Ton)</span>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)', display: 'block' }}>Kapasitas Maks</span>
                  <span style={{ fontWeight: 700, color: 'var(--amber-primary)' }}>11.600 Kg / 8 m³</span>
                </div>
              </div>

              <div style={{ display: 'flex', justifyContent: 'space-between', padding: '0.75rem', background: 'rgba(0,0,0,0.3)', borderRadius: 'var(--radius-sm)' }}>
                <div>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)', display: 'block' }}>Asal Muatan</span>
                  <span style={{ fontWeight: 600, color: 'var(--text-main)' }}>Stockyard Curah Rumpin</span>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)', display: 'block' }}>Jembatan Timbang</span>
                  <span style={{ fontWeight: 600, color: 'var(--emerald-status)' }}>Auto-Weighbridge Valid</span>
                </div>
              </div>
            </div>

            <div style={{ marginTop: '1.25rem', paddingTop: '1rem', borderTop: '1px solid var(--border-subtle)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <a href="#kalkulator-curah" className="btn btn-primary btn-sm" style={{ width: '100%', textAlign: 'center' }}>
                Buka Kalkulator Kubikasi ke Tonase &rarr;
              </a>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
