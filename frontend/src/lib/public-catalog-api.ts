import type {
  PaginatedPublicProducts,
  PublicCategory,
  PublicProduct,
  PublicProductGroup,
  PublicProductListParams,
} from "@/types/public-catalog";

const API_URL = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ?? "";
export const PUBLIC_CATALOG_REVALIDATE_SECONDS = 60;

type PublicFetchOptions = {
  revalidate?: number;
};

export class PublicCatalogError extends Error {
  status: number;

  constructor(message: string, status = 500) {
    super(message);
    this.name = "PublicCatalogError";
    this.status = status;
  }
}

function buildQuery(params: PublicProductListParams = {}): string {
  const query = new URLSearchParams();
  if (params.page) query.set("page", String(params.page));
  if (params.limit) query.set("limit", String(params.limit));
  if (params.search?.trim()) query.set("search", params.search.trim());
  if (params.categoryID != null && params.categoryID !== "") {
    query.set("categoryID", String(params.categoryID));
  }
  if (params.categorySlug) query.set("categorySlug", params.categorySlug);
  if (params.groupID != null && params.groupID !== "") {
    query.set("groupID", String(params.groupID));
  }
  if (params.groupSlug) query.set("groupSlug", params.groupSlug);
  if (params.onSale) query.set("onSale", "true");
  if (params.homepage) query.set("homepage", "true");
  if (params.inStock) query.set("inStock", "true");
  if (params.random) query.set("random", "true");
  if (params.excludeId) query.set("excludeId", String(params.excludeId));
  if (params.sort) query.set("sort", params.sort);
  const qs = query.toString();
  return qs ? `?${qs}` : "";
}

async function publicFetch<T>(
  path: string,
  options: PublicFetchOptions = {},
): Promise<T> {
  if (!API_URL) {
    throw new PublicCatalogError("Kataloški servis trenutno nije dostupan.", 503);
  }

  let response: Response;
  try {
    const fetchOptions: RequestInit & { next?: { revalidate?: number } } = {
      method: "GET",
      headers: { Accept: "application/json" },
    };
    if (options.revalidate != null) {
      fetchOptions.cache = "force-cache";
      fetchOptions.next = { revalidate: options.revalidate };
    } else {
      fetchOptions.cache = "no-store";
    }
    response = await fetch(`${API_URL}${path}`, fetchOptions);
  } catch {
    throw new PublicCatalogError("Kataloški servis trenutno nije dostupan.", 503);
  }

  if (!response.ok) {
    let message = "Došlo je do greške. Pokušajte ponovo.";
    try {
      const body = (await response.json()) as { message?: string };
      if (body.message?.trim()) {
        message = body.message;
      }
    } catch {
      /* ignore */
    }
    throw new PublicCatalogError(message, response.status);
  }

  return (await response.json()) as T;
}

export async function fetchPublicProducts(
  params: PublicProductListParams = {},
  options: PublicFetchOptions = {},
): Promise<PaginatedPublicProducts> {
  return publicFetch<PaginatedPublicProducts>(
    `/public/products${buildQuery(params)}`,
    options,
  );
}

export async function fetchPublicProductBySlug(
  slug: string,
): Promise<PublicProduct> {
  return publicFetch<PublicProduct>(
    `/public/products/${encodeURIComponent(slug)}`,
  );
}

export async function fetchPublicCategories(
  options: PublicFetchOptions = {},
): Promise<PublicCategory[]> {
  return publicFetch<PublicCategory[]>("/public/categories", options);
}

export async function fetchPublicCategoryBySlug(
  slug: string,
): Promise<PublicCategory> {
  return publicFetch<PublicCategory>(
    `/public/categories/${encodeURIComponent(slug)}`,
  );
}

export async function fetchPublicProductGroups(params?: {
  categoryID?: number | string;
  categorySlug?: string;
}): Promise<PublicProductGroup[]> {
  const query = new URLSearchParams();
  if (params?.categoryID != null && params.categoryID !== "") {
    query.set("categoryID", String(params.categoryID));
  }
  if (params?.categorySlug) {
    query.set("categorySlug", params.categorySlug);
  }
  const qs = query.toString();
  return publicFetch<PublicProductGroup[]>(
    `/public/product-groups${qs ? `?${qs}` : ""}`,
  );
}

/** Safe helpers for Server Components — never throw during empty/offline states. */
export async function safeFetchPublicProducts(
  params: PublicProductListParams = {},
  options: PublicFetchOptions = {},
): Promise<PaginatedPublicProducts | null> {
  try {
    return await fetchPublicProducts(params, options);
  } catch {
    return null;
  }
}

export async function safeFetchPublicCategories(
  options: PublicFetchOptions = {},
): Promise<PublicCategory[]> {
  try {
    return await fetchPublicCategories(options);
  } catch {
    return [];
  }
}
