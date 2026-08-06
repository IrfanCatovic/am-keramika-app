/**
 * Centralni formatteri za cijelu AM Keramika aplikaciju.
 */

const rsdNumberFormatInteger = new Intl.NumberFormat("sr-RS", {
  minimumFractionDigits: 0,
  maximumFractionDigits: 0,
});

const rsdNumberFormatDecimal = new Intl.NumberFormat("sr-RS", {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});

const quantityFormat = new Intl.NumberFormat("sr-RS", {
  maximumFractionDigits: 3,
});

function hasMeaningfulFraction(amount: number): boolean {
  return Math.round(amount * 100) % 100 !== 0;
}

/**
 * Formatira iznos u RSD.
 * Primjeri: 5800 → "5.800 RSD", 1500.5 → "1.500,50 RSD"
 */
export function formatMoney(amount: number): string {
  const value = Number.isFinite(amount) ? amount : 0;
  const formatted = hasMeaningfulFraction(value)
    ? rsdNumberFormatDecimal.format(value)
    : rsdNumberFormatInteger.format(Math.round(value));
  return `${formatted} RSD`;
}

export function formatQuantity(value: number): string {
  const amount = Number.isFinite(value) ? value : 0;
  return quantityFormat.format(amount);
}

export function formatCount(value: number): string {
  return rsdNumberFormatInteger.format(Number.isFinite(value) ? value : 0);
}
