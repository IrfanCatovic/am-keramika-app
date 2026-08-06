export interface Category {
  id: number;
  name: string;
  slug: string;
  isActive: boolean;
  createdAt: string;
}

export interface CategoryMutationResponse {
  message: string;
  data: Category;
}

export interface MessageResponse {
  message: string;
}
