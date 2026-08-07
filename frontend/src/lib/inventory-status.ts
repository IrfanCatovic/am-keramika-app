export type StockStatusKind = "in_stock" | "low" | "out";

export function getStockStatus(
  stockQuantity: number,
  minStockQuantity: number,
): StockStatusKind {
  if (stockQuantity <= 0) {
    return "out";
  }
  if (stockQuantity <= minStockQuantity) {
    return "low";
  }
  return "in_stock";
}

export function stockStatusLabel(status: StockStatusKind): string {
  switch (status) {
    case "out":
      return "Nema na stanju";
    case "low":
      return "Nizak lager";
    default:
      return "Na stanju";
  }
}

export function stockStatusClassName(status: StockStatusKind): string {
  switch (status) {
    case "out":
      return "border-red-200 bg-red-50 text-red-800";
    case "low":
      return "border-amber-200 bg-amber-50 text-amber-900";
    default:
      return "border-stone-200 bg-stone-50 text-stone-700";
  }
}

export function movementTypeLabel(type: string): string {
  switch (type) {
    case "sale":
      return "Prodaja";
    case "return":
      return "Povrat";
    case "adjust":
      return "Korekcija";
    case "in":
      return "Ulaz";
    default:
      return type;
  }
}

export function signedMovementQuantity(type: string, quantity: number): number {
  switch (type) {
    case "sale":
      return -Math.abs(quantity);
    case "return":
    case "in":
      return Math.abs(quantity);
    default:
      return quantity;
  }
}

export function formatSignedQuantity(value: number): string {
  const prefix = value > 0 ? "+" : "";
  return `${prefix}${value}`;
}
