import { ApiError, apiRequest } from "@/lib/api";
import {
  Category,
  CategoryMutationResponse,
  MessageResponse,
} from "@/types/category";
import {
  ProductGroup,
  ProductGroupListResponse,
  ProductGroupMutationResponse,
} from "@/types/product-group";

/** Preferira poslovnu `error` poruku iz backend body-ja (npr. 409 konflikti). */
export function getApiBusinessMessage(
  error: unknown,
  fallback: string,
): string {
  if (error instanceof ApiError) {
    if (error.body && typeof error.body === "object") {
      const body = error.body as Record<string, unknown>;
      if (typeof body.error === "string" && body.error.trim()) {
        return body.error.trim();
      }
      if (typeof body.message === "string" && body.message.trim()) {
        return body.message.trim();
      }
    }
    if (error.message.trim()) {
      return error.message;
    }
  }
  if (error instanceof Error && error.message.trim()) {
    return error.message;
  }
  return fallback;
}

export async function fetchCategories(
  includeInactive = true,
): Promise<Category[]> {
  const params = new URLSearchParams({
    includeInactive: includeInactive ? "true" : "false",
  });
  return apiRequest<Category[]>(`/categories?${params}`);
}

export async function createCategory(name: string): Promise<Category> {
  const response = await apiRequest<CategoryMutationResponse>("/categories", {
    method: "POST",
    body: { name },
  });
  return response.data;
}

export async function updateCategory(
  id: number,
  name: string,
): Promise<Category> {
  const response = await apiRequest<CategoryMutationResponse>(
    `/categories/${id}`,
    {
      method: "PUT",
      body: { name },
    },
  );
  return response.data;
}

export async function updateCategoryStatus(
  id: number,
  isActive: boolean,
): Promise<Category> {
  const response = await apiRequest<CategoryMutationResponse>(
    `/categories/${id}/status`,
    {
      method: "PUT",
      body: { isActive },
    },
  );
  return response.data;
}

export async function deleteCategory(id: number): Promise<void> {
  await apiRequest<MessageResponse>(`/categories/${id}`, {
    method: "DELETE",
  });
}

export async function fetchProductGroups(
  categoryID: number,
): Promise<ProductGroup[]> {
  const params = new URLSearchParams({
    categoryID: String(categoryID),
  });
  const response = await apiRequest<ProductGroupListResponse>(
    `/product-groups?${params}`,
  );
  return response.data ?? [];
}

export async function createProductGroup(
  name: string,
  categoryID: number,
): Promise<ProductGroup> {
  const response = await apiRequest<ProductGroupMutationResponse>(
    "/product-groups",
    {
      method: "POST",
      body: { name, categoryID },
    },
  );
  return response.data;
}

export async function updateProductGroup(
  id: number,
  name: string,
  categoryID: number,
): Promise<ProductGroup> {
  const response = await apiRequest<ProductGroupMutationResponse>(
    `/product-groups/${id}`,
    {
      method: "PUT",
      body: { name, categoryID },
    },
  );
  return response.data;
}

export async function deleteProductGroup(id: number): Promise<void> {
  await apiRequest<MessageResponse>(`/product-groups/${id}`, {
    method: "DELETE",
  });
}
