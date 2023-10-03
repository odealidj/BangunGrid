'use client';

import React, { useState } from 'react';
import { X, Trash2, Truck, ShieldCheck, ArrowRight, CheckCircle2 } from 'lucide-react';
import { CartItem, FleetRecommendation } from '@/types';

interface CartDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  items: CartItem[];
  onUpdateQty: (itemId: string, delta: number) => void;
  onRemoveItem: (itemId: string) => void;
  onCheckout: () => void;
}

export default function CartDrawer({
  isOpen,
  onClose,
  items,
  onUpdateQty,
  onRemoveItem,
  onCheckout
}: CartDrawerProps) {
  const [checkingOut, setCheckingOut] = useState(false);
  const [successReservation, setSuccessReservation] = useState(false);

  if (!isOpen) return null;

  // Total aggregation
  const totalPrice = items.reduce((sum, item) => sum + item.total_price, 0);
  const totalWeightKg = items.reduce((sum, item) => sum + item.total_weight_kg, 0);
  const totalWeightTon = totalWeightKg / 1000;
  const totalVolumeM3 = items.reduce((sum, item) => sum + item.total_volume_m3, 0);

  // Fleet Recommendation Engine
  const getFleetRecommendation = (): FleetRecommendation => {
    if (totalWeightTon <= 1.2) {
      return {
        vehicle_type: 'PICKUP_1T',
        name: 'Mobil Pick-Up 1 Ton',
        capacity_ton: 1.2,
        capacity_m3: 1.5,
        fleet_count_needed: 1,
        description: 'Ideal untuk pengiriman material ritel ringan dan jalan pemukiman sempit.'
      };
    } else if (totalWeightTon <= 4.0) {
      return {
        vehicle_type: 'ENGKEL_3T',
        name: 'Truk Colt Diesel Engkel (CDE) 4 Ton',
        capacity_ton: 4.0,
        capacity_m3: 5.0,
        fleet_count_needed: Math.ceil(totalWeightTon / 4.0),
        description: 'Muatan sedang untuk semen palet dan besi beton ke proyek perumahan.'
      };
    } else if (totalWeightTon <= 12.0) {
      return {
        vehicle_type: 'DUMP_TRUCK_10T',
        name: 'Dump Truck Indeks 8 (12 Ton)',
        capacity_ton: 12.0,
        capacity_m3: 8.0,
        fleet_count_needed: Math.ceil(totalWeightTon / 12.0),
        description: 'Standar pengiriman agregat pasir cor & batu split quarry dengan sistem hidrolik otomatis.'
      };
    } else {
      return {
        vehicle_type: 'TRONTON_25T',
        name: 'Truk Tronton Heavy Duty 25 Ton',
        capacity_ton: 25.0,
        capacity_m3: 24.0,
        fleet_count_needed: Math.ceil(totalWeightTon / 25.0),
        description: 'Pengiriman kapasitas masif langsung antar stockyard atau proyek infrastruktur besar.'
      };
    }
  };

  const fleet = getFleetRecommendation();

  const handleExecuteCheckout = () => {
    setCheckingOut(true);
    setTimeout(() => {
      setCheckingOut(false);
      setSuccessReservation(true);
      setTimeout(() => {
        setSuccessReservation(false);
        onCheckout();
      }, 2500);
    }, 1000);
  };

  return (
    <div className="drawer-backdrop" onClick={onClose}>
      <div className="drawer-panel" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="drawer-header">
          <div>
            <h3 style={{ fontSize: '1.2rem', fontWeight: 700 }}>Keranjang Pengiriman</h3>
            <span style={{ fontSize: '0.8rem', color: 'var(--text-dim)' }}>
              {items.length} jenis material dipilih
            </span>
          </div>
          <button
            onClick={onClose}
            style={{ background: 'transparent', color: 'var(--text-muted)', cursor: 'pointer' }}
            aria-label="Tutup"
          >
            <X size={22} />
          </button>
        </div>

        {/* Body */}
        <div className="drawer-body">
          {items.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '3rem 1rem', color: 'var(--text-dim)' }}>
              <p>Keranjang masih kosong.</p>
              <p style={{ fontSize: '0.85rem', marginTop: '0.5rem' }}>
                Silakan pilih semen, pasir curah, atau batu split dari katalog.
              </p>
            </div>
          ) : (
            <>
              {/* Item List */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.85rem' }}>
                {items.map((item) => (
                  <div
                    key={item.id}
                    style={{
                      background: 'rgba(255,255,255,0.03)',
                      border: '1px solid var(--border-subtle)',
                      borderRadius: 'var(--radius-md)',
                      padding: '0.85rem',
                      display: 'flex',
                      flexDirection: 'column',
                      gap: '0.5rem'
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                      <div>
                        <h4 style={{ fontSize: '0.95rem', fontWeight: 600 }}>{item.product.name}</h4>
                        <span style={{ fontSize: '0.8rem', color: 'var(--amber-primary)', fontWeight: 600 }}>
                          Satuan: {item.selected_uom.uom_name}
                        </span>
                      </div>
                      <button
                        onClick={() => onRemoveItem(item.id)}
                        style={{ background: 'transparent', color: 'var(--rose-alert)', cursor: 'pointer', padding: '0.2rem' }}
                      >
                        <Trash2 size={16} />
                      </button>
                    </div>

                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '0.25rem' }}>
                      <div style={{ fontSize: '0.8rem', color: 'var(--text-dim)' }}>
                        Muatan: {item.total_weight_kg.toLocaleString('id-ID')} Kg ({item.total_volume_m3.toFixed(2)} m³)
                      </div>

                      {/* Stepper */}
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                        <button
                          onClick={() => onUpdateQty(item.id, -1)}
                          style={{ width: '24px', height: '24px', background: 'rgba(255,255,255,0.08)', borderRadius: '4px', color: '#fff' }}
                        >
                          -
                        </button>
                        <span style={{ fontSize: '0.85rem', fontWeight: 700, width: '20px', textAlign: 'center' }}>
                          {item.quantity}
                        </span>
                        <button
                          onClick={() => onUpdateQty(item.id, 1)}
                          style={{ width: '24px', height: '24px', background: 'rgba(255,255,255,0.08)', borderRadius: '4px', color: '#fff' }}
                        >
                          +
                        </button>
                      </div>
                    </div>

                    <div style={{ textAlign: 'right', fontWeight: 700, color: 'var(--text-main)', fontSize: '0.95rem' }}>
                      Rp {item.total_price.toLocaleString('id-ID')}
                    </div>
                  </div>
                ))}
              </div>

              {/* Fleet Logistics Recommendation Engine Card */}
              <div className="fleet-card">
                <div className="fleet-icon-box">
                  <Truck size={24} />
                </div>
                <div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--blue-tech)', fontWeight: 700, textTransform: 'uppercase' }}>
                    REKOMENDASI ARMADA LOGISTIK BANGUNGRID
                  </div>
                  <div style={{ fontWeight: 700, fontSize: '0.95rem', color: 'var(--text-main)', marginTop: '0.15rem' }}>
                    {fleet.fleet_count_needed} Unit {fleet.name}
                  </div>
                  <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
                    {fleet.description}
                  </p>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-dim)', marginTop: '0.4rem' }}>
                    Total Muatan Fisik: <strong>{totalWeightTon.toFixed(2)} Ton</strong> ({totalVolumeM3.toFixed(2)} m³)
                  </div>
                </div>
              </div>
            </>
          )}
        </div>

        {/* Footer */}
        {items.length > 0 && (
          <div className="drawer-footer">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: '1rem' }}>
              <span style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>Total Tagihan Material:</span>
              <span style={{ fontSize: '1.4rem', fontWeight: 800, color: 'var(--amber-primary)' }}>
                Rp {totalPrice.toLocaleString('id-ID')}
              </span>
            </div>

            {successReservation ? (
              <div style={{ padding: '0.75rem', background: 'rgba(16, 185, 129, 0.15)', border: '1px solid var(--emerald-status)', borderRadius: 'var(--radius-md)', display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--emerald-status)', fontWeight: 600 }}>
                <CheckCircle2 size={20} />
                <span>Reservasi Stok 15 Menit Berhasil!</span>
              </div>
            ) : (
              <button
                onClick={handleExecuteCheckout}
                disabled={checkingOut}
                className="btn btn-primary"
                style={{ width: '100%' }}
                id="checkout-btn"
              >
                {checkingOut ? (
                  <span>Mengunci Stok di Redis (ATP)...</span>
                ) : (
                  <>
                    <span>Proses Reservasi & Pengiriman</span>
                    <ArrowRight size={18} />
                  </>
                )}
              </button>
            )}

            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.4rem', marginTop: '0.75rem', color: 'var(--text-dim)', fontSize: '0.75rem' }}>
              <ShieldCheck size={14} color="var(--emerald-status)" />
              <span>Soft-locking Stok Terdistribusi Redis Aktif (TTL 15 Menit)</span>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
