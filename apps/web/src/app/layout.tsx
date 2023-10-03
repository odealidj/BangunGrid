import type { Metadata } from 'next';
import '@/styles/globals.css';

export const metadata: Metadata = {
  title: 'BangunGrid - Platform Bahan Bangunan & Logistik Berat',
  description: 'Solusi pengadaan material konstruksi agregat curah dan semen SNI langsung ke lokasi proyek dengan optimasi armada berat.',
  manifest: '/manifest.json',
  themeColor: '#070a12',
  icons: {
    icon: '/icons/icon-192x192.svg',
    apple: '/icons/icon-192x192.svg'
  }
};

export default function RootLayout({
  children
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="id">
      <head>
        <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=5" />
        <link rel="manifest" href="/manifest.json" />
        <link rel="apple-touch-icon" href="/icons/icon-192x192.svg" />
      </head>
      <body>{children}</body>
    </html>
  );
}
