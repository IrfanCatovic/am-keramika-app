/** Client-side cart snapshot. Checkout sends requested quantity only. */
export interface CartItem {
  productId: number;
  slug: string;
  name: string;
  imageUrl: string | null;
  unit: string;
  /** Quantity requested by the customer; kept as `quantity` for API compatibility. */
  quantity: number;
  requestedQuantity: number;
  actualQuantity: number;
  packageCount: number;
  packageQuantity: number;
  packagePrice?: number | null;
  saleByPackage: boolean;
  /** Last known display prices — refreshed from public API on cart page. */
  salePrice: number;
  effectiveSalePrice: number;
  isOnSale: boolean;
  discountPercent: number;
  categoryName?: string;
  groupName?: string;
}

export interface CartPersistedState {
  version: 2;
  items: CartItem[];
}

export type CartAddInput = Omit<
  CartItem,
  "quantity" | "requestedQuantity" | "actualQuantity" | "packageCount"
> & {
  quantity: number;
};
