// Domain Models & API Types for BangunGrid Web PWA

export interface UOMConversion {
  id: string;
  product_id: string;
  uom_code: string;         // misal: 'SAK', 'PALET', 'TON', 'M3'
  uom_name: string;         // misal: 'Sak 50kg', 'Palet (40 Sak)', 'Ton'
  conversion_factor: number;// Pengali ke Base UOM (misal: 1, 40, 25)
  weight_kg: number;
  volume_m3: number;
  retail_price: number;
  is_base_uom: boolean;
}

export interface PriceTier {
  id: string;
  min_quantity: number;
  discounted_price: number;
}

export interface Product {
  id: string;
  tenant_id: string;
  category_id: string;
  sku: string;
  name: string;
  description: string;
  base_uom: string;
  is_bulk_material: boolean;
  bulk_material_code?: string;
  image_url?: string;
  uom_conversions: UOMConversion[];
  price_tiers?: PriceTier[];
}

export interface BulkDensityMatrix {
  id: string;
  material_code: string;
  material_name: string;
  standard_density_kg_per_m3: number;
  normal_tolerance_pct: number;
}

export interface BulkCalculationResult {
  material_code: string;
  volume_m3: number;
  expected_weight_kg: number;
  expected_weight_ton: number;
  min_tolerance_kg: number;
  max_tolerance_kg: number;
  standard_density_kg_per_m3: number;
  tolerance_pct: number;
}

export interface Warehouse {
  id: string;
  tenant_id: string;
  code: string;
  name: string;
  type: 'CENTRAL_WAREHOUSE' | 'STOCKYARD' | 'SILO' | 'BRANCH_STORE';
  address: string;
  city: string;
  postal_code: string;
  latitude: number;
  longitude: number;
  is_active: boolean;
}

export interface ATPStock {
  warehouse_id: string;
  product_id: string;
  physical_qty: number;
  reserved_qty: number;
  available_qty: number;    // Available-To-Promise
  base_uom: string;
}

export interface User {
  id: string;
  tenant_id: string;
  email: string;
  phone: string;
  user_type: 'CUSTOMER_B2C' | 'CONTRACTOR_B2B' | 'DRIVER' | 'CASHIER' | 'ADMIN';
  role: string;
  is_active: boolean;
  created_at: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

export interface CartItem {
  id: string;
  product: Product;
  selected_uom: UOMConversion;
  quantity: number;
  unit_price: number;
  total_price: number;
  total_weight_kg: number;
  total_volume_m3: number;
}

export interface FleetRecommendation {
  vehicle_type: 'PICKUP_1T' | 'ENGKEL_3T' | 'DUMP_TRUCK_10T' | 'TRONTON_25T';
  name: string;
  capacity_ton: number;
  capacity_m3: number;
  fleet_count_needed: number;
  description: string;
}
