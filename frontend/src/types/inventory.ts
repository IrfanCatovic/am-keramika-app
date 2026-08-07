export type InventoryStockStatus = "all" | "low" | "out";

export type InventoryTab = "stock" | "history";

export interface ProductImage {
  id: number;
  url: string;
  isPrimary: boolean;
  sortOrder: number;
  width?: number | null;
  height?: number | null;
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

export interface InventoryProductRow {
  id: number;
  name: string;
  unit: string;
  stockQuantity: number;
  minStockQuantity: number;
  missingQuantity?: number;
  category: LowStockCategory | null;
  group: LowStockGroup | null;
  primaryImage: ProductImage | null;
}

export type LowStockProduct = InventoryProductRow;

export interface InventoryPagination {
  page: number;
  limit: number;
  totalItems: number;
  totalPages: number;
}

export interface PaginatedLowStockResponse {
  products: InventoryProductRow[];
  pagination: InventoryPagination;
}

export interface InventorySummary {
  lowStockCount: number;
  outOfStockCount: number;
}

export interface AdjustStockPayload {
  productID: number;
  newQuantity: number;
  note?: string;
}

export interface AdjustStockResponse {
  productID: number;
  previousStock: number;
  newStock: number;
  change: number;
  movementID: number;
}

export interface InventoryMovementUser {
  id: number;
  username: string;
}

export interface InventoryMovement {
  id: number;
  productID: number;
  productName: string;
  productUnit: string;
  type: "in" | "adjust" | "sale" | "return" | string;
  quantity: number;
  note?: string;
  createdAt: string;
  createdByUser?: InventoryMovementUser | null;
}

export interface PaginatedInventoryMovements {
  movements: InventoryMovement[];
  pagination: InventoryPagination;
}

export interface InventoryStockListParams {
  page?: number;
  limit?: number;
  search?: string;
  categoryID?: number;
  groupID?: number;
  status?: InventoryStockStatus;
}

export interface InventoryMovementListParams {
  page?: number;
  limit?: number;
  productID?: number;
  type?: string;
  fromDate?: string;
  toDate?: string;
}
