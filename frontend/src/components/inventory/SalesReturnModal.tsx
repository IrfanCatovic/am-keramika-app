"use client";

import { FormEvent, useEffect, useId, useMemo, useState } from "react";

import { Modal } from "@/components/ui/Modal";
import { formatMoney, formatQuantity, formatUnit } from "@/lib/format";
import { fetchProducts } from "@/lib/products-api";
import {
  createSalesReturn,
  getApiBusinessMessage,
} from "@/lib/sales-return-api";
import { Product } from "@/types/product";

type CashRefundedChoice = boolean | null;

type DraftItem = {
  key: string;
  product: Product | null;
  search: string;
  quantityInput: string;
  unitPriceInput: string;
  productError: string | null;
  quantityError: string | null;
  unitPriceError: string | null;
};

const DESCRIPTION_MAX = 1000;

function createEmptyItem(): DraftItem {
  return {
    key: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    product: null,
    search: "",
    quantityInput: "",
    unitPriceInput: "",
    productError: null,
    quantityError: null,
    unitPriceError: null,
  };
}

function parsePositiveNumber(value: string): number | null {
  const normalized = value.trim().replace(",", ".");
  if (!normalized) {
    return null;
  }
  const parsed = Number(normalized);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return null;
  }
  return parsed;
}

function roundMoney(value: number): number {
  return Math.round(value * 100) / 100;
}

function formatPriceInput(value: number): string {
  if (!Number.isFinite(value)) {
    return "";
  }
  const rounded = roundMoney(value);
  return String(rounded).replace(".", ",");
}

function saleModeLabel(product: Product): string {
  if (product.saleByPackage) {
    return "Prodaje se po pakovanju";
  }
  const unit = formatUnit(product.unit);
  if (unit === "kom") {
    return "Prodaje se po komadu";
  }
  return `Prodaje se po ${unit}`;
}

function ProductSearchField({
  item,
  disabled,
  excludeProductIds,
  onSearchChange,
  onSelect,
}: {
  item: DraftItem;
  disabled: boolean;
  excludeProductIds: Set<number>;
  onSearchChange: (value: string) => void;
  onSelect: (product: Product) => void;
}) {
  const listId = useId();
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [results, setResults] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [duplicateHint, setDuplicateHint] = useState<string | null>(null);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      setDebouncedSearch(item.search.trim());
    }, 350);
    return () => window.clearTimeout(timer);
  }, [item.search]);

  useEffect(() => {
    if (!open || item.product || debouncedSearch.length < 1) {
      return;
    }

    let cancelled = false;
    const timer = window.setTimeout(() => {
      setLoading(true);
      setError(null);
      void (async () => {
        try {
          const response = await fetchProducts({
            page: 1,
            limit: 20,
            search: debouncedSearch,
            includeInactive: true,
          });
          if (!cancelled) {
            setResults(response.products ?? []);
          }
        } catch (err) {
          if (!cancelled) {
            setResults([]);
            setError(
              getApiBusinessMessage(err, "Pretraga proizvoda nije uspela."),
            );
          }
        } finally {
          if (!cancelled) {
            setLoading(false);
          }
        }
      })();
    }, 0);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [debouncedSearch, item.product, open]);

  const visibleResults = debouncedSearch.length < 1 ? [] : results;

  if (item.product) {
    return (
      <div className="rounded-xl border border-stone-200 bg-stone-50 px-3 py-3">
        <div className="flex flex-wrap items-start justify-between gap-2">
          <div className="min-w-0">
            <p className="font-medium text-stone-900">{item.product.name}</p>
            <p className="mt-1 text-sm text-stone-600">
              {saleModeLabel(item.product)}
            </p>
            {item.product.saleByPackage && item.product.packageQuantity > 0 ? (
              <p className="mt-0.5 text-sm text-stone-600">
                1 paket = {formatQuantity(item.product.packageQuantity)}{" "}
                {formatUnit(item.product.unit)}
              </p>
            ) : null}
            <p className="mt-1 text-sm text-stone-600">
              Trenutna cena: {formatMoney(item.product.effectiveSalePrice)}/
              {formatUnit(item.product.unit)}
            </p>
            {!item.product.isActive ? (
              <span className="mt-2 inline-flex rounded-full border border-amber-200 bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-800">
                Neaktivan
              </span>
            ) : null}
          </div>
          <button
            type="button"
            disabled={disabled}
            onClick={() => {
              onSearchChange("");
              setOpen(true);
              setDuplicateHint(null);
            }}
            className="shrink-0 text-sm font-medium text-stone-600 underline-offset-2 hover:text-stone-900 hover:underline disabled:opacity-60"
          >
            Promeni
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="relative">
      <label
        htmlFor={listId}
        className="mb-1.5 block text-sm font-medium text-stone-700"
      >
        Pretraži proizvod
      </label>
      <input
        id={listId}
        type="search"
        autoComplete="off"
        disabled={disabled}
        value={item.search}
        placeholder="Naziv proizvoda…"
        onChange={(event) => {
          onSearchChange(event.target.value);
          setOpen(true);
          setDuplicateHint(null);
        }}
        onFocus={() => setOpen(true)}
        className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2 disabled:opacity-60"
      />
      {item.productError ? (
        <p className="mt-1.5 text-sm text-red-700">{item.productError}</p>
      ) : null}
      {duplicateHint ? (
        <p className="mt-1.5 text-sm text-amber-800">{duplicateHint}</p>
      ) : null}

      {open && !item.product ? (
        <div className="absolute z-20 mt-1 max-h-56 w-full overflow-y-auto rounded-xl border border-stone-200 bg-white shadow-lg">
          {loading ? (
            <p className="px-3 py-2.5 text-sm text-stone-500">Pretraga…</p>
          ) : error ? (
            <p className="px-3 py-2.5 text-sm text-red-700">{error}</p>
          ) : debouncedSearch.length < 1 ? (
            <p className="px-3 py-2.5 text-sm text-stone-500">
              Unesite naziv proizvoda.
            </p>
          ) : visibleResults.length === 0 ? (
            <p className="px-3 py-2.5 text-sm text-stone-500">
              Nema rezultata.
            </p>
          ) : (
            <ul className="py-1">
              {visibleResults.map((product) => {
                const alreadyAdded = excludeProductIds.has(product.id);
                return (
                  <li key={product.id}>
                    <button
                      type="button"
                      disabled={disabled}
                      onClick={() => {
                        if (alreadyAdded) {
                          setDuplicateHint(
                            "Ovaj proizvod je već dodat u povrat.",
                          );
                          return;
                        }
                        setDuplicateHint(null);
                        setOpen(false);
                        onSelect(product);
                      }}
                      className="flex w-full items-start justify-between gap-2 px-3 py-2.5 text-left text-sm hover:bg-stone-50 disabled:opacity-60"
                    >
                      <span className="min-w-0">
                        <span className="block font-medium text-stone-900">
                          {product.name}
                        </span>
                        <span className="mt-0.5 block text-stone-500">
                          {formatUnit(product.unit)}
                          {product.saleByPackage
                            ? " · pakovanje"
                            : ""}
                        </span>
                      </span>
                      {!product.isActive ? (
                        <span className="shrink-0 rounded-full border border-amber-200 bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-800">
                          Neaktivan
                        </span>
                      ) : null}
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      ) : null}
    </div>
  );
}

export function SalesReturnModal({
  open,
  onClose,
  onSuccess,
}: {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [items, setItems] = useState<DraftItem[]>([createEmptyItem()]);
  const [description, setDescription] = useState("");
  const [cashRefunded, setCashRefunded] = useState<CashRefundedChoice>(null);
  const [cashError, setCashError] = useState<string | null>(null);
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const selectedProductIds = useMemo(() => {
    const ids = new Set<number>();
    for (const item of items) {
      if (item.product) {
        ids.add(item.product.id);
      }
    }
    return ids;
  }, [items]);

  const lineTotals = useMemo(() => {
    return items.map((item) => {
      const quantity = parsePositiveNumber(item.quantityInput);
      const unitPrice = parsePositiveNumber(item.unitPriceInput);
      if (quantity == null || unitPrice == null) {
        return null;
      }
      return roundMoney(quantity * unitPrice);
    });
  }, [items]);

  const grandTotal = useMemo(() => {
    return roundMoney(
      lineTotals.reduce<number>((sum, value) => sum + (value ?? 0), 0),
    );
  }, [lineTotals]);

  const completedItemsCount = items.filter(
    (item) =>
      item.product &&
      parsePositiveNumber(item.quantityInput) != null &&
      parsePositiveNumber(item.unitPriceInput) != null,
  ).length;

  function updateItem(key: string, patch: Partial<DraftItem>) {
    setItems((current) =>
      current.map((item) => (item.key === key ? { ...item, ...patch } : item)),
    );
  }

  function handleSelectProduct(key: string, product: Product) {
    if (
      items.some(
        (item) => item.key !== key && item.product?.id === product.id,
      )
    ) {
      updateItem(key, {
        productError: "Ovaj proizvod je već dodat u povrat.",
      });
      return;
    }
    updateItem(key, {
      product,
      search: product.name,
      unitPriceInput: formatPriceInput(product.effectiveSalePrice),
      productError: null,
      unitPriceError: null,
    });
  }

  function handleClearProduct(key: string) {
    updateItem(key, {
      product: null,
      search: "",
      quantityInput: "",
      unitPriceInput: "",
      productError: null,
      quantityError: null,
      unitPriceError: null,
    });
  }

  function validate(): boolean {
    let ok = true;
    const next = items.map((item) => {
      let productError: string | null = null;
      let quantityError: string | null = null;
      let unitPriceError: string | null = null;

      if (!item.product) {
        productError = "Izaberite proizvod.";
        ok = false;
      }
      const quantity = parsePositiveNumber(item.quantityInput);
      if (quantity == null) {
        quantityError = "Unesite količinu veću od 0.";
        ok = false;
      }
      const unitPrice = parsePositiveNumber(item.unitPriceInput);
      if (unitPrice == null) {
        unitPriceError = "Unesite ispravnu cenu veću od 0.";
        ok = false;
      }

      return {
        ...item,
        productError,
        quantityError,
        unitPriceError,
      };
    });

    const seen = new Set<number>();
    for (let i = 0; i < next.length; i += 1) {
      const productId = next[i].product?.id;
      if (productId == null) {
        continue;
      }
      if (seen.has(productId)) {
        next[i] = {
          ...next[i],
          productError: "Ovaj proizvod je već dodat u povrat.",
        };
        ok = false;
      } else {
        seen.add(productId);
      }
    }

    if (next.length === 0) {
      setFormError("Dodajte najmanje jedan artikal.");
      ok = false;
    }

    if (cashRefunded == null) {
      setCashError("Izaberite da li je novac vraćen kupcu.");
      ok = false;
    } else {
      setCashError(null);
    }

    setItems(next);
    return ok;
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    if (submitting) {
      return;
    }
    setFormError(null);
    if (!validate()) {
      return;
    }
    if (cashRefunded == null) {
      return;
    }

    const payloadItems = items.flatMap((item) => {
      const quantity = parsePositiveNumber(item.quantityInput);
      const unitPrice = parsePositiveNumber(item.unitPriceInput);
      if (!item.product || quantity == null || unitPrice == null) {
        return [];
      }
      return [
        {
          productID: item.product.id,
          quantity,
          unitPrice: roundMoney(unitPrice),
        },
      ];
    });

    if (payloadItems.length === 0) {
      setFormError("Dodajte najmanje jedan artikal.");
      return;
    }

    setSubmitting(true);
    try {
      await createSalesReturn({
        description: description.trim() || undefined,
        cashRefunded,
        items: payloadItems,
      });
      onSuccess();
      onClose();
    } catch (err) {
      setFormError(
        getApiBusinessMessage(err, "Povrat robe nije uspeo."),
      );
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal
      open={open}
      title="Novi povrat robe"
      description="Unesite robu koju je kupac fizički vratio."
      onClose={() => {
        if (!submitting) {
          onClose();
        }
      }}
      size="lg"
      footer={
        <div className="space-y-3">
          <div className="rounded-xl border border-stone-200 bg-stone-50 px-3 py-3">
            <p className="text-sm text-stone-500">Ukupna vrednost povrata</p>
            <p className="mt-1 text-xl font-semibold tabular-nums text-stone-900">
              {formatMoney(grandTotal)}
            </p>
            <p className="mt-1 text-sm text-stone-500">
              {completedItemsCount === 1
                ? "1 artikal"
                : completedItemsCount >= 2 && completedItemsCount <= 4
                  ? `${completedItemsCount} artikla`
                  : `${completedItemsCount} artikala`}
            </p>
          </div>

          {formError ? (
            <p className="rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
              {formError}
            </p>
          ) : null}

          <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <button
              type="button"
              disabled={submitting}
              onClick={onClose}
              className="inline-flex min-h-11 items-center justify-center rounded-xl border border-stone-200 px-4 text-sm font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-60"
            >
              Odustani
            </button>
            <button
              type="submit"
              form="sales-return-form"
              disabled={submitting}
              className="inline-flex min-h-11 items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white hover:bg-stone-800 disabled:opacity-60"
            >
              {submitting ? "Evidentiranje…" : "Potvrdi povrat"}
            </button>
          </div>
        </div>
      }
    >
      <form id="sales-return-form" onSubmit={handleSubmit} className="space-y-5">
        <div className="space-y-4">
          {items.map((item, index) => {
            const quantity = parsePositiveNumber(item.quantityInput);
            const unitPrice = parsePositiveNumber(item.unitPriceInput);
            const lineTotal = lineTotals[index];
            const unitLabel = formatUnit(item.product?.unit ?? "");
            const excludeIds = new Set(selectedProductIds);
            if (item.product) {
              excludeIds.delete(item.product.id);
            }

            return (
              <div
                key={item.key}
                className="rounded-2xl border border-stone-200 bg-white p-4"
              >
                <div className="mb-3 flex items-center justify-between gap-2">
                  <p className="text-sm font-semibold uppercase tracking-wide text-stone-500">
                    Artikal {index + 1}
                  </p>
                  {items.length > 1 ? (
                    <button
                      type="button"
                      disabled={submitting}
                      onClick={() =>
                        setItems((current) =>
                          current.filter((row) => row.key !== item.key),
                        )
                      }
                      className="text-sm font-medium text-stone-500 underline-offset-2 hover:text-red-700 hover:underline disabled:opacity-60"
                    >
                      Ukloni
                    </button>
                  ) : null}
                </div>

                <ProductSearchField
                  item={item}
                  disabled={submitting}
                  excludeProductIds={excludeIds}
                  onSearchChange={(value) => {
                    if (item.product) {
                      handleClearProduct(item.key);
                    }
                    updateItem(item.key, {
                      search: value,
                      product: null,
                      productError: null,
                    });
                  }}
                  onSelect={(product) => handleSelectProduct(item.key, product)}
                />

                {item.product ? (
                  <div className="mt-4 grid gap-4 sm:grid-cols-2">
                    <div>
                      <label
                        htmlFor={`qty-${item.key}`}
                        className="mb-1.5 block text-sm font-medium text-stone-700"
                      >
                        Količina za povrat
                      </label>
                      <div className="flex items-center gap-2">
                        <input
                          id={`qty-${item.key}`}
                          inputMode="decimal"
                          disabled={submitting}
                          value={item.quantityInput}
                          onChange={(event) =>
                            updateItem(item.key, {
                              quantityInput: event.target.value,
                              quantityError: null,
                            })
                          }
                          className="min-w-0 flex-1 rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2 disabled:opacity-60"
                        />
                        <span className="shrink-0 text-sm text-stone-500">
                          {unitLabel}
                        </span>
                      </div>
                      {item.quantityError ? (
                        <p className="mt-1.5 text-sm text-red-700">
                          {item.quantityError}
                        </p>
                      ) : null}
                    </div>

                    <div>
                      <label
                        htmlFor={`price-${item.key}`}
                        className="mb-1.5 block text-sm font-medium text-stone-700"
                      >
                        Cena povrata
                      </label>
                      <div className="flex items-center gap-2">
                        <input
                          id={`price-${item.key}`}
                          inputMode="decimal"
                          disabled={submitting}
                          value={item.unitPriceInput}
                          onChange={(event) =>
                            updateItem(item.key, {
                              unitPriceInput: event.target.value,
                              unitPriceError: null,
                            })
                          }
                          className="min-w-0 flex-1 rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2 disabled:opacity-60"
                        />
                        <span className="shrink-0 text-sm text-stone-500">
                          RSD/{unitLabel}
                        </span>
                      </div>
                      <p className="mt-1.5 text-xs text-stone-500">
                        Po potrebi promenite cenu prema originalnom računu.
                      </p>
                      {item.unitPriceError ? (
                        <p className="mt-1.5 text-sm text-red-700">
                          {item.unitPriceError}
                        </p>
                      ) : null}
                    </div>
                  </div>
                ) : null}

                {item.product &&
                quantity != null &&
                unitPrice != null &&
                lineTotal != null ? (
                  <div className="mt-4 rounded-xl border border-stone-100 bg-stone-50 px-3 py-3 text-sm text-stone-700">
                    <p className="text-stone-500">Vrednost povrata</p>
                    <p className="mt-1 tabular-nums">
                      {formatQuantity(quantity)} {unitLabel} ×{" "}
                      {formatMoney(unitPrice)}
                    </p>
                    <p className="mt-1 font-semibold tabular-nums text-stone-900">
                      = {formatMoney(lineTotal)}
                    </p>
                  </div>
                ) : null}
              </div>
            );
          })}
        </div>

        <button
          type="button"
          disabled={submitting}
          onClick={() => setItems((current) => [...current, createEmptyItem()])}
          className="inline-flex min-h-11 w-full items-center justify-center rounded-xl border border-dashed border-stone-300 px-4 text-sm font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-60 sm:w-auto"
        >
          + Dodaj još artikal
        </button>

        <div>
          <label
            htmlFor="sales-return-description"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            Opis
          </label>
          <textarea
            id="sales-return-description"
            rows={3}
            maxLength={DESCRIPTION_MAX}
            disabled={submitting}
            value={description}
            onChange={(event) => setDescription(event.target.value)}
            placeholder="Npr. Marko Marković - višak materijala nakon radova"
            className="w-full resize-y rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2 disabled:opacity-60"
          />
          <p className="mt-1 text-xs text-stone-400">
            {description.length}/{DESCRIPTION_MAX}
          </p>
        </div>

        <fieldset className="space-y-3">
          <legend className="text-sm font-medium text-stone-700">
            Da li je kupcu vraćen novac?
          </legend>
          <label className="flex cursor-pointer items-start gap-3 rounded-xl border border-stone-200 px-3 py-3 hover:bg-stone-50">
            <input
              type="radio"
              name="cashRefunded"
              className="mt-1"
              checked={cashRefunded === true}
              disabled={submitting}
              onChange={() => {
                setCashRefunded(true);
                setCashError(null);
              }}
            />
            <span>
              <span className="block text-sm font-medium text-stone-900">
                Da, novac je isplaćen kupcu
              </span>
              {cashRefunded === true ? (
                <span className="mt-1 block text-sm text-stone-600">
                  Za isplatu kupcu: {formatMoney(grandTotal)}. Ovaj iznos će biti
                  evidentiran kao finansijski povrat.
                </span>
              ) : null}
            </span>
          </label>
          <label className="flex cursor-pointer items-start gap-3 rounded-xl border border-stone-200 px-3 py-3 hover:bg-stone-50">
            <input
              type="radio"
              name="cashRefunded"
              className="mt-1"
              checked={cashRefunded === false}
              disabled={submitting}
              onChange={() => {
                setCashRefunded(false);
                setCashError(null);
              }}
            />
            <span>
              <span className="block text-sm font-medium text-stone-900">
                Ne, bez isplate novca
              </span>
              {cashRefunded === false ? (
                <span className="mt-1 block text-sm text-stone-600">
                  Roba će biti vraćena na lager bez finansijske isplate.
                </span>
              ) : null}
            </span>
          </label>
          {cashError ? (
            <p className="text-sm text-red-700">{cashError}</p>
          ) : null}
        </fieldset>
      </form>
    </Modal>
  );
}
