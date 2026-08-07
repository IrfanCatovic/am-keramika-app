import { PricingMode } from "@/types/product";
import { UserRole } from "@/types/auth";

export function canViewSensitivePricing(role: UserRole): boolean {
  return role === "developer" || role === "sef" || role === "menadzer";
}

export function detectPricingMode(
  marginPercent: number,
  vatPercent: number,
): PricingMode {
  return marginPercent > 0 || vatPercent > 0 ? "calculated" : "manual";
}

export interface PricingPreview {
  mode: PricingMode;
  priceAfterMargin: number;
  rawSalePrice: number;
  finalSalePrice: number;
}

/**
 * Frontend-only preview. Backend remains source of truth.
 * raw = purchase * (1+margin/100) * (1+vat/100)
 * rounded2 = Math.round(raw*100)/100
 * final = (rounded2 % 10 < 1e-9) ? rounded2 : Math.ceil(rounded2/10)*10
 */
export function previewCalculatedSalePrice(
  purchasePrice: number,
  marginPercent: number,
  vatPercent: number,
): PricingPreview {
  const purchase = Number.isFinite(purchasePrice) ? purchasePrice : 0;
  const margin = Number.isFinite(marginPercent) ? marginPercent : 0;
  const vat = Number.isFinite(vatPercent) ? vatPercent : 0;
  const mode = detectPricingMode(margin, vat);

  if (mode === "manual" || purchase <= 0) {
    return {
      mode: "manual",
      priceAfterMargin: 0,
      rawSalePrice: 0,
      finalSalePrice: 0,
    };
  }

  const priceAfterMargin = purchase * (1 + margin / 100);
  const raw = priceAfterMargin * (1 + vat / 100);
  const rounded2 = Math.round(raw * 100) / 100;
  const final =
    rounded2 % 10 < 1e-9 ? rounded2 : Math.ceil(rounded2 / 10) * 10;

  return {
    mode: "calculated",
    priceAfterMargin: Math.round(priceAfterMargin * 100) / 100,
    rawSalePrice: rounded2,
    finalSalePrice: final,
  };
}

/** Matches backend RoundUpToTen. */
export function roundUpToTen(raw: number): number {
  const twoDecimals = Math.round(raw * 100) / 100;
  const remainder = twoDecimals % 10;
  if (remainder < 1e-9 || remainder > 10 - 1e-9) {
    return Math.round(twoDecimals * 100) / 100;
  }
  return Math.ceil(twoDecimals / 10) * 10;
}

/** Frontend preview of effective sale price; backend remains source of truth. */
export function getEffectiveSalePrice(
  salePrice: number,
  isOnSale: boolean,
  discountPercent: number,
): number {
  if (!isOnSale || !(discountPercent > 0)) {
    return salePrice;
  }
  return roundUpToTen(salePrice * (1 - discountPercent / 100));
}

export function getDiscountedRawSalePrice(
  salePrice: number,
  discountPercent: number,
): number {
  return Math.round(salePrice * (1 - discountPercent / 100) * 100) / 100;
}

/** Prefer API effectiveSalePrice; fall back to local preview. */
export function resolveProductUnitPrice(product: {
  salePrice: number;
  effectiveSalePrice?: number;
  isOnSale?: boolean;
  discountPercent?: number;
}): number {
  if (
    typeof product.effectiveSalePrice === "number" &&
    Number.isFinite(product.effectiveSalePrice)
  ) {
    return product.effectiveSalePrice;
  }
  return getEffectiveSalePrice(
    product.salePrice,
    Boolean(product.isOnSale),
    product.discountPercent ?? 0,
  );
}
