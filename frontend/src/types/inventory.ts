export interface ProductImage {
  id: number;
  url: string;
  isPrimary: boolean;
  sortOrder: number;
  width?: number;
  height?: number;
  format?: string;
}

export interface LowStockCategory {
  id: number;
  name: string;
}

export interface LowStockGroup {
  id: number;
  name: string;
}

export interface LowStockProduct {
  id: number;
  name: string;
  unit: string;
  stockQuantity: number;
  minStockQuantity: number;
  missingQuantity: number;
  category: LowStockCategory | null;
  group: LowStockGroup | null;
  primaryImage: ProductImage | null;
}

export interface LowStockPagination {
  page: number;
  limit: number;
  totalItems: number;
  totalPages: number;
}

export interface PaginatedLowStockResponse {
  products: LowStockProduct[];
  pagination: LowStockPagination;
}
