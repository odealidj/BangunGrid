import React from 'react';
import { HardHat, ShieldCheck, MapPin, Phone, Mail } from 'lucide-react';

export default function Footer() {
  return (
    <footer className="footer-section">
      <div className="container">
        <div className="footer-grid">
          {/* Brand Col */}
          <div>
            <div className="brand-badge" style={{ marginBottom: '1rem' }}>
              <div className="brand-icon-box">
                <HardHat size={20} />
              </div>
              <span>Bangun<span className="brand-accent">Grid</span></span>
            </div>
            <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', lineHeight: '1.6' }}>
              Platform B2B & B2C terintegrasi untuk pengadaan material konstruksi berskala besar, agregat curah quarry, dan optimasi armada logistik berat berbasis PostGIS & Go Microservices.
            </p>
          </div>

          {/* Standards & Certifications */}
          <div>
            <h4 style={{ fontSize: '0.95rem', fontWeight: 700, marginBottom: '1rem', color: 'var(--text-main)' }}>
              Sertifikasi & Kepatuhan
            </h4>
            <ul style={{ listStyle: 'none', fontSize: '0.85rem', color: 'var(--text-muted)', display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              <li style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                <ShieldCheck size={16} color="var(--emerald-status)" />
                <span>SNI 7064:2014 (Semen PCC)</span>
              </li>
              <li style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                <ShieldCheck size={16} color="var(--emerald-status)" />
                <span>SNI 2052:2017 (Baja Tulangan Beton)</span>
              </li>
              <li style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                <ShieldCheck size={16} color="var(--emerald-status)" />
                <span>Metrologi Legal Jembatan Timbang</span>
              </li>
              <li style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                <ShieldCheck size={16} color="var(--blue-tech)" />
                <span>Kriptografi Token JWT Asimetris RSA-256</span>
              </li>
            </ul>
          </div>

          {/* Network Hubs */}
          <div>
            <h4 style={{ fontSize: '0.95rem', fontWeight: 700, marginBottom: '1rem', color: 'var(--text-main)' }}>
              Jaringan Gudang & Stockyard
            </h4>
            <ul style={{ listStyle: 'none', fontSize: '0.85rem', color: 'var(--text-muted)', display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              <li style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                <MapPin size={15} color="var(--amber-primary)" />
                <span>Central Hub Narogong, Bekasi</span>
              </li>
              <li style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                <MapPin size={15} color="var(--amber-primary)" />
                <span>Stockyard Agregat Curah Rumpin, Bogor</span>
              </li>
              <li style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                <MapPin size={15} color="var(--amber-primary)" />
                <span>Terminal Silo Marunda, Jakarta Utara</span>
              </li>
            </ul>
          </div>

          {/* Architecture Contact */}
          <div>
            <h4 style={{ fontSize: '0.95rem', fontWeight: 700, marginBottom: '1rem', color: 'var(--text-main)' }}>
              Infrastruktur Terdistribusi
            </h4>
            <p style={{ fontSize: '0.8rem', color: 'var(--text-dim)', marginBottom: '0.75rem' }}>
              Didukung oleh KrakenD API Gateway, Go Chi Hexagonal Microservices, PostgreSQL 16 PostGIS, Redis 7 Distributed Locking, dan NATS JetStream Event Broker.
            </p>
            <div style={{ fontSize: '0.8rem', color: 'var(--emerald-status)', fontWeight: 600 }}>
              &bull; Sistem Operasional 24/7 Siaga Proyek
            </div>
          </div>
        </div>

        <div style={{ paddingTop: '2rem', borderTop: '1px solid var(--border-subtle)', textAlign: 'center', fontSize: '0.8rem', color: 'var(--text-dim)' }}>
          &copy; 2023 BangunGrid Enterprise System. Hak Cipta Dilindungi Undang-Undang.
        </div>
      </div>
    </footer>
  );
}
