/** Client-side cart snapshot. Checkout must send only productId + quantity. */
export interface CartItem {
  productId: number;
  slug: string;
  name: string;
  imageUrl: string | null;
  unit: string;
  quantity: number;
  /** Last known display prices — refreshed from public API on cart page. */
  salePrice: number;
  effectiveSalePrice: number;
  isOnSale: boolean;
  discountPercent: number;
  categoryName?: string;
  groupName?: string;
}

export interface CartPersistedState {
  version: 1;
  items: CartItem[];
}

export type CartAddInput = Omit<CartItem, "quantity"> & {
  quantity: number;
};
