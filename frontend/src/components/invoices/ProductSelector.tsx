"use client";

import { useCallback, useEffect, useState } from "react";

import { ListSkeleton } from "@/components/ui/EmptyState";
import { ProductSalePrice } from "@/components/products/ProductSalePrice";
import { fetchCategories, fetchProductGroups } from "@/lib/categories-api";
import { fetchProducts, getApiBusinessMessage } from "@/lib/products-api";
import { Category } from "@/types/category";
import { Product } from "@/types/product";
import { ProductGroup } from "@/types/product-group";

export function ProductSelector({
  open,
  onClose,
  onSelect,
  excludeFullySelected,
}: {
  open: boolean;
  onClose: () => void;
  onSelect: (product: Product) => void;
  /** productID -> already selected qty; hide add if stock fully used */
  excludeFullySelected?: Map<number, number>;
}) {
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [categoryID, setCategoryID] = useState<number | "">("");
  const [groupID, setGroupID] = useState<number | "">("");
  const [categories, setCategories] = useState<Category[]>([]);
  const [groups, setGroups] = useState<ProductGroup[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) {
      return;
    }
    const timer = window.setTimeout(() => {
      setDebouncedSearch(search.trim());
    }, 350);
    return () => window.clearTimeout(timer);
  }, [open, search]);

  useEffect(() => {
    if (!open) {
      return;
    }
    let cancelled = false;
    void (async () => {
      try {
        const data = await fetchCategories(false);
        if (!cancelled) {
          setCategories(data.filter((item) => item.isActive));
        }
      } catch {
        if (!cancelled) {
          setCategories([]);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [open]);

  useEffect(() => {
    if (!open) {
      return;
    }
    if (!categoryID) {
      const timer = window.setTimeout(() => {
        setGroups([]);
      }, 0);
      return () => window.clearTimeout(timer);
    }
    let cancelled = false;
    void (async () => {
      try {
        const data = await fetchProductGroups(Number(categoryID));
        if (!cancelled) {
          setGroups(data);
        }
      } catch {
        if (!cancelled) {
          setGroups([]);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [open, categoryID]);

  const loadProducts = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetchProducts({
        page: 1,
        limit: 30,
        search: debouncedSearch || undefined,
        categoryID: categoryID ? Number(categoryID) : undefined,
        groupID: groupID ? Number(groupID) : undefined,
        includeInactive: false,
      });
      setProducts((response.products ?? []).filter((item) => item.isActive));
    } catch (err) {
      setProducts([]);
      setError(getApiBusinessMessage(err, "Nije moguće učitati proizvode."));
    } finally {
      setLoading(false);
    }
  }, [categoryID, debouncedSearch, groupID]);

  useEffect(() => {
    if (!open) {
      return;
    }
    const timer = window.setTimeout(() => {
      void loadProducts();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [open, loadProducts]);

  useEffect(() => {
    if (!open) {
      return;
    }
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onClose();
      }
    }
    window.addEventListener("keydown", onKeyDown);
    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener("keydown", onKeyDown);
    };
  }, [open, onClose]);

  if (!open) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center sm:items-center sm:p-4">
      <button
        type="button"
        aria-label="Zatvori"
        className="absolute inset-0 bg-stone-950/45 backdrop-blur-[1px]"
        onClick={onClose}
      />
      <div
        role="dialog"
        aria-modal="true"
        aria-label="Izbor proizvoda"
        className="relative z-10 flex max-h-[92vh] w-full max-w-3xl flex-col overflow-hidden rounded-t-2xl border border-stone-200 bg-white shadow-xl sm:rounded-2xl"
      >
        <div className="flex items-start justify-between gap-3 border-b border-stone-100 px-4 py-4 sm:px-5">
          <div>
            <h2 className="text-lg font-semibold text-stone-900">
              Dodaj proizvod
            </h2>
            <p className="mt-1 text-sm text-stone-500">
              Samo aktivni proizvodi
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-stone-200 text-stone-600 hover:bg-stone-50"
          >
            ×
          </button>
        </div>

        <div className="space-y-3 border-b border-stone-100 px-4 py-3 sm:px-5">
          <input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Pretraži proizvode"
            className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
          />
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
            <select
              value={categoryID}
              onChange={(event) => {
                setCategoryID(
                  event.target.value ? Number(event.target.value) : "",
                );
                setGroupID("");
              }}
              className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm"
            >
              <option value="">Sve kategorije</option>
              {categories.map((category) => (
                <option key={category.id} value={category.id}>
                  {category.name}
                </option>
              ))}
            </select>
            <select
              value={groupID}
              disabled={!categoryID}
              onChange={(event) =>
                setGroupID(event.target.value ? Number(event.target.value) : "")
              }
              className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm disabled:opacity-50"
            >
              <option value="">Sve grupe</option>
              {groups.map((group) => (
                <option key={group.id} value={group.id}>
                  {group.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto px-4 py-3 sm:px-5">
          {loading ? <ListSkeleton rows={4} /> : null}
          {!loading && error ? (
            <p className="text-sm text-red-700">{error}</p>
          ) : null}
          {!loading && !error && products.length === 0 ? (
            <p className="py-8 text-center text-sm text-stone-500">
              Nema proizvoda za prikaz.
            </p>
          ) : null}
          {!loading && !error ? (
            <ul className="space-y-2">
              {products.map((product) => {
                const used = excludeFullySelected?.get(product.id) ?? 0;
                const available = Math.max(0, product.stockQuantity - used);
                const outOfStock = available <= 0;
                const imageUrl =
                  product.primaryImage?.url ??
                  product.images?.find((img) => img.isPrimary)?.url ??
                  product.images?.[0]?.url ??
                  null;

                return (
                  <li key={product.id}>
                    <button
                      type="button"
                      disabled={outOfStock}
                      onClick={() => {
                        onSelect(product);
                        onClose();
                      }}
                      className="flex w-full min-w-0 items-center gap-3 rounded-xl border border-stone-200 px-3 py-3 text-left transition hover:border-[#c4a484]/60 hover:bg-stone-50 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      <div className="h-12 w-12 shrink-0 overflow-hidden rounded-lg bg-stone-100">
                        {imageUrl ? (
                          // eslint-disable-next-line @next/next/no-img-element
                          <img
                            src={imageUrl}
                            alt=""
                            className="h-full w-full object-cover"
                          />
                        ) : (
                          <div className="flex h-full w-full items-center justify-center text-[10px] text-stone-400">
                            N/A
                          </div>
                        )}
                      </div>
                      <div className="min-w-0 flex-1">
                        <p className="truncate font-medium text-stone-900">
                          {product.name}
                        </p>
                        <p className="mt-0.5 truncate text-xs text-stone-500">
                          {[product.category?.name, product.group?.name]
                            .filter(Boolean)
                            .join(" · ") || "Bez grupe"}
                          {" · "}
                          {product.unit}
                        </p>
                        <p className="mt-1 text-sm text-stone-700">
                          <ProductSalePrice product={product} /> · Stanje{" "}
                          {product.stockQuantity}
                          {outOfStock ? " · Nema na stanju" : ""}
                        </p>
                      </div>
                    </button>
                  </li>
                );
              })}
            </ul>
          ) : null}
        </div>
      </div>
    </div>
  );
}
