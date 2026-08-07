"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

import { AdjustStockModal } from "@/components/inventory/AdjustStockModal";
import {
  InventoryFilterValues,
  InventoryFilters,
} from "@/components/inventory/InventoryFilters";
import { InventoryHistoryList } from "@/components/inventory/InventoryHistoryList";
import { InventoryStockList } from "@/components/inventory/InventoryStockList";
import { InlineError, ListSkeleton } from "@/components/ui/EmptyState";
import {
  fetchCategories,
  fetchProductGroups,
} from "@/lib/categories-api";
import {
  fetchInventoryMovements,
  fetchInventoryStock,
  fetchInventorySummary,
  getApiBusinessMessage,
} from "@/lib/inventory-api";
import { Category } from "@/types/category";
import {
  InventoryMovement,
  InventoryPagination,
  InventoryProductRow,
  InventoryStockStatus,
  InventorySummary,
  InventoryTab,
} from "@/types/inventory";
import { ProductGroup } from "@/types/product-group";

function parsePositiveInt(value: string | null): number | null {
  if (!value) {
    return null;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

function parseStatus(value: string | null): InventoryStockStatus {
  if (value === "low" || value === "out") {
    return value;
  }
  return "all";
}

function parseTab(value: string | null): InventoryTab {
  return value === "history" ? "history" : "stock";
}

export function InventoryWorkspace() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const page = parsePositiveInt(searchParams.get("page")) ?? 1;
  const categoryID = parsePositiveInt(searchParams.get("categoryID"));
  const groupID = parsePositiveInt(searchParams.get("groupID"));
  const status = parseStatus(searchParams.get("status"));
  const tab = parseTab(searchParams.get("tab"));
  const searchFromUrl = searchParams.get("search") ?? "";
  const movementType = searchParams.get("type") ?? "";

  const [searchInput, setSearchInput] = useState(searchFromUrl);
  const [categories, setCategories] = useState<Category[]>([]);
  const [groups, setGroups] = useState<ProductGroup[]>([]);
  const [groupsLoading, setGroupsLoading] = useState(false);

  const [products, setProducts] = useState<InventoryProductRow[]>([]);
  const [pagination, setPagination] = useState<InventoryPagination | null>(
    null,
  );
  const [movements, setMovements] = useState<InventoryMovement[]>([]);
  const [movementPagination, setMovementPagination] =
    useState<InventoryPagination | null>(null);
  const [summary, setSummary] = useState<InventorySummary | null>(null);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reloadToken, setReloadToken] = useState(0);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const [adjustProduct, setAdjustProduct] =
    useState<InventoryProductRow | null>(null);
  const [adjustOpen, setAdjustOpen] = useState(false);

  const visibleGroups = categoryID ? groups : [];

  const filterValues: InventoryFilterValues = useMemo(
    () => ({
      search: searchInput,
      categoryID,
      groupID,
      status,
    }),
    [searchInput, categoryID, groupID, status],
  );

  const syncQuery = useCallback(
    (patch: {
      page?: number;
      search?: string;
      categoryID?: number | null;
      groupID?: number | null;
      status?: InventoryStockStatus;
      tab?: InventoryTab;
      type?: string;
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

      const nextGroup = patch.groupID !== undefined ? patch.groupID : groupID;
      if (nextGroup) {
        params.set("groupID", String(nextGroup));
      } else {
        params.delete("groupID");
      }

      const nextStatus = patch.status !== undefined ? patch.status : status;
      if (nextStatus !== "all") {
        params.set("status", nextStatus);
      } else {
        params.delete("status");
      }

      const nextTab = patch.tab !== undefined ? patch.tab : tab;
      if (nextTab === "history") {
        params.set("tab", "history");
      } else {
        params.delete("tab");
      }

      const nextType = patch.type !== undefined ? patch.type : movementType;
      if (nextType) {
        params.set("type", nextType);
      } else {
        params.delete("type");
      }

      router.replace(`${pathname}?${params.toString()}`, { scroll: false });
    },
    [
      categoryID,
      groupID,
      movementType,
      page,
      pathname,
      router,
      searchFromUrl,
      searchParams,
      status,
      tab,
    ],
  );

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void fetchCategories(false).then(setCategories);
    }, 0);
    return () => window.clearTimeout(timer);
  }, []);

  useEffect(() => {
    let cancelled = false;
    const timer = window.setTimeout(() => {
      if (!categoryID) {
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

  useEffect(() => {
    const timer = window.setTimeout(() => {
      if (searchInput !== searchFromUrl) {
        syncQuery({ search: searchInput, page: 1 });
      }
    }, 350);
    return () => window.clearTimeout(timer);
  }, [searchInput, searchFromUrl, syncQuery]);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const summaryPromise = fetchInventorySummary();
      if (tab === "history") {
        const history = await fetchInventoryMovements({
          page,
          limit: 20,
          type: movementType || undefined,
        });
        setMovements(history.movements);
        setMovementPagination(history.pagination);
        setProducts([]);
        setPagination(null);
      } else {
        const stock = await fetchInventoryStock({
          page,
          limit: 20,
          search: searchFromUrl,
          categoryID: categoryID ?? undefined,
          groupID: groupID ?? undefined,
          status,
        });
        setProducts(stock.products);
        setPagination(stock.pagination);
        setMovements([]);
        setMovementPagination(null);
      }
      setSummary(await summaryPromise);
    } catch (err) {
      setProducts([]);
      setPagination(null);
      setMovements([]);
      setMovementPagination(null);
      setError(getApiBusinessMessage(err, "Lager nije učitan."));
    } finally {
      setLoading(false);
    }
  }, [
    categoryID,
    groupID,
    movementType,
    page,
    searchFromUrl,
    status,
    tab,
  ]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadData();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [loadData, reloadToken]);

  useEffect(() => {
    if (!successMessage) {
      return;
    }
    const timer = window.setTimeout(() => setSuccessMessage(null), 3500);
    return () => window.clearTimeout(timer);
  }, [successMessage]);

  function openAdjust(product?: InventoryProductRow) {
    if (product) {
      setAdjustProduct(product);
    }
    setAdjustOpen(true);
  }

  const activePagination = tab === "history" ? movementPagination : pagination;

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-stone-900">
            Lager
          </h1>
          <p className="mt-1 text-sm text-stone-500">
            Pregled stanja, korekcije i istorije kretanja.
          </p>
        </div>
        <button
          type="button"
          onClick={() => openAdjust()}
          disabled={tab !== "stock"}
          className="inline-flex min-h-11 items-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white hover:bg-stone-800 disabled:cursor-not-allowed disabled:opacity-50"
        >
          Korekcija lagera
        </button>
      </div>

      {summary ? (
        <div className="flex flex-wrap gap-3 text-sm">
          <div className="rounded-xl border border-stone-200 bg-white px-3 py-2">
            <span className="text-stone-500">Nizak lager:</span>{" "}
            <span className="font-semibold tabular-nums text-stone-900">
              {summary.lowStockCount}
            </span>
          </div>
          <div className="rounded-xl border border-stone-200 bg-white px-3 py-2">
            <span className="text-stone-500">Nema na stanju:</span>{" "}
            <span className="font-semibold tabular-nums text-stone-900">
              {summary.outOfStockCount}
            </span>
          </div>
        </div>
      ) : null}

      {successMessage ? (
        <p className="rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-800">
          {successMessage}
        </p>
      ) : null}

      <div className="inline-flex rounded-xl border border-stone-200 bg-white p-1">
        <button
          type="button"
          onClick={() => syncQuery({ tab: "stock", page: 1 })}
          className={`rounded-lg px-4 py-2 text-sm font-medium transition ${
            tab === "stock"
              ? "bg-stone-900 text-white"
              : "text-stone-600 hover:bg-stone-50"
          }`}
        >
          Stanje lagera
        </button>
        <button
          type="button"
          onClick={() => syncQuery({ tab: "history", page: 1 })}
          className={`rounded-lg px-4 py-2 text-sm font-medium transition ${
            tab === "history"
              ? "bg-stone-900 text-white"
              : "text-stone-600 hover:bg-stone-50"
          }`}
        >
          Istorija kretanja
        </button>
      </div>

      {tab === "stock" ? (
        <InventoryFilters
          values={filterValues}
          categories={categories}
          groups={visibleGroups}
          groupsLoading={groupsLoading}
          onSearchChange={setSearchInput}
          onChange={(patch) =>
            syncQuery({
              ...patch,
              page: 1,
              groupID:
                patch.categoryID !== undefined && patch.categoryID !== categoryID
                  ? null
                  : patch.groupID,
            })
          }
          onReset={() => {
            setSearchInput("");
            syncQuery({
              search: "",
              categoryID: null,
              groupID: null,
              status: "all",
              page: 1,
            });
          }}
        />
      ) : (
        <div className="rounded-2xl border border-stone-200 bg-white p-4">
          <label
            htmlFor="movement-type-filter"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            Tip kretanja
          </label>
          <select
            id="movement-type-filter"
            value={movementType}
            onChange={(event) =>
              syncQuery({ type: event.target.value, page: 1 })
            }
            className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2 sm:max-w-xs"
          >
            <option value="">Svi tipovi</option>
            <option value="sale">Prodaja</option>
            <option value="return">Povrat</option>
            <option value="adjust">Korekcija</option>
            <option value="in">Ulaz</option>
          </select>
        </div>
      )}

      {error ? (
        <InlineError
          message={error}
          onRetry={() => setReloadToken((value) => value + 1)}
        />
      ) : loading ? (
        <ListSkeleton rows={5} />
      ) : tab === "stock" ? (
        <InventoryStockList products={products} onAdjust={openAdjust} />
      ) : (
        <InventoryHistoryList movements={movements} />
      )}

      {activePagination && activePagination.totalPages > 1 ? (
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-stone-200 bg-white px-4 py-3 text-sm">
          <p className="text-stone-600">
            Stranica {activePagination.page} / {activePagination.totalPages}
          </p>
          <div className="flex gap-2">
            <button
              type="button"
              disabled={activePagination.page <= 1}
              onClick={() => syncQuery({ page: activePagination.page - 1 })}
              className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 font-medium text-stone-700 disabled:opacity-50"
            >
              Prethodna
            </button>
            <button
              type="button"
              disabled={
                activePagination.page >= activePagination.totalPages
              }
              onClick={() => syncQuery({ page: activePagination.page + 1 })}
              className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 font-medium text-stone-700 disabled:opacity-50"
            >
              Sledeća
            </button>
          </div>
        </div>
      ) : null}

      <AdjustStockModal
        open={adjustOpen}
        product={adjustProduct}
        productOptions={products}
        onClose={() => {
          setAdjustOpen(false);
          setAdjustProduct(null);
        }}
        onSuccess={() => {
          setSuccessMessage("Stanje lagera je ažurirano.");
          setReloadToken((value) => value + 1);
        }}
      />
    </div>
  );
}
