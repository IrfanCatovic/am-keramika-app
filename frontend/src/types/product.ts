export type PricingMode = "manual" | "calculated";

export interface ProductImage {
  id: number;
  url: string;
  isPrimary: boolean;
  sortOrder: number;
  width?: number | null;
  height?: number | null;
  format?: string;
}

export interface ProductCategorySummary {
  id: number;
  name: string;
  slug: string;
}

export interface ProductGroupSummary {
  id: number;
  name: string;
  slug: string;
}

export interface Product {
  id: number;
  name: string;
  slug: string;
  description: string;
  categoryID: number;
  category?: ProductCategorySummary;
  groupID: number | null;
  group?: ProductGroupSummary;
  unit: string;
  salePrice: number;
  stockQuantity: number;
  minStockQuantity: number;
  isActive: boolean;
  isOnSale: boolean;
  showOnHomepage: boolean;
  pricingMode: PricingMode;
  purchasePrice?: number;
  marginPercent?: number;
  vatPercent?: number;
  images?: ProductImage[];
  primaryImage: ProductImage | null;
}

export interface ProductPagination {
  page: number;
  limit: number;
  totalItems: number;
  totalPages: number;
}

export interface PaginatedProducts {
  products: Product[];
  pagination: ProductPagination;
}

export interface ProductListParams {
  page?: number;
  limit?: number;
  search?: string;
  categoryID?: number;
  groupID?: number;
  ungrouped?: boolean;
  includeInactive?: boolean;
  stockStatus?: "out";
}

export interface CreateProductPayload {
  name: string;
  categoryID: number;
  groupID?: number | null;
  unit: string;
  salePrice?: number;
  stockQuantity: number;
  minStockQuantity: number;
  description?: string;
  purchasePrice?: number;
  marginPercent?: number;
  vatPercent?: number;
  isOnSale?: boolean;
  showOnHomepage?: boolean;
}

export interface UpdateProductPayload {
  name: string;
  categoryID: number;
  groupID?: number | null;
  unit: string;
  salePrice?: number;
  stockQuantity: number;
  minStockQuantity: number;
  description?: string;
  purchasePrice?: number | null;
  marginPercent?: number | null;
  vatPercent?: number | null;
  isActive?: boolean;
  isOnSale?: boolean;
  showOnHomepage?: boolean;
}

export interface MessageResponse {
  message: string;
}
