'use client';

import React, { useEffect, useState } from 'react';
import { Smartphone, Download, X } from 'lucide-react';

interface PwaInstallBannerProps {
  onRegisterPrompt?: (promptFn: () => void) => void;
}

export default function PwaInstallBanner({ onRegisterPrompt }: PwaInstallBannerProps) {
  const [deferredPrompt, setDeferredPrompt] = useState<any>(null);
  const [showBanner, setShowBanner] = useState<boolean>(false);

  useEffect(() => {
    // 1. Daftarkan Service Worker
    if (typeof window !== 'undefined' && 'serviceWorker' in navigator) {
      navigator.serviceWorker
        .register('/sw.js')
        .then((reg) => {
          console.log('[PWA] Service Worker aktif:', reg.scope);
        })
        .catch((err) => {
          console.warn('[PWA] Gagal mendaftarkan Service Worker:', err);
        });
    }

    // 2. Tangkap beforeinstallprompt
    const handleBeforeInstallPrompt = (e: Event) => {
      e.preventDefault();
      setDeferredPrompt(e);
      setShowBanner(true);
      if (onRegisterPrompt) {
        onRegisterPrompt(() => {
          (e as any).prompt();
        });
      }
    };

    window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt);

    return () => {
      window.removeEventListener('beforeinstallprompt', handleBeforeInstallPrompt);
    };
  }, [onRegisterPrompt]);

  const handleInstallClick = async () => {
    if (!deferredPrompt) return;
    deferredPrompt.prompt();
    const { outcome } = await deferredPrompt.userChoice;
    if (outcome === 'accepted') {
      setShowBanner(false);
    }
    setDeferredPrompt(null);
  };

  if (!showBanner) return null;

  return (
    <div className="pwa-banner" id="pwa-install-banner">
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.65rem' }}>
        <Smartphone size={18} color="var(--amber-primary)" />
        <span>
          <strong>Instal Aplikasi BangunGrid (PWA)</strong> untuk akses cepat tanpa browser & pemesanan material offline di lapangan proyek.
        </span>
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
        <button onClick={handleInstallClick} className="btn btn-primary btn-sm">
          <Download size={14} />
          <span>Pasang Sekarang</span>
        </button>
        <button
          onClick={() => setShowBanner(false)}
          style={{ background: 'transparent', color: 'var(--text-dim)', padding: '0.2rem' }}
          aria-label="Tutup Banner"
        >
          <X size={16} />
        </button>
      </div>
    </div>
  );
}
