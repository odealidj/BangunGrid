'use client';

import React, { useState, useEffect } from 'react';
import Header from '@/components/Header';
import HeroBanner from '@/components/HeroBanner';
import BulkCalculator from '@/components/BulkCalculator';
import ProductCatalog from '@/components/ProductCatalog';
import WarehouseStockViewer from '@/components/WarehouseStockViewer';
import CartDrawer from '@/components/CartDrawer';
import ContractorPortalModal from '@/components/ContractorPortalModal';
import PwaInstallBanner from '@/components/PwaInstallBanner';
import Footer from '@/components/Footer';
import { apiService, FALLBACK_PRODUCTS } from '@/services/api';
import { Product, UOMConversion, CartItem, User } from '@/types';

export default function HomePage() {
  const [products, setProducts] = useState<Product[]>(FALLBACK_PRODUCTS);
  const [cartItems, setCartItems] = useState<CartItem[]>([]);
  const [isCartOpen, setIsCartOpen] = useState(false);
  const [isAuthOpen, setIsAuthOpen] = useState(false);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [canInstallPwa, setCanInstallPwa] = useState(false);
  const [triggerInstallPrompt, setTriggerInstallPrompt] = useState<(() => void) | null>(null);

  // Load products & check local auth token on mount
  useEffect(() => {
    async function init() {
      // 1. Load products
      try {
        const data = await apiService.getProducts();
        if (data && data.length > 0) {
          setProducts(data);
        }
      } catch (e) {
        console.warn('Failed to fetch products:', e);
      }

      // 2. Check token in localStorage
      const savedToken = localStorage.getItem('bangungrid_token');
      if (savedToken) {
        try {
          const user = await apiService.getProfile(savedToken);
          setCurrentUser(user);
        } catch {
          localStorage.removeItem('bangungrid_token');
        }
      }
    }
    init();
  }, []);

  // Add to cart handler
  const handleAddToCart = (product: Product, uom: UOMConversion, quantity: number) => {
    setCartItems((prev) => {
      const existingIndex = prev.findIndex(
        (item) => item.product.id === product.id && item.selected_uom.id === uom.id
      );

      const itemWeight = uom.weight_kg * quantity;
      const itemVolume = uom.volume_m3 * quantity;
      const itemPrice = uom.retail_price * quantity;

      if (existingIndex > -1) {
        const next = [...prev];
        const existing = next[existingIndex];
        const newQty = existing.quantity + quantity;
        next[existingIndex] = {
          ...existing,
          quantity: newQty,
          total_price: uom.retail_price * newQty,
          total_weight_kg: uom.weight_kg * newQty,
          total_volume_m3: uom.volume_m3 * newQty
        };
        return next;
      }

      const newItem: CartItem = {
        id: `${product.id}-${uom.id}-${Date.now()}`,
        product,
        selected_uom: uom,
        quantity,
        unit_price: uom.retail_price,
        total_price: itemPrice,
        total_weight_kg: itemWeight,
        total_volume_m3: itemVolume
      };

      return [...prev, newItem];
    });
  };

  const handleUpdateCartQty = (itemId: string, delta: number) => {
    setCartItems((prev) =>
      prev
        .map((item) => {
          if (item.id === itemId) {
            const nextQty = Math.max(1, item.quantity + delta);
            return {
              ...item,
              quantity: nextQty,
              total_price: item.selected_uom.retail_price * nextQty,
              total_weight_kg: item.selected_uom.weight_kg * nextQty,
              total_volume_m3: item.selected_uom.volume_m3 * nextQty
            };
          }
          return item;
        })
        .filter((item) => item.quantity > 0)
    );
  };

  const handleRemoveCartItem = (itemId: string) => {
    setCartItems((prev) => prev.filter((item) => item.id !== itemId));
  };

  const handleCheckoutSuccess = () => {
    setCartItems([]);
    setIsCartOpen(false);
  };

  const handleLogout = () => {
    localStorage.removeItem('bangungrid_token');
    setCurrentUser(null);
  };

  const totalWeightTon = cartItems.reduce((acc, item) => acc + item.total_weight_kg, 0) / 1000;
  const cartItemCount = cartItems.reduce((acc, item) => acc + item.quantity, 0);

  return (
    <>
      {/* PWA Floating / Top Install Notification */}
      <PwaInstallBanner
        onRegisterPrompt={(promptFn) => {
          setCanInstallPwa(true);
          setTriggerInstallPrompt(() => promptFn);
        }}
      />

      {/* Global Header */}
      <Header
        cartItemCount={cartItemCount}
        totalWeightTon={totalWeightTon}
        onOpenCart={() => setIsCartOpen(true)}
        onOpenAuth={() => setIsAuthOpen(true)}
        currentUser={currentUser}
        onLogout={handleLogout}
        canInstallPwa={canInstallPwa}
        onInstallPwa={() => triggerInstallPrompt && triggerInstallPrompt()}
      />

      <main>
        {/* Hero Section */}
        <HeroBanner onSearch={setSearchQuery} searchQuery={searchQuery} />

        {/* Bulk Materials Engine Calculator */}
        <BulkCalculator onAddToCart={handleAddToCart} />

        {/* Product Catalog with Multi-UOM */}
        <ProductCatalog
          products={products}
          onAddToCart={handleAddToCart}
          searchFilter={searchQuery}
        />

        {/* Available-To-Promise (ATP) Multi-Warehouse Stock Viewer */}
        <WarehouseStockViewer />
      </main>

      {/* Footer */}
      <Footer />

      {/* Cart Drawer with Fleet Logistics Recommendation Engine */}
      <CartDrawer
        isOpen={isCartOpen}
        onClose={() => setIsCartOpen(false)}
        items={cartItems}
        onUpdateQty={handleUpdateCartQty}
        onRemoveItem={handleRemoveCartItem}
        onCheckout={handleCheckoutSuccess}
      />

      {/* Contractor & Customer Portal Modal */}
      <ContractorPortalModal
        isOpen={isAuthOpen}
        onClose={() => setIsAuthOpen(false)}
        onSuccessAuth={(user) => setCurrentUser(user)}
      />
    </>
  );
}
