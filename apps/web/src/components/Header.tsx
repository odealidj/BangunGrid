'use client';

import React from 'react';
import { HardHat, ShoppingCart, ShieldCheck, UserCheck, Smartphone, Warehouse, Calculator } from 'lucide-react';
import { User } from '@/types';

interface HeaderProps {
  cartItemCount: number;
  totalWeightTon: number;
  onOpenCart: () => void;
  onOpenAuth: () => void;
  currentUser: User | null;
  onLogout: () => void;
  onInstallPwa?: () => void;
  canInstallPwa?: boolean;
}

export default function Header({
  cartItemCount,
  totalWeightTon,
  onOpenCart,
  onOpenAuth,
  currentUser,
  onLogout,
  onInstallPwa,
  canInstallPwa
}: HeaderProps) {
  return (
    <header className="header-glass">
      <div className="container header-content">
        {/* Brand Logo */}
        <a href="#beranda" className="brand-badge" id="brand-logo">
          <div className="brand-icon-box">
            <HardHat size={22} />
          </div>
          <div>
            Bangun<span className="brand-accent">Grid</span>
            <span style={{ fontSize: '0.65rem', display: 'block', color: 'var(--text-dim)', fontWeight: 500, letterSpacing: '0.05em' }}>
              HEAVY LOGISTICS & AGGREGATE
            </span>
          </div>
        </a>

        {/* Navigation Links */}
        <nav className="nav-links" style={{ display: 'none' }} id="desktop-nav">
          <a href="#katalog" className="nav-item">
            Katalog Material
          </a>
          <a href="#kalkulator-curah" className="nav-item">
            <Calculator size={16} /> Kalkulator Curah
          </a>
          <a href="#gudang-atp" className="nav-item">
            <Warehouse size={16} /> Stok Gudang (ATP)
          </a>
        </nav>

        {/* Actions */}
        <div className="header-actions">
          {/* Status Gateway Pill */}
          <div className="status-pill" title="KrakenD API Gateway aktif di port 8088">
            <span className="pulse-dot" />
            <span>Gateway :8088</span>
          </div>

          {/* PWA Install Button (jika tersedia prompt) */}
          {canInstallPwa && (
            <button
              onClick={onInstallPwa}
              className="btn btn-secondary btn-sm"
              id="install-pwa-btn"
              title="Instal aplikasi BangunGrid di smartphone atau desktop"
            >
              <Smartphone size={15} />
              <span>Instal PWA</span>
            </button>
          )}

          {/* User Account / Contractor Button */}
          {currentUser ? (
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <span style={{ fontSize: '0.85rem', color: 'var(--amber-primary)', fontWeight: 600 }}>
                {currentUser.email.split('@')[0]}
              </span>
              <button onClick={onLogout} className="btn btn-secondary btn-sm">
                Keluar
              </button>
            </div>
          ) : (
            <button
              onClick={onOpenAuth}
              className="btn btn-secondary btn-sm"
              id="auth-modal-btn"
            >
              <UserCheck size={16} />
              <span>Portal Kontraktor</span>
            </button>
          )}

          {/* Cart Drawer Trigger with Gross Weight */}
          <button
            onClick={onOpenCart}
            className="btn-cart"
            id="cart-drawer-trigger"
            aria-label="Buka Keranjang Pengiriman"
          >
            <ShoppingCart size={18} />
            {cartItemCount > 0 && (
              <span className="cart-count" id="cart-item-count">
                {cartItemCount}
              </span>
            )}
            {totalWeightTon > 0 && (
              <span style={{ fontSize: '0.75rem', color: 'var(--text-amber)', fontWeight: 700 }}>
                {totalWeightTon.toFixed(2)}T
              </span>
            )}
          </button>
        </div>
      </div>
    </header>
  );
}
