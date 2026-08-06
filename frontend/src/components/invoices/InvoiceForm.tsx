"use client";

import { useRouter, useSearchParams } from "next/navigation";
import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
} from "react";

import { CustomerSelector } from "@/components/customers/CustomerSelector";
import { InvoiceCart } from "@/components/invoices/pos/InvoiceCart";
import { InvoiceSaleTypeSwitch } from "@/components/invoices/pos/InvoiceSaleTypeSwitch";
import { InvoiceStickyCartPanel } from "@/components/invoices/pos/InvoiceStickySummary";
import { MobileInvoiceBottomBar } from "@/components/invoices/pos/MobileInvoiceBottomBar";
import { MobileInvoiceCartDrawer } from "@/components/invoices/pos/MobileInvoiceCartDrawer";
import { PosCategoryChips } from "@/components/invoices/pos/PosCategoryChips";
import {
  PosProductSearch,
  type PosProductSearchHandle,
} from "@/components/invoices/pos/PosProductSearch";
import { PosProductResults } from "@/components/invoices/pos/PosProductResults";
import { PosQuickProducts } from "@/components/invoices/pos/PosQuickProducts";
import { fetchCategories, fetchProductGroups } from "@/lib/categories-api";
import { fetchCustomer } from "@/lib/customers-api";
import {
  createInvoice,
  getApiBusinessMessage,
} from "@/lib/invoices-api";
import { fetchProducts } from "@/lib/products-api";
import { Category } from "@/types/category";
import { CustomerListItem } from "@/types/customer";
import { InvoiceFormLine } from "@/types/invoice";
import { Product } from "@/types/product";
import { ProductGroup } from "@/types/product-group";

type CustomerMode = "cash" | "customer";

function productImageUrl(product: Product): string | null {
  return (
    product.primaryImage?.url ??
    product.images?.find((img) => img.isPrimary)?.url ??
    product.images?.[0]?.url ??
    null
  );
}

function roundQty(value: number): number {
  return Math.round(value * 100) / 100;
}

function markMatchingLines(
  lines: InvoiceFormLine[],
  message: string,
): Record<number, string> {
  const lower = message.toLowerCase();
  const next: Record<number, string> = {};
  for (const line of lines) {
    if (lower.includes(line.name.toLowerCase())) {
      next[line.productID] = message;
    }
  }
  if (Object.keys(next).length === 0 && lower.includes("lager")) {
    for (const line of lines) {
      next[line.productID] =
        "Provjerite lager — backend je odbio količinu.";
    }
  }
  return next;
}

export function InvoiceForm({
  initialCustomerID,
}: {
  initialCustomerID?: number | null;
}) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const searchRef = useRef<PosProductSearchHandle>(null);
  const highlightTimer = useRef<number | null>(null);

  const [customerMode, setCustomerMode] = useState<CustomerMode>(
    initialCustomerID ? "customer" : "cash",
  );
  const [customer, setCustomer] = useState<CustomerListItem | null>(null);
  const [customerPrefillError, setCustomerPrefillError] = useState<string | null>(
    null,
  );
  const [lines, setLines] = useState<InvoiceFormLine[]>([]);
  const [lineErrors, setLineErrors] = useState<Record<number, string>>({});
  const [highlightedProductID, setHighlightedProductID] = useState<number | null>(
    null,
  );
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [mobileDrawerOpen, setMobileDrawerOpen] = useState(false);

  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [resultsOpen, setResultsOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);
  const [searchProducts, setSearchProducts] = useState<Product[]>([]);
  const [searchLoading, setSearchLoading] = useState(false);

  const [categories, setCategories] = useState<Category[]>([]);
  const [groups, setGroups] = useState<ProductGroup[]>([]);
  const [categoryID, setCategoryID] = useState<number | null>(null);
  const [groupID, setGroupID] = useState<number | null>(null);

  const [quickProducts, setQuickProducts] = useState<Product[]>([]);
  const [quickLoading, setQuickLoading] = useState(false);

  const dirty = lines.length > 0 || (customerMode === "customer" && customer);
  const searchActive = debouncedSearch.length > 0;

  useEffect(() => {
    if (!dirty) {
      return;
    }
    function onBeforeUnload(event: BeforeUnloadEvent) {
      event.preventDefault();
      event.returnValue = "";
    }
    window.addEventListener("beforeunload", onBeforeUnload);
    return () => window.removeEventListener("beforeunload", onBeforeUnload);
  }, [dirty]);

  useEffect(() => {
    if (!initialCustomerID || initialCustomerID <= 0) {
      return;
    }
    let cancelled = false;
    void (async () => {
      try {
        const data = await fetchCustomer(initialCustomerID);
        if (cancelled) {
          return;
        }
        if (!data.isActive) {
          setCustomerPrefillError(
            "Kupac iz linka nije aktivan. Izaberite drugog kupca ili gotovinsku prodaju.",
          );
          setCustomerMode("customer");
          setCustomer(null);
          return;
        }
        setCustomer({
          id: data.id,
          name: data.name,
          phone: data.phone,
          isActive: data.isActive,
        });
        setCustomerMode("customer");
        setCustomerPrefillError(null);
      } catch {
        if (!cancelled) {
          setCustomerPrefillError(
            "Kupac iz linka nije pronađen. Izaberite kupca ručno.",
          );
          setCustomerMode("customer");
          setCustomer(null);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [initialCustomerID]);

  useEffect(() => {
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
  }, []);

  useEffect(() => {
    if (categoryID == null) {
      const timer = window.setTimeout(() => setGroups([]), 0);
      return () => window.clearTimeout(timer);
    }
    let cancelled = false;
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
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [categoryID]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      setDebouncedSearch(search.trim());
      setActiveIndex(0);
    }, 250);
    return () => window.clearTimeout(timer);
  }, [search]);

  useEffect(() => {
    if (!searchActive) {
      const timer = window.setTimeout(() => {
        setSearchProducts([]);
        setResultsOpen(false);
      }, 0);
      return () => window.clearTimeout(timer);
    }
    let cancelled = false;
    void (async () => {
      setSearchLoading(true);
      try {
        const response = await fetchProducts({
          page: 1,
          limit: 20,
          search: debouncedSearch,
          categoryID: categoryID ?? undefined,
          groupID: groupID ?? undefined,
          includeInactive: false,
        });
        if (!cancelled) {
          const items = (response.products ?? []).filter((p) => p.isActive);
          setSearchProducts(items);
          setResultsOpen(true);
          setActiveIndex(0);
        }
      } catch {
        if (!cancelled) {
          setSearchProducts([]);
          setResultsOpen(true);
        }
      } finally {
        if (!cancelled) {
          setSearchLoading(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [searchActive, debouncedSearch, categoryID, groupID]);

  useEffect(() => {
    if (searchActive) {
      return;
    }
    let cancelled = false;
    void (async () => {
      setQuickLoading(true);
      try {
        const response = await fetchProducts({
          page: 1,
          limit: 24,
          categoryID: categoryID ?? undefined,
          groupID: groupID ?? undefined,
          includeInactive: false,
        });
        if (!cancelled) {
          setQuickProducts(
            (response.products ?? []).filter((item) => item.isActive),
          );
        }
      } catch {
        if (!cancelled) {
          setQuickProducts([]);
        }
      } finally {
        if (!cancelled) {
          setQuickLoading(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [searchActive, categoryID, groupID]);

  useEffect(() => {
    function onGlobalKey(event: KeyboardEvent) {
      if (event.key !== "/" || event.metaKey || event.ctrlKey || event.altKey) {
        return;
      }
      const target = event.target as HTMLElement | null;
      const tag = target?.tagName?.toLowerCase();
      if (
        tag === "input" ||
        tag === "textarea" ||
        tag === "select" ||
        target?.isContentEditable
      ) {
        return;
      }
      event.preventDefault();
      searchRef.current?.focus();
    }
    window.addEventListener("keydown", onGlobalKey);
    return () => window.removeEventListener("keydown", onGlobalKey);
  }, []);

  useEffect(() => {
    return () => {
      if (highlightTimer.current != null) {
        window.clearTimeout(highlightTimer.current);
      }
    };
  }, []);

  const selectedQtyByProduct = useMemo(() => {
    const map = new Map<number, number>();
    for (const line of lines) {
      map.set(line.productID, line.quantity);
    }
    return map;
  }, [lines]);

  const flashHighlight = useCallback((productID: number) => {
    setHighlightedProductID(productID);
    if (highlightTimer.current != null) {
      window.clearTimeout(highlightTimer.current);
    }
    highlightTimer.current = window.setTimeout(() => {
      setHighlightedProductID(null);
    }, 700);
  }, []);

  const addProduct = useCallback(
    (product: Product) => {
      if (product.stockQuantity <= 0) {
        return;
      }
      setError(null);
      setLines((current) => {
        const existing = current.find((line) => line.productID === product.id);
        if (existing) {
          const nextQty = roundQty(existing.quantity + 1);
          if (nextQty > product.stockQuantity) {
            setLineErrors((errors) => ({
              ...errors,
              [product.id]: `Maksimalna količina je ${product.stockQuantity}.`,
            }));
            return current;
          }
          setLineErrors((errors) => {
            const next = { ...errors };
            delete next[product.id];
            return next;
          });
          return current.map((line) =>
            line.productID === product.id
              ? {
                  ...line,
                  quantity: nextQty,
                  stockQuantity: product.stockQuantity,
                  salePrice: product.salePrice,
                }
              : line,
          );
        }
        return [
          ...current,
          {
            productID: product.id,
            name: product.name,
            unit: product.unit,
            salePrice: product.salePrice,
            stockQuantity: product.stockQuantity,
            imageUrl: productImageUrl(product),
            quantity: 1,
          },
        ];
      });
      flashHighlight(product.id);
      setSearch("");
      setDebouncedSearch("");
      setResultsOpen(false);
      setSearchProducts([]);
      window.requestAnimationFrame(() => {
        searchRef.current?.focus();
      });
    },
    [flashHighlight],
  );

  function updateQuantity(productID: number, quantity: number) {
    setLines((current) =>
      current.map((line) => {
        if (line.productID !== productID) {
          return line;
        }
        return { ...line, quantity };
      }),
    );
    setLineErrors((errors) => {
      const next = { ...errors };
      delete next[productID];
      return next;
    });
  }

  function removeLine(productID: number) {
    setLines((current) => current.filter((line) => line.productID !== productID));
    setLineErrors((errors) => {
      const next = { ...errors };
      delete next[productID];
      return next;
    });
  }

  function validate(): boolean {
    if (customerMode === "customer" && !customer) {
      setError("Izaberite kupca ili prebacite na gotovinsku prodaju.");
      return false;
    }
    if (lines.length === 0) {
      setError("Dodajte najmanje jednu stavku.");
      return false;
    }
    const nextErrors: Record<number, string> = {};
    for (const line of lines) {
      if (!Number.isFinite(line.quantity) || line.quantity <= 0) {
        nextErrors[line.productID] = "Količina mora biti veća od 0.";
      } else if (line.quantity > line.stockQuantity) {
        nextErrors[line.productID] =
          `Količina ne smije prelaziti lager (${line.stockQuantity}).`;
      }
    }
    setLineErrors(nextErrors);
    if (Object.keys(nextErrors).length > 0) {
      setError("Ispravite količine u označenim stavkama.");
      return false;
    }
    return true;
  }

  const canSubmit =
    lines.length > 0 &&
    !(customerMode === "customer" && !customer) &&
    lines.every(
      (line) =>
        Number.isFinite(line.quantity) &&
        line.quantity > 0 &&
        line.quantity <= line.stockQuantity,
    ) &&
    Object.keys(lineErrors).length === 0;

  async function handleSubmit(withPrint: boolean) {
    if (submitting || !validate()) {
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      const invoice = await createInvoice({
        customerID: customerMode === "customer" ? customer?.id : null,
        items: lines.map((line) => ({
          productID: line.productID,
          quantity: line.quantity,
        })),
      });
      setLines([]);
      setMobileDrawerOpen(false);
      if (withPrint) {
        router.replace(`/invoices/${invoice.id}/print?autoprint=1`);
      } else {
        router.replace(`/invoices/${invoice.id}`);
      }
    } catch (err) {
      const message = getApiBusinessMessage(
        err,
        "Kreiranje računa nije uspjelo.",
      );
      setError(message);
      setLineErrors((current) => ({
        ...current,
        ...markMatchingLines(lines, message),
      }));
      setMobileDrawerOpen(true);
    } finally {
      setSubmitting(false);
    }
  }

  function handleSaleTypeChange(next: CustomerMode) {
    setCustomerMode(next);
    setError(null);
    if (next === "cash") {
      setCustomer(null);
      setCustomerPrefillError(null);
      if (searchParams.get("customerID")) {
        router.replace("/invoices/new");
      }
    }
  }

  function handleSearchKeyDown(event: ReactKeyboardEvent<HTMLInputElement>) {
    if (event.key === "Escape") {
      setResultsOpen(false);
      return;
    }
    if (!resultsOpen && searchActive && searchProducts.length > 0) {
      setResultsOpen(true);
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      if (searchProducts.length === 0) {
        return;
      }
      setResultsOpen(true);
      setActiveIndex((index) => (index + 1) % searchProducts.length);
      return;
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      if (searchProducts.length === 0) {
        return;
      }
      setResultsOpen(true);
      setActiveIndex((index) =>
        index <= 0 ? searchProducts.length - 1 : index - 1,
      );
      return;
    }
    if (event.key === "Enter") {
      event.preventDefault();
      if (searchProducts.length === 0) {
        return;
      }
      const exact = searchProducts.filter(
        (product) =>
          product.name.trim().toLowerCase() === debouncedSearch.toLowerCase(),
      );
      const candidate =
        exact.length === 1
          ? exact[0]
          : searchProducts[Math.min(activeIndex, searchProducts.length - 1)];
      if (!candidate || candidate.stockQuantity <= 0) {
        return;
      }
      const used = selectedQtyByProduct.get(candidate.id) ?? 0;
      if (candidate.stockQuantity - used <= 0) {
        return;
      }
      addProduct(candidate);
    }
  }

  const customerLabel =
    customerMode === "cash"
      ? "Gotovinska prodaja"
      : customer?.name
        ? customer.name
        : "Kupac nije izabran";

  const previewTotal = lines.reduce(
    (sum, line) =>
      sum +
      (Number.isFinite(line.quantity) ? line.salePrice * line.quantity : 0),
    0,
  );

  const cartNode = (
    <InvoiceCart
      lines={lines}
      lineErrors={lineErrors}
      highlightedProductID={highlightedProductID}
      onQuantityChange={updateQuantity}
      onRemove={removeLine}
    />
  );

  return (
    <div className="min-w-0 pb-24 lg:pb-0">
      <header className="mb-4 flex flex-wrap items-center justify-end gap-3">
        <InvoiceSaleTypeSwitch
          value={customerMode}
          onChange={handleSaleTypeChange}
          disabled={submitting}
        />
      </header>

      {customerMode === "customer" ? (
        <div className="mb-4 rounded-2xl border border-stone-200 bg-white p-3 sm:p-4">
          <CustomerSelector
            value={customer}
            onChange={(next) => {
              setCustomer(next);
              setCustomerPrefillError(null);
            }}
          />
          {customerPrefillError ? (
            <p className="mt-2 break-words text-sm text-amber-800">
              {customerPrefillError}
            </p>
          ) : null}
        </div>
      ) : null}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1.95fr)_minmax(18rem,1fr)] lg:items-start">
        <div className="min-w-0 space-y-3">
          <div className="relative">
            <PosProductSearch
              ref={searchRef}
              value={search}
              onChange={(value) => {
                setSearch(value);
                if (value.trim()) {
                  setResultsOpen(true);
                }
              }}
              onKeyDown={handleSearchKeyDown}
              loading={searchLoading}
              autoFocus
            />
            {resultsOpen && searchActive ? (
              <div className="absolute z-20 mt-1.5 w-full overflow-hidden rounded-2xl border border-stone-200 bg-white shadow-lg">
                <PosProductResults
                  products={searchProducts}
                  activeIndex={activeIndex}
                  onSelect={addProduct}
                  selectedQtyByProduct={selectedQtyByProduct}
                  emptyLabel={
                    searchLoading ? "Pretraga…" : "Nema proizvoda za ovaj upit."
                  }
                />
              </div>
            ) : null}
          </div>

          <PosCategoryChips
            categories={categories}
            groups={groups}
            categoryID={categoryID}
            groupID={groupID}
            onCategoryChange={setCategoryID}
            onGroupChange={setGroupID}
          />

          {!searchActive ? (
            <PosQuickProducts
              products={quickProducts}
              loading={quickLoading}
              onSelect={addProduct}
              selectedQtyByProduct={selectedQtyByProduct}
            />
          ) : null}

          {/* Kompaktan pregled stavki na mobilnom (desktop koristi sticky panel). */}
          <section className="lg:hidden">
            <h2 className="mb-2 text-sm font-semibold text-stone-800">
              Odabrane stavke
            </h2>
            {cartNode}
            {error ? (
              <p className="mt-2 break-words rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
                {error}
              </p>
            ) : null}
          </section>
        </div>

        <div className="hidden lg:block">
          <InvoiceStickyCartPanel
            lines={lines}
            customerLabel={customerLabel}
            isCashSale={customerMode === "cash"}
            submitting={submitting}
            error={error}
            canSubmit={canSubmit}
            onSubmitPrint={() => void handleSubmit(true)}
            onSubmitNoPrint={() => void handleSubmit(false)}
            cart={cartNode}
          />
        </div>
      </div>

      <MobileInvoiceBottomBar
        itemCount={lines.length}
        previewTotal={previewTotal}
        onReview={() => setMobileDrawerOpen(true)}
      />

      <MobileInvoiceCartDrawer
        open={mobileDrawerOpen}
        onClose={() => setMobileDrawerOpen(false)}
        lines={lines}
        lineErrors={lineErrors}
        highlightedProductID={highlightedProductID}
        customerLabel={customerLabel}
        isCashSale={customerMode === "cash"}
        submitting={submitting}
        error={error}
        canSubmit={canSubmit}
        onQuantityChange={updateQuantity}
        onRemove={removeLine}
        onSubmitPrint={() => void handleSubmit(true)}
        onSubmitNoPrint={() => void handleSubmit(false)}
      />
    </div>
  );
}
