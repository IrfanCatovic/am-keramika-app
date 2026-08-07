import { apiRequest } from "@/lib/api";
import { fetchProducts } from "@/lib/products-api";
import { getApiBusinessMessage } from "@/lib/categories-api";
import {
  AdjustStockPayload,
  AdjustStockResponse,
  InventoryMovementListParams,
  InventoryStockListParams,
  InventorySummary,
  PaginatedInventoryMovements,
  PaginatedLowStockResponse,
} from "@/types/inventory";
import { PaginatedProducts, Product } from "@/types/product";

export { getApiBusinessMessage };

function buildQuery(
  params: Record<string, string | number | boolean | undefined | null>,
): string {
  const searchParams = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === "") {
      continue;
    }
    if (typeof value === "boolean") {
      if (value) {
        searchParams.set(key, "true");
      }
      continue;
    }
    searchParams.set(key, String(value));
  }
  const query = searchParams.toString();
  return query ? `?${query}` : "";
}

function mapProductToInventoryRow(product: Product) {
  return {
    id: product.id,
    name: product.name,
    unit: product.unit,
    stockQuantity: product.stockQuantity,
    minStockQuantity: product.minStockQuantity,
    category: product.category
      ? { id: product.category.id, name: product.category.name }
      : null,
    group: product.group
      ? { id: product.group.id, name: product.group.name }
      : null,
    primaryImage: product.primaryImage,
  };
}

export async function fetchInventorySummary(): Promise<InventorySummary> {
  return apiRequest<InventorySummary>("/inventory/summary");
}

export async function fetchInventoryStock(
  params: InventoryStockListParams = {},
): Promise<PaginatedLowStockResponse> {
  const page = params.page ?? 1;
  const limit = params.limit ?? 20;
  const search = params.search?.trim() ?? "";
  const categoryID = params.categoryID;
  const groupID = params.groupID;
  const status = params.status ?? "all";

  if (status === "low") {
    return apiRequest<PaginatedLowStockResponse>(
      `/inventory/low-stock${buildQuery({
        page,
        limit,
        search,
        categoryID,
        groupID,
        excludeOutOfStock: true,
      })}`,
    );
  }

  const productsResponse: PaginatedProducts = await fetchProducts({
    page,
    limit,
    search,
    categoryID,
    groupID,
    stockStatus: status === "out" ? "out" : undefined,
  });

  return {
    products: productsResponse.products.map(mapProductToInventoryRow),
    pagination: productsResponse.pagination,
  };
}

export async function fetchInventoryMovements(
  params: InventoryMovementListParams = {},
): Promise<PaginatedInventoryMovements> {
  return apiRequest<PaginatedInventoryMovements>(
    `/inventory/movements${buildQuery({
      page: params.page,
      limit: params.limit,
      productID: params.productID,
      type: params.type,
      fromDate: params.fromDate,
      toDate: params.toDate,
    })}`,
  );
}

export async function adjustInventoryStock(
  payload: AdjustStockPayload,
): Promise<AdjustStockResponse> {
  const response = await apiRequest<{ data: AdjustStockResponse }>(
    "/inventory/adjust",
    {
      method: "POST",
      body: payload,
    },
  );
  return response.data;
}
