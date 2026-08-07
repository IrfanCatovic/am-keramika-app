import { ApiError, apiRequest } from "@/lib/api";
import { clearToken, getToken } from "@/lib/auth-token";
import { getApiBusinessMessage } from "@/lib/categories-api";
import {
  CreateProductPayload,
  MessageResponse,
  PaginatedProducts,
  Product,
  ProductImage,
  ProductListParams,
  UpdateProductPayload,
} from "@/types/product";

const API_URL = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ?? "";

export { getApiBusinessMessage };

function buildListQuery(params: ProductListParams): string {
  const searchParams = new URLSearchParams();
  if (params.page && params.page > 0) {
    searchParams.set("page", String(params.page));
  }
  if (params.limit && params.limit > 0) {
    searchParams.set("limit", String(params.limit));
  }
  if (params.search?.trim()) {
    searchParams.set("search", params.search.trim());
  }
  if (params.categoryID && params.categoryID > 0) {
    searchParams.set("categoryID", String(params.categoryID));
  }
  if (params.groupID && params.groupID > 0) {
    searchParams.set("groupID", String(params.groupID));
  }
  if (params.ungrouped) {
    searchParams.set("ungrouped", "true");
  }
  if (params.includeInactive) {
    searchParams.set("includeInactive", "true");
  }
  if (params.stockStatus) {
    searchParams.set("stockStatus", params.stockStatus);
  }
  const query = searchParams.toString();
  return query ? `?${query}` : "";
}

export async function fetchProducts(
  params: ProductListParams = {},
): Promise<PaginatedProducts> {
  return apiRequest<PaginatedProducts>(`/products${buildListQuery(params)}`);
}

export async function fetchProduct(id: number): Promise<Product> {
  return apiRequest<Product>(`/products/${id}`);
}

export async function createProduct(
  payload: CreateProductPayload,
): Promise<Product> {
  return apiRequest<Product>("/products", {
    method: "POST",
    body: payload,
  });
}

export async function updateProduct(
  id: number,
  payload: UpdateProductPayload,
): Promise<Product> {
  return apiRequest<Product>(`/products/${id}`, {
    method: "PUT",
    body: payload,
  });
}

export async function activateProduct(id: number): Promise<void> {
  await apiRequest<MessageResponse>(`/products/${id}/activate`, {
    method: "PUT",
  });
}

export async function deactivateProduct(id: number): Promise<void> {
  await apiRequest<MessageResponse>(`/products/${id}/deactivate`, {
    method: "PUT",
  });
}

export async function setPrimaryProductImage(
  productId: number,
  imageId: number,
): Promise<ProductImage> {
  const response = await apiRequest<{ data: ProductImage }>(
    `/products/${productId}/images/${imageId}/primary`,
    { method: "PUT" },
  );
  return response.data;
}

export async function reorderProductImages(
  productId: number,
  imageIDs: number[],
): Promise<ProductImage[]> {
  const response = await apiRequest<{ data: ProductImage[] }>(
    `/products/${productId}/images/reorder`,
    {
      method: "PUT",
      body: { imageIDs },
    },
  );
  return response.data ?? [];
}

export async function deleteProductImage(
  productId: number,
  imageId: number,
): Promise<void> {
  await apiRequest<MessageResponse>(
    `/products/${productId}/images/${imageId}`,
    { method: "DELETE" },
  );
}

/**
 * Multipart upload — apiRequest je JSON-only.
 */
export async function uploadProductImages(
  productId: number,
  files: File[],
): Promise<ProductImage[]> {
  if (!API_URL) {
    throw new ApiError(
      "API adresa nije podešena. Proverite .env.local datoteku.",
      500,
    );
  }
  if (files.length === 0) {
    return [];
  }

  const formData = new FormData();
  for (const file of files) {
    formData.append("images", file);
  }

  const headers = new Headers({ Accept: "application/json" });
  const token = getToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${API_URL}/products/${productId}/images`, {
    method: "POST",
    headers,
    body: formData,
  });

  const rawText = await response.text();
  let payload: unknown = null;
  if (rawText) {
    try {
      payload = JSON.parse(rawText);
    } catch {
      payload = rawText;
    }
  }

  if (!response.ok) {
    if (response.status === 401) {
      clearToken();
    }
    const fallback =
      response.status === 401
        ? "Sesija je istekla. Prijavite se ponovo."
        : response.status === 403
          ? "Nemate dozvolu za ovu akciju."
          : "Greška pri uploadu slika.";
    throw new ApiError(
      getApiBusinessMessage(
        new ApiError(fallback, response.status, payload),
        fallback,
      ),
      response.status,
      payload,
    );
  }

  if (payload && typeof payload === "object" && "data" in payload) {
    const data = (payload as { data: ProductImage[] }).data;
    return Array.isArray(data) ? data : [];
  }
  return [];
}
