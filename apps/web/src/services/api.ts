import {
  Product,
  BulkDensityMatrix,
  BulkCalculationResult,
  Warehouse,
  ATPStock,
  User,
  AuthTokens
} from '@/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8088/api/v1';

// Seed cadangan jika database baru di-reset dan belum memiliki produk katalog
export const FALLBACK_PRODUCTS: Product[] = [
  {
    id: 'prod-001',
    tenant_id: 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    category_id: 'cat-semen',
    sku: 'SMN-TIGA-RODA-50',
    name: 'Semen Tiga Roda Portland Composite (PCC) 50 Kg',
    description: 'Semen hidrolis berkualitas standar SNI 7064:2014 untuk pengecoran struktur kokoh, plesteran, dan acian halus.',
    base_uom: 'SAK',
    is_bulk_material: false,
    image_url: 'https://images.unsplash.com/photo-1589939705384-5185137a7f0f?auto=format&fit=crop&w=600&q=80',
    uom_conversions: [
      {
        id: 'uom-01',
        product_id: 'prod-001',
        uom_code: 'SAK',
        uom_name: 'Sak (50 Kg)',
        conversion_factor: 1,
        weight_kg: 50,
        volume_m3: 0.035,
        retail_price: 68000,
        is_base_uom: true
      },
      {
        id: 'uom-02',
        product_id: 'prod-001',
        uom_code: 'PALET',
        uom_name: 'Palet (40 Sak)',
        conversion_factor: 40,
        weight_kg: 2000,
        volume_m3: 1.4,
        retail_price: 2640000, // Diskon grosir Rp 66.000 / sak
        is_base_uom: false
      },
      {
        id: 'uom-03',
        product_id: 'prod-001',
        uom_code: 'TRUCK_TON',
        uom_name: 'Delivery Coltt / 10 Ton',
        conversion_factor: 200,
        weight_kg: 10000,
        volume_m3: 7.0,
        retail_price: 12800000, // Rp 64.000 / sak
        is_base_uom: false
      }
    ]
  },
  {
    id: 'prod-002',
    tenant_id: 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    category_id: 'cat-curah',
    sku: 'AGR-PASIR-COR-01',
    name: 'Pasir Cor / Beton Kasar Gunung Galunggung (SNI)',
    description: 'Pasir silika alam bebas lumpur (<2%), gradasi butir kasar ideal untuk mutu beton K-250 hingga K-400.',
    base_uom: 'M3',
    is_bulk_material: true,
    bulk_material_code: 'BULK-SAND-02',
    image_url: 'https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=600&q=80',
    uom_conversions: [
      {
        id: 'uom-04',
        product_id: 'prod-002',
        uom_code: 'M3',
        uom_name: 'Meter Kubik (m³)',
        conversion_factor: 1,
        weight_kg: 1450,
        volume_m3: 1.0,
        retail_price: 295000,
        is_base_uom: true
      },
      {
        id: 'uom-05',
        product_id: 'prod-002',
        uom_code: 'DUMP_8M3',
        uom_name: 'Dump Truck Indeks 8 m³ (11.6 Ton)',
        conversion_factor: 8,
        weight_kg: 11600,
        volume_m3: 8.0,
        retail_price: 2280000,
        is_base_uom: false
      },
      {
        id: 'uom-06',
        product_id: 'prod-002',
        uom_code: 'TRONTON_24M3',
        uom_name: 'Tronton Muatan 24 m³ (34.8 Ton)',
        conversion_factor: 24,
        weight_kg: 34800,
        volume_m3: 24.0,
        retail_price: 6600000,
        is_base_uom: false
      }
    ]
  },
  {
    id: 'prod-003',
    tenant_id: 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    category_id: 'cat-curah',
    sku: 'AGR-BATU-SPLIT-12',
    name: 'Batu Split Cor 1-2 (10-20 mm) Quarry Rumpin',
    description: 'Batu andesit pecah mesin stone crusher, bebas debu lempung, bentuk kubikal optimum interlocking.',
    base_uom: 'M3',
    is_bulk_material: true,
    bulk_material_code: 'BULK-STN-01',
    image_url: 'https://images.unsplash.com/photo-1590069261209-f8e9b8642343?auto=format&fit=crop&w=600&q=80',
    uom_conversions: [
      {
        id: 'uom-07',
        product_id: 'prod-003',
        uom_code: 'M3',
        uom_name: 'Meter Kubik (m³)',
        conversion_factor: 1,
        weight_kg: 1350,
        volume_m3: 1.0,
        retail_price: 310000,
        is_base_uom: true
      },
      {
        id: 'uom-08',
        product_id: 'prod-003',
        uom_code: 'DUMP_8M3',
        uom_name: 'Dump Truck Indeks 8 m³ (10.8 Ton)',
        conversion_factor: 8,
        weight_kg: 10800,
        volume_m3: 8.0,
        retail_price: 2400000,
        is_base_uom: false
      }
    ]
  },
  {
    id: 'prod-004',
    tenant_id: 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    category_id: 'cat-besi',
    sku: 'BSI-ULIR-D13-SNI',
    name: 'Besi Beton Ulir D13 mm x 12m SNI TS280 / TS420B',
    description: 'Besi tulangan deformasi bersertifikat SNI 2052:2017 toleransi ketat, kuat tarik tinggi tahan gempa.',
    base_uom: 'BTG',
    is_bulk_material: false,
    image_url: 'https://images.unsplash.com/photo-1504917599217-d4dc5ebe6122?auto=format&fit=crop&w=600&q=80',
    uom_conversions: [
      {
        id: 'uom-09',
        product_id: 'prod-004',
        uom_code: 'BTG',
        uom_name: 'Batang 12m (12.5 Kg)',
        conversion_factor: 1,
        weight_kg: 12.5,
        volume_m3: 0.002,
        retail_price: 138000,
        is_base_uom: true
      },
      {
        id: 'uom-10',
        product_id: 'prod-004',
        uom_code: 'IKAT_80',
        uom_name: 'Ikat Pabrik (80 Batang / 1 Ton)',
        conversion_factor: 80,
        weight_kg: 1000,
        volume_m3: 0.16,
        retail_price: 10800000,
        is_base_uom: false
      }
    ]
  }
];

export const FALLBACK_WAREHOUSES: Warehouse[] = [
  {
    id: 'wh-01',
    tenant_id: 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    code: 'WH-BEKASI-CENTRAL',
    name: 'Central Warehouse & Hub Narogong',
    type: 'CENTRAL_WAREHOUSE',
    address: 'Jl. Raya Narogong KM 14.5, Cileungsi',
    city: 'Bekasi / Bogor',
    postal_code: '16820',
    latitude: -6.3752,
    longitude: 106.9681,
    is_active: true
  },
  {
    id: 'wh-02',
    tenant_id: 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    code: 'SY-RUMPIN-CURAH',
    name: 'Stockyard Agregat Curah Rumpin',
    type: 'STOCKYARD',
    address: 'Kawasan Tambang Quarry Rumpin, Parung Panjang',
    city: 'Bogor Barat',
    postal_code: '16630',
    latitude: -6.4215,
    longitude: 106.6128,
    is_active: true
  },
  {
    id: 'wh-03',
    tenant_id: 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    code: 'SILO-MARUNDA-01',
    name: 'Terminal Silo Semen Pelabuhan Marunda',
    type: 'SILO',
    address: 'Kawasan KBN Marunda Blok A3, Cilincing',
    city: 'Jakarta Utara',
    postal_code: '14150',
    latitude: -6.1084,
    longitude: 106.9532,
    is_active: true
  }
];

export const apiService = {
  // 1. Matriks Densitas SNI Material Curah
  async getBulkMatrices(): Promise<BulkDensityMatrix[]> {
    try {
      const res = await fetch(`${API_BASE_URL}/catalog/bulk/matrices`, {
        headers: { 'Accept': 'application/json' },
        cache: 'no-store'
      });
      if (res.ok) {
        const json = await res.json();
        if (json.data && Array.isArray(json.data) && json.data.length > 0) {
          return json.data;
        }
      }
    } catch (e) {
      console.warn('[API] Fallback bulk matrices:', e);
    }

    return [
      { id: '1', material_code: 'BULK-DUST-01', material_name: 'Abu Batu', standard_density_kg_per_m3: 1300, normal_tolerance_pct: 2.0 },
      { id: '2', material_code: 'BULK-FILL-01', material_name: 'Tanah Urug / Sirtu', standard_density_kg_per_m3: 1500, normal_tolerance_pct: 3.0 },
      { id: '3', material_code: 'BULK-SAND-01', material_name: 'Pasir Pasang Halus', standard_density_kg_per_m3: 1400, normal_tolerance_pct: 2.0 },
      { id: '4', material_code: 'BULK-SAND-02', material_name: 'Pasir Cor / Beton Kasar', standard_density_kg_per_m3: 1450, normal_tolerance_pct: 2.5 },
      { id: '5', material_code: 'BULK-STN-01', material_name: 'Batu Split 1-2 (10-20mm)', standard_density_kg_per_m3: 1350, normal_tolerance_pct: 1.5 },
      { id: '6', material_code: 'BULK-STN-02', material_name: 'Batu Split 2-3 (20-30mm)', standard_density_kg_per_m3: 1380, normal_tolerance_pct: 1.5 }
    ];
  },

  // 2. Kalkulator Material Curah (m³ <-> Tonase)
  async calculateBulk(materialCode: string, volumeM3: number): Promise<BulkCalculationResult> {
    try {
      const res = await fetch(`${API_BASE_URL}/catalog/bulk/calculate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify({ material_code: materialCode, volume_m3: volumeM3 })
      });
      if (res.ok) {
        const json = await res.json();
        if (json.data) {
          return json.data;
        }
      }
    } catch (e) {
      console.warn('[API] Fallback bulk calculation:', e);
    }

    // Perhitungan lokal berbasis rumus SNI jika backend offline
    const densities: Record<string, { name: string; density: number; tol: number }> = {
      'BULK-SAND-02': { name: 'Pasir Cor', density: 1450, tol: 2.5 },
      'BULK-SAND-01': { name: 'Pasir Pasang', density: 1400, tol: 2.0 },
      'BULK-STN-01': { name: 'Batu Split 1-2', density: 1350, tol: 1.5 },
      'BULK-STN-02': { name: 'Batu Split 2-3', density: 1380, tol: 1.5 },
      'BULK-DUST-01': { name: 'Abu Batu', density: 1300, tol: 2.0 },
      'BULK-FILL-01': { name: 'Tanah Urug', density: 1500, tol: 3.0 }
    };
    const ref = densities[materialCode] || { name: 'Material', density: 1400, tol: 2.0 };
    const expectedKg = volumeM3 * ref.density;
    const tolDelta = expectedKg * (ref.tol / 100);

    return {
      material_code: materialCode,
      volume_m3: volumeM3,
      expected_weight_kg: Math.round(expectedKg),
      expected_weight_ton: Number((expectedKg / 1000).toFixed(2)),
      min_tolerance_kg: Math.round(expectedKg - tolDelta),
      max_tolerance_kg: Math.round(expectedKg + tolDelta),
      standard_density_kg_per_m3: ref.density,
      tolerance_pct: ref.tol
    };
  },

  // 3. Daftar Produk Katalog Multi-UoM
  async getProducts(limit = 10): Promise<Product[]> {
    try {
      const res = await fetch(`${API_BASE_URL}/catalog/products?limit=${limit}`, {
        headers: { 'Accept': 'application/json' },
        cache: 'no-store'
      });
      if (res.ok) {
        const json = await res.json();
        if (json.data && Array.isArray(json.data) && json.data.length > 0) {
          return json.data;
        }
      }
    } catch (e) {
      console.warn('[API] Using fallback products:', e);
    }
    return FALLBACK_PRODUCTS;
  },

  // 4. Daftar Gudang & Stockyard
  async getWarehouses(): Promise<Warehouse[]> {
    try {
      const res = await fetch(`${API_BASE_URL}/inventory/warehouses`, {
        headers: { 'Accept': 'application/json' },
        cache: 'no-store'
      });
      if (res.ok) {
        const json = await res.json();
        if (json.data && Array.isArray(json.data) && json.data.length > 0) {
          return json.data;
        }
      }
    } catch (e) {
      console.warn('[API] Using fallback warehouses:', e);
    }
    return FALLBACK_WAREHOUSES;
  },

  // 5. Cek Stok ATP (Available-To-Promise)
  async getStockATP(warehouseId: string, productId: string): Promise<ATPStock> {
    try {
      const res = await fetch(`${API_BASE_URL}/inventory/stock/${warehouseId}/${productId}`, {
        headers: { 'Accept': 'application/json' },
        cache: 'no-store'
      });
      if (res.ok) {
        const json = await res.json();
        if (json.data) {
          return json.data;
        }
      }
    } catch (e) {
      console.warn('[API] Using simulated ATP stock:', e);
    }

    // Simulasi ATP Stock sehat
    return {
      warehouse_id: warehouseId,
      product_id: productId,
      physical_qty: 450,
      reserved_qty: 70,
      available_qty: 380,
      base_uom: 'SAK'
    };
  },

  // 6. Login Akun B2C / B2B Kontraktor
  async login(email: string, password: string):Promise<{ user: User; tokens: AuthTokens }> {
    const res = await fetch(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });
    const json = await res.json();
    if (!res.ok) {
      throw new Error(json.message || 'Gagal melakukan login');
    }
    return json.data;
  },

  // 7. Pendaftaran Kontraktor Proyek (B2B)
  async registerContractor(payload: {
    name: string;
    email: string;
    phone: string;
    password: string;
    company_name: string;
    tax_id: string;
    business_reg_num: string;
  }): Promise<{ user: User; tokens: AuthTokens }> {
    const res = await fetch(`${API_BASE_URL}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        tenant_id: 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
        name: payload.name,
        email: payload.email,
        phone: payload.phone,
        password: payload.password,
        user_type: 'CONTRACTOR_B2B',
        company_name: payload.company_name,
        tax_id: payload.tax_id,
        business_reg_num: payload.business_reg_num
      })
    });
    const json = await res.json();
    if (!res.ok) {
      throw new Error(json.message || 'Pendaftaran kontraktor gagal');
    }
    return json.data;
  },

  // 8. Ambil Profil Pengguna
  async getProfile(token: string): Promise<User> {
    const res = await fetch(`${API_BASE_URL}/auth/profile`, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Accept': 'application/json'
      }
    });
    const json = await res.json();
    if (!res.ok) {
      throw new Error(json.message || 'Token tidak valid');
    }
    return json.data.user;
  }
};
