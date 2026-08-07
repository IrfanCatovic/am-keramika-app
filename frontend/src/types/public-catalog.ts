export interface PublicProductImage {
  id: number;
  url: string;
  isPrimary: boolean;
  sortOrder: number;
  width?: number | null;
  height?: number | null;
  format?: string;
}

export interface PublicCategorySummary {
  id: number;
  name: string;
  slug: string;
}

export interface PublicGroupSummary {
  id: number;
  name: string;
  slug: string;
}

export interface PublicProduct {
  id: number;
  name: string;
  slug: string;
  description: string;
  category?: PublicCategorySummary;
  group?: PublicGroupSummary;
  unit: string;
  salePrice: number;
  effectiveSalePrice: number;
  isOnSale: boolean;
  discountPercent: number;
  inStock: boolean;
  showOnHomepage: boolean;
  images?: PublicProductImage[];
  primaryImage: PublicProductImage | null;
}

export interface PublicCategory {
  id: number;
  name: string;
  slug: string;
}

export interface PublicProductGroup {
  id: number;
  name: string;
  slug: string;
  categoryID: number;
}

export interface PublicPagination {
  page: number;
  limit: number;
  totalItems: number;
  totalPages: number;
}

export interface PaginatedPublicProducts {
  products: PublicProduct[];
  pagination: PublicPagination;
}

export type PublicProductSort =
  | "recommended"
  | "price_asc"
  | "price_desc"
  | "name_asc"
  | "name_desc";

export interface PublicProductListParams {
  page?: number;
  limit?: number;
  search?: string;
  categoryID?: number | string;
  categorySlug?: string;
  groupID?: number | string;
  groupSlug?: string;
  onSale?: boolean;
  homepage?: boolean;
  inStock?: boolean;
  random?: boolean;
  excludeId?: number;
  sort?: PublicProductSort | string;
}
