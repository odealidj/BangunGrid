'use client';

import React, { useState } from 'react';
import { ShoppingCart, Check, Layers, Info } from 'lucide-react';
import { Product, UOMConversion } from '@/types';

interface ProductCatalogProps {
  products: Product[];
  onAddToCart: (product: Product, uom: UOMConversion, quantity: number) => void;
  searchFilter: string;
}

export default function ProductCatalog({ products, onAddToCart, searchFilter }: ProductCatalogProps) {
  const [selectedUoms, setSelectedUoms] = useState<Record<string, string>>({});
  const [quantities, setQuantities] = useState<Record<string, number>>({});
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [justAdded, setJustAdded] = useState<Record<string, boolean>>({});

  // Filter products by category & search query
  const filteredProducts = products.filter((p) => {
    const matchesSearch =
      p.name.toLowerCase().includes(searchFilter.toLowerCase()) ||
      p.description.toLowerCase().includes(searchFilter.toLowerCase()) ||
      p.sku.toLowerCase().includes(searchFilter.toLowerCase());

    if (!matchesSearch) return false;
    if (selectedCategory === 'all') return true;
    if (selectedCategory === 'semen' && p.category_id.includes('semen')) return true;
    if (selectedCategory === 'curah' && p.is_bulk_material) return true;
    if (selectedCategory === 'besi' && p.category_id.includes('besi')) return true;
    return true;
  });

  const getActiveUom = (product: Product): UOMConversion => {
    const activeId = selectedUoms[product.id];
    if (activeId) {
      const found = product.uom_conversions.find((u) => u.id === activeId);
      if (found) return found;
    }
    // Default base UOM
    return product.uom_conversions[0];
  };

  const handleSelectUom = (productId: string, uomId: string) => {
    setSelectedUoms((prev) => ({ ...prev, [productId]: uomId }));
  };

  const handleQtyChange = (productId: string, delta: number) => {
    const current = quantities[productId] || 1;
    const next = Math.max(1, current + delta);
    setQuantities((prev) => ({ ...prev, [productId]: next }));
  };

  const handleAddToCart = (product: Product) => {
    const uom = getActiveUom(product);
    const qty = quantities[product.id] || 1;
    onAddToCart(product, uom, qty);

    setJustAdded((prev) => ({ ...prev, [product.id]: true }));
    setTimeout(() => {
      setJustAdded((prev) => ({ ...prev, [product.id]: false }));
    }, 1500);
  };

  return (
    <section className="container" id="katalog" style={{ margin: '4rem auto' }}>
      <div className="section-header">
        <span className="section-tag">KATALOG MATERIAL KONSTRUKSI</span>
        <h2 className="section-title">Solusi Material Multi-Satuan (Multi-UoM)</h2>
        <p className="section-subtitle">
          Pesan satuan eceran harian (Sak/Batang) atau beralih ke satuan borongan proyek (Palet/Tonase) dengan
          potongan harga grosir otomatis untuk kontraktor.
        </p>
      </div>

      {/* Category Pills */}
      <div className="tab-group" style={{ maxWidth: '480px', marginBottom: '2rem' }}>
        <button
          className={`tab-btn ${selectedCategory === 'all' ? 'active' : ''}`}
          onClick={() => setSelectedCategory('all')}
        >
          Semua Material
        </button>
        <button
          className={`tab-btn ${selectedCategory === 'semen' ? 'active' : ''}`}
          onClick={() => setSelectedCategory('semen')}
        >
          Semen SNI
        </button>
        <button
          className={`tab-btn ${selectedCategory === 'curah' ? 'active' : ''}`}
          onClick={() => setSelectedCategory('curah')}
        >
          Agregat Curah
        </button>
        <button
          className={`tab-btn ${selectedCategory === 'besi' ? 'active' : ''}`}
          onClick={() => setSelectedCategory('besi')}
        >
          Besi Beton
        </button>
      </div>

      {/* Grid */}
      <div className="product-grid">
        {filteredProducts.map((product) => {
          const activeUom = getActiveUom(product);
          const currentQty = quantities[product.id] || 1;
          const totalPrice = activeUom.retail_price * currentQty;
          const isAdded = justAdded[product.id];

          return (
            <div key={product.id} className="glass-card product-card">
              <div className="product-media">
                {product.image_url && (
                  <img src={product.image_url} alt={product.name} className="product-img" loading="lazy" />
                )}
                <span className="product-badge">
                  {product.is_bulk_material ? 'AGREGAT CURAH' : 'STANDAR SNI'}
                </span>
              </div>

              <h3 className="product-title">{product.name}</h3>
              <p className="product-desc">{product.description}</p>

              {/* Multi-UOM Selector */}
              <div style={{ marginBottom: '0.5rem' }}>
                <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)', fontWeight: 600, display: 'block', marginBottom: '0.35rem' }}>
                  PILIH SATUAN PEMBELIAN (MULTI-UOM):
                </span>
                <div className="uom-pill-group">
                  {product.uom_conversions.map((uom) => (
                    <button
                      key={uom.id}
                      type="button"
                      className={`uom-pill ${activeUom.id === uom.id ? 'active' : ''}`}
                      onClick={() => handleSelectUom(product.id, uom.id)}
                    >
                      {uom.uom_name}
                    </button>
                  ))}
                </div>
              </div>

              {/* Physical specs summary */}
              <div style={{ fontSize: '0.75rem', color: 'var(--text-dim)', marginBottom: '1rem', background: 'rgba(0,0,0,0.25)', padding: '0.5rem', borderRadius: 'var(--radius-sm)' }}>
                <span>Bobot Satuan: <strong>{activeUom.weight_kg.toLocaleString('id-ID')} Kg</strong></span>
                <span style={{ margin: '0 0.5rem' }}>&bull;</span>
                <span>Volume: <strong>{activeUom.volume_m3} m³</strong></span>
              </div>

              {/* Price & Quantity */}
              <div className="price-row">
                <div>
                  <div className="price-amount">
                    Rp {totalPrice.toLocaleString('id-ID')}
                  </div>
                  <div className="price-sub">
                    Rp {activeUom.retail_price.toLocaleString('id-ID')} / {activeUom.uom_code}
                  </div>
                </div>

                {/* Qty Stepper */}
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                  <button
                    onClick={() => handleQtyChange(product.id, -1)}
                    style={{ width: '28px', height: '28px', background: 'rgba(255,255,255,0.08)', borderRadius: '4px', color: '#fff', fontWeight: 'bold' }}
                  >
                    -
                  </button>
                  <span style={{ width: '28px', textAlign: 'center', fontWeight: 'bold', fontSize: '0.9rem' }}>
                    {currentQty}
                  </span>
                  <button
                    onClick={() => handleQtyChange(product.id, 1)}
                    style={{ width: '28px', height: '28px', background: 'rgba(255,255,255,0.08)', borderRadius: '4px', color: '#fff', fontWeight: 'bold' }}
                  >
                    +
                  </button>
                </div>
              </div>

              {/* Add to Cart Button */}
              <button
                onClick={() => handleAddToCart(product)}
                className={`btn ${isAdded ? 'btn-secondary' : 'btn-primary'}`}
                style={{ width: '100%', marginTop: 'auto' }}
              >
                {isAdded ? (
                  <>
                    <Check size={16} color="var(--emerald-status)" />
                    <span>Tersimpan di Keranjang!</span>
                  </>
                ) : (
                  <>
                    <ShoppingCart size={16} />
                    <span>Tambah {currentQty} {activeUom.uom_code}</span>
                  </>
                )}
              </button>
            </div>
          );
        })}
      </div>
    </section>
  );
}
