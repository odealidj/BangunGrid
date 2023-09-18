# BangunGrid

> **BangunGrid** — Platform Ekosistem Penjualan Bahan Bangunan & Rantai Pasok Konstruksi Tingkat Enterprise.

---

## 🏗️ Gambaran Proyek

BangunGrid adalah platform distribusi dan penjualan material bangunan modern yang menghubungkan rantai pasok dari produsen/principal, distributor, kontraktor B2B, hingga pembeli ritel B2C secara terintegrasi.

### Pilar Utama:
- **Model Bisnis Hybrid**: Penjualan instan omnichannel ritel (B2C) & alur RFQ/Termin tender untuk kontraktor proyek (B2B).
- **Logistik Material Khusus**: Pengelolaan armada berat (Pick-up, Engkel, Dump Truck, Tronton) dengan sertifikasi Indeks Bak Armada dan verifikasi muatan digital (*e-POD*).
- **Material Curah (Pasir & Batu)**: Pendekatan *Dual-Measurement* ($m^3$ komersial + Tonase timbangan ODOL) dengan penanganan susut wajar.
- **Finansial & Kepatuhan Pajak**: Plafon kredit bertempo (TOP), integrasi e-Faktur Pajak PPN, dan deteksi keterlambatan piutang (*Aging AR*).

---

## 🛠️ Arsitektur Teknologi

- **Backend**: Go (Golang) Microservices dengan *Hexagonal Architecture (Ports and Adapters)* & Chi Router (`go-chi/chi/v5`).
- **API Gateway**: KrakenD Gateway (Stateless, High Performance, dengan jalur migrasi terencana ke Kong).
- **Event Bus & Messaging**: NATS & NATS JetStream (Pub/Sub & Persistent Event Streaming untuk Saga Pattern).
- **Basis Data**: PostgreSQL (Database-per-Service + PostGIS Spasial) & Redis (Distributed Locking & Idempotency).
- **Web Frontend**: React PWA (Kombinasi SSR untuk SEO & CSR untuk Portal B2B).
- **Mobile Apps**: Flutter (Pelanggan, Mandor B2B, Driver Logistik e-POD, dan POS Tablet Kasir Cabang).

---

## 📄 Lisensi
Hak Cipta © BangunGrid. Seluruh hak cipta dilindungi undang-undang.
