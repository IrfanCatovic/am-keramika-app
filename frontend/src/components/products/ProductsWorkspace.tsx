"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import {
  ProductFilterValues,
  ProductFilters,
} from "@/components/products/ProductFilters";
import { ProductList } from "@/components/products/ProductList";
import { ConfirmDialog } from "@/components/ui/ConfirmDialog";
import {
  fetchCategories,
  fetchProductGroups,
} from "@/lib/categories-api";
import {
  activateProduct,
  deactivateProduct,
  fetchProducts,
  getApiBusinessMessage,
} from "@/lib/products-api";
import { Category } from "@/types/category";
import { Product, ProductPagination } from "@/types/product";
import { ProductGroup } from "@/types/product-group";

function parsePositiveInt(value: string | null): number | null {
  if (!value) {
    return null;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

export function ProductsWorkspace() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const page = parsePositiveInt(searchParams.get("page")) ?? 1;
  const categoryID = parsePositiveInt(searchParams.get("categoryID"));
  const groupID = parsePositiveInt(searchParams.get("groupID"));
  const ungrouped = searchParams.get("ungrouped") === "true";
  const includeInactive = searchParams.get("includeInactive") === "true";
  const searchFromUrl = searchParams.get("search") ?? "";

  const [searchInput, setSearchInput] = useState(searchFromUrl);
  const [categories, setCategories] = useState<Category[]>([]);
  const [groups, setGroups] = useState<ProductGroup[]>([]);
  const [groupsLoading, setGroupsLoading] = useState(false);

  const [products, setProducts] = useState<Product[]>([]);
  const [pagination, setPagination] = useState<ProductPagination | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<number | null>(null);

  const [confirm, setConfirm] = useState<
    | { open: false }
    | { open: true; product: Product }
  >({ open: false });
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [confirmError, setConfirmError] = useState<string | null>(null);

  const filterValues: ProductFilterValues = useMemo(
    () => ({
      search: searchInput,
      categoryID,
      groupID,
      ungrouped,
      includeInactive,
    }),
    [searchInput, categoryID, groupID, ungrouped, includeInactive],
  );

  const syncQuery = useCallback(
    (patch: {
      page?: number;
      search?: string;
      categoryID?: number | null;
      groupID?: number | null;
      ungrouped?: boolean;
      includeInactive?: boolean;
    }) => {
      const params = new URLSearchParams(searchParams.toString());

      const nextPage = patch.page ?? page;
      if (nextPage > 1) {
        params.set("page", String(nextPage));
      } else {
        params.delete("page");
      }

      const nextSearch =
        patch.search !== undefined ? patch.search : searchFromUrl;
      if (nextSearch.trim()) {
        params.set("search", nextSearch.trim());
      } else {
        params.delete("search");
      }

      const nextCategory =
        patch.categoryID !== undefined ? patch.categoryID : categoryID;
      if (nextCategory) {
        params.set("categoryID", String(nextCategory));
      } else {
        params.delete("categoryID");
      }

      const nextUngrouped =
        patch.ungrouped !== undefined ? patch.ungrouped : ungrouped;
      const nextGroup =
        patch.groupID !== undefined ? patch.groupID : groupID;

      if (nextUngrouped) {
        params.set("ungrouped", "true");
        params.delete("groupID");
      } else {
        params.delete("ungrouped");
        if (nextGroup) {
          params.set("groupID", String(nextGroup));
        } else {
          params.delete("groupID");
        }
      }

      const nextInactive =
        patch.includeInactive !== undefined
          ? patch.includeInactive
          : includeInactive;
      if (nextInactive) {
        params.set("includeInactive", "true");
      } else {
        params.delete("includeInactive");
      }

      const query = params.toString();
      router.replace(query ? `${pathname}?${query}` : pathname, {
        scroll: false,
      });
    },
    [
      searchParams,
      page,
      searchFromUrl,
      categoryID,
      groupID,
      ungrouped,
      includeInactive,
      pathname,
      router,
    ],
  );

  useEffect(() => {
    const timer = window.setTimeout(() => {
      setSearchInput(searchFromUrl);
    }, 0);
    return () => window.clearTimeout(timer);
  }, [searchFromUrl]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      if (searchInput === searchFromUrl) {
        return;
      }
      syncQuery({ search: searchInput, page: 1 });
    }, 300);
    return () => window.clearTimeout(timer);
  }, [searchInput, searchFromUrl, syncQuery]);

  useEffect(() => {
    let cancelled = false;
    const timer = window.setTimeout(() => {
      void (async () => {
        try {
          const data = await fetchCategories(true);
          if (!cancelled) {
            setCategories(data);
          }
        } catch {
          if (!cancelled) {
            setCategories([]);
          }
        }
      })();
    }, 0);
    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, []);

  useEffect(() => {
    let cancelled = false;
    const timer = window.setTimeout(() => {
      if (!categoryID) {
        if (!cancelled) {
          setGroups([]);
          setGroupsLoading(false);
        }
        return;
      }
      setGroupsLoading(true);
      void (async () => {
        try {
          const data = await fetchProductGroups(categoryID);
          if (!cancelled) {
            setGroups(data);
          }
        } catch {
          if (!cancelled) {
            setGroups([]);
          }
        } finally {
          if (!cancelled) {
            setGroupsLoading(false);
          }
        }
      })();
    }, 0);
    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [categoryID]);

  const loadProducts = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await fetchProducts({
        page,
        limit: 20,
        search: searchFromUrl || undefined,
        categoryID: categoryID ?? undefined,
        groupID: ungrouped ? undefined : (groupID ?? undefined),
        ungrouped,
        includeInactive,
      });
      setProducts(result.products ?? []);
      setPagination(result.pagination);
    } catch (err) {
      setProducts([]);
      setPagination(null);
      setError(getApiBusinessMessage(err, "Nije moguće učitati proizvode."));
    } finally {
      setLoading(false);
    }
  }, [
    page,
    searchFromUrl,
    categoryID,
    groupID,
    ungrouped,
    includeInactive,
  ]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadProducts();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [loadProducts]);

  async function handleConfirmToggle() {
    if (!confirm.open) {
      return;
    }
    const product = confirm.product;
    setConfirmLoading(true);
    setConfirmError(null);
    setBusyId(product.id);
    try {
      if (product.isActive) {
        await deactivateProduct(product.id);
      } else {
        await activateProduct(product.id);
      }
      setConfirm({ open: false });
      await loadProducts();
    } catch (err) {
      setConfirmError(
        getApiBusinessMessage(err, "Akcija nije uspjela. Pokušajte ponovo."),
      );
    } finally {
      setConfirmLoading(false);
      setBusyId(null);
    }
  }

  return (
    <div className="min-w-0 space-y-4 sm:space-y-5">
      <header className="dash-enter flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div className="min-w-0">
          <p className="text-[11px] font-medium uppercase tracking-[0.16em] text-[#8a6a45]">
            AM Keramika
          </p>
          <h1 className="mt-1 break-words text-2xl font-semibold tracking-tight text-stone-900 sm:text-3xl">
            Proizvodi
          </h1>
          <p className="mt-1 max-w-2xl break-words text-sm text-stone-500">
            Katalog artikala — pretraga, filteri i upravljanje statusom.
          </p>
        </div>
        <Link
          href="/products/new"
          className="inline-flex min-h-11 shrink-0 items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white transition hover:bg-stone-800"
        >
          Novi proizvod
        </Link>
      </header>

      <ProductFilters
        values={filterValues}
        categories={categories}
        groups={groups}
        groupsLoading={groupsLoading}
        onSearchChange={setSearchInput}
        onChange={(patch) => {
          syncQuery({
            ...patch,
            page: 1,
          });
        }}
      />

      <ProductList
        products={products}
        pagination={pagination}
        loading={loading}
        error={error}
        busyId={busyId}
        onRetry={() => void loadProducts()}
        onToggleActive={(product) => {
          setConfirmError(null);
          setConfirm({ open: true, product });
        }}
        onPageChange={(nextPage) => syncQuery({ page: nextPage })}
      />

      <ConfirmDialog
        open={confirm.open}
        title={
          confirm.open && confirm.product.isActive
            ? "Deaktiviraj proizvod"
            : "Aktiviraj proizvod"
        }
        message={
          confirm.open
            ? confirm.product.isActive
              ? `Deaktivirati proizvod „${confirm.product.name}”?`
              : `Aktivirati proizvod „${confirm.product.name}”?`
            : ""
        }
        confirmLabel={
          confirm.open && confirm.product.isActive
            ? "Deaktiviraj"
            : "Aktiviraj"
        }
        tone="neutral"
        loading={confirmLoading}
        error={confirmError}
        onClose={() => {
          if (!confirmLoading) {
            setConfirm({ open: false });
          }
        }}
        onConfirm={() => void handleConfirmToggle()}
      />
    </div>
  );
}
