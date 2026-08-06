export interface ProductGroup {
  id: number;
  name: string;
  slug: string;
  categoryID: number;
}

export interface ProductGroupListResponse {
  data: ProductGroup[];
}

export interface ProductGroupMutationResponse {
  message: string;
  data: ProductGroup;
}
