"use client";

import { formatMoney } from "@/lib/format";
import {
  detectPricingMode,
  previewCalculatedSalePrice,
} from "@/lib/product-pricing";

export type PricingFieldValues = {
  purchasePrice: string;
  marginPercent: string;
  vatPercent: string;
  salePrice: string;
};

function parseAmount(value: string): number {
  const normalized = value.trim().replace(",", ".");
  if (!normalized) {
    return 0;
  }
  const parsed = Number(normalized);
  return Number.isFinite(parsed) ? parsed : 0;
}

export function PricingFields({
  values,
  onChange,
  disabled = false,
}: {
  values: PricingFieldValues;
  onChange: (patch: Partial<PricingFieldValues>) => void;
  disabled?: boolean;
}) {
  const purchase = parseAmount(values.purchasePrice);
  const margin = parseAmount(values.marginPercent);
  const vat = parseAmount(values.vatPercent);
  const mode = detectPricingMode(margin, vat);
  const preview = previewCalculatedSalePrice(purchase, margin, vat);
  const calculated = mode === "calculated";

  return (
    <section className="space-y-3 rounded-2xl border border-stone-200 bg-stone-50/70 p-4">
      <div>
        <h3 className="text-sm font-semibold text-stone-900">Cijena</h3>
        <p className="mt-0.5 text-xs text-stone-500">
          Režim:{" "}
          <span className="font-medium text-[#8a6a45]">
            {calculated ? "Automatski obračun" : "Ručni unos"}
          </span>
          {calculated
            ? " — prodajna cijena se računa iz nabavne, marže i PDV-a."
            : " — unesite prodajnu cijenu ručno."}
        </p>
      </div>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <div>
          <label
            htmlFor="purchase-price"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            Nabavna cijena {calculated ? "*" : ""}
          </label>
          <input
            id="purchase-price"
            type="text"
            inputMode="decimal"
            disabled={disabled}
            value={values.purchasePrice}
            onChange={(event) => onChange({ purchasePrice: event.target.value })}
            className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
            placeholder="0"
          />
        </div>
        <div>
          <label
            htmlFor="margin-percent"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            Marža %
          </label>
          <input
            id="margin-percent"
            type="text"
            inputMode="decimal"
            disabled={disabled}
            value={values.marginPercent}
            onChange={(event) => onChange({ marginPercent: event.target.value })}
            className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
            placeholder="0"
          />
        </div>
        <div>
          <label
            htmlFor="vat-percent"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            PDV %
          </label>
          <input
            id="vat-percent"
            type="text"
            inputMode="decimal"
            disabled={disabled}
            value={values.vatPercent}
            onChange={(event) => onChange({ vatPercent: event.target.value })}
            className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
            placeholder="0"
          />
        </div>
      </div>

      <div>
        <label
          htmlFor="sale-price"
          className="mb-1.5 block text-sm font-medium text-stone-700"
        >
          Prodajna cijena {!calculated ? "*" : ""}
        </label>
        <input
          id="sale-price"
          type="text"
          inputMode="decimal"
          disabled={disabled || calculated}
          readOnly={calculated}
          value={
            calculated
              ? purchase > 0
                ? String(preview.finalSalePrice)
                : ""
              : values.salePrice
          }
          onChange={(event) => onChange({ salePrice: event.target.value })}
          className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:bg-stone-100 disabled:opacity-80"
          placeholder={calculated ? "Automatski" : "0"}
        />
      </div>

      {calculated && purchase > 0 ? (
        <div className="rounded-xl border border-[#c4a484]/35 bg-[#faf6f1] px-3 py-3 text-sm text-stone-700">
          <p className="font-medium text-stone-900">Pregled obračuna</p>
          <ul className="mt-2 space-y-1 text-xs sm:text-sm">
            <li>
              Nakon marže:{" "}
              <span className="font-medium tabular-nums">
                {formatMoney(preview.priceAfterMargin)}
              </span>
            </li>
            <li>
              Sa PDV-om (2 dec.):{" "}
              <span className="font-medium tabular-nums">
                {formatMoney(preview.rawSalePrice)}
              </span>
            </li>
            <li>
              Finalna (zaokruženo na 10):{" "}
              <span className="font-medium tabular-nums text-[#8a6a45]">
                {formatMoney(preview.finalSalePrice)}
              </span>
            </li>
          </ul>
        </div>
      ) : null}
    </section>
  );
}
