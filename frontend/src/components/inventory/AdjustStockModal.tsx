"use client";

import { FormEvent, useMemo, useState } from "react";

import { Modal } from "@/components/ui/Modal";
import {
  adjustInventoryStock,
  getApiBusinessMessage,
} from "@/lib/inventory-api";
import { formatQuantity } from "@/lib/format";
import { InventoryProductRow } from "@/types/inventory";

function parseQuantityInput(value: string): number | null {
  const normalized = value.trim().replace(",", ".");
  if (!normalized) {
    return null;
  }
  const parsed = Number(normalized);
  if (!Number.isFinite(parsed) || parsed < 0) {
    return null;
  }
  return parsed;
}

function AdjustStockForm({
  product,
  onClose,
  onSuccess,
}: {
  product: InventoryProductRow;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [actualInput, setActualInput] = useState(() =>
    String(product.stockQuantity).replace(".", ","),
  );
  const [reason, setReason] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const actualQuantity = parseQuantityInput(actualInput);
  const change =
    actualQuantity != null ? actualQuantity - product.stockQuantity : null;

  const preview = useMemo(() => {
    if (actualQuantity == null) {
      return null;
    }
    return {
      current: product.stockQuantity,
      next: actualQuantity,
      change,
    };
  }, [product.stockQuantity, actualQuantity, change]);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    if (submitting) {
      return;
    }
    if (actualQuantity == null) {
      setError("Unesite ispravno stvarno stanje.");
      return;
    }

    setSubmitting(true);
    setError(null);
    try {
      await adjustInventoryStock({
        productID: product.id,
        newQuantity: actualQuantity,
        note: reason.trim() || undefined,
      });
      onSuccess();
      onClose();
    } catch (err) {
      setError(
        getApiBusinessMessage(err, "Korekcija lagera nije uspjela."),
      );
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="rounded-xl border border-stone-200 bg-stone-50 px-3 py-3 text-sm">
        <p className="text-stone-600">
          Trenutno stanje:{" "}
          <span className="font-semibold text-stone-900">
            {formatQuantity(product.stockQuantity)} {product.unit}
          </span>
        </p>
      </div>

      <div>
        <label
          htmlFor="actual-stock"
          className="mb-1.5 block text-sm font-medium text-stone-700"
        >
          Stvarno stanje
        </label>
        <div className="flex items-center gap-2">
          <input
            id="actual-stock"
            inputMode="decimal"
            value={actualInput}
            disabled={submitting}
            onChange={(event) => setActualInput(event.target.value)}
            className="min-w-0 flex-1 rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2 disabled:opacity-60"
          />
          <span className="shrink-0 text-sm text-stone-500">{product.unit}</span>
        </div>
      </div>

      {preview ? (
        <div className="rounded-xl border border-stone-200 bg-white px-3 py-3 text-sm text-stone-700">
          <p>
            Trenutno:{" "}
            <span className="font-medium tabular-nums">
              {formatQuantity(preview.current)} {product.unit}
            </span>
          </p>
          <p className="mt-1">
            Novo stanje:{" "}
            <span className="font-medium tabular-nums">
              {formatQuantity(preview.next)} {product.unit}
            </span>
          </p>
          <p className="mt-1">
            Promjena:{" "}
            <span
              className={`font-semibold tabular-nums ${
                (preview.change ?? 0) < 0
                  ? "text-red-700"
                  : (preview.change ?? 0) > 0
                    ? "text-emerald-700"
                    : "text-stone-700"
              }`}
            >
              {preview.change != null && preview.change > 0 ? "+" : ""}
              {formatQuantity(preview.change ?? 0)} {product.unit}
            </span>
          </p>
        </div>
      ) : null}

      <div>
        <label
          htmlFor="adjust-reason"
          className="mb-1.5 block text-sm font-medium text-stone-700"
        >
          Razlog korekcije
        </label>
        <input
          id="adjust-reason"
          value={reason}
          disabled={submitting}
          onChange={(event) => setReason(event.target.value)}
          placeholder="npr. Fizički popis"
          className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2 disabled:opacity-60"
        />
      </div>

      {error ? (
        <p className="rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
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
          disabled={submitting || actualQuantity == null}
          className="inline-flex min-h-11 items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white hover:bg-stone-800 disabled:opacity-60"
        >
          {submitting ? "Spremanje…" : "Sačuvaj korekciju"}
        </button>
      </div>
    </form>
  );
}

export function AdjustStockModal({
  open,
  product,
  productOptions = [],
  onClose,
  onSuccess,
}: {
  open: boolean;
  product: InventoryProductRow | null;
  productOptions?: InventoryProductRow[];
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [selectedProductId, setSelectedProductId] = useState<number | null>(
    product?.id ?? productOptions[0]?.id ?? null,
  );

  const activeProduct =
    product ??
    productOptions.find((item) => item.id === selectedProductId) ??
    null;

  if (!open) {
    return null;
  }

  return (
    <Modal
      open={open}
      title="Korekcija lagera"
      description={activeProduct?.name}
      onClose={onClose}
    >
      {!product && productOptions.length > 0 ? (
        <div className="mb-4">
          <label
            htmlFor="adjust-product"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            Proizvod
          </label>
          <select
            id="adjust-product"
            value={selectedProductId ?? ""}
            onChange={(event) =>
              setSelectedProductId(Number(event.target.value))
            }
            className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2"
          >
            {productOptions.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </select>
        </div>
      ) : null}

      {!activeProduct ? (
        <p className="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">
          Odaberite proizvod iz liste lagera.
        </p>
      ) : (
        <AdjustStockForm
          key={activeProduct.id}
          product={activeProduct}
          onClose={onClose}
          onSuccess={onSuccess}
        />
      )}
    </Modal>
  );
}
