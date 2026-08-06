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
