"use client";

import { InvoiceCartItem } from "@/components/invoices/pos/InvoiceCartItem";
import { InvoiceFormLine } from "@/types/invoice";

export function InvoiceCart({
  lines,
  lineErrors,
  highlightedProductID,
  onQuantityChange,
  onRemove,
  emptyLabel = "Račun je prazan. Dodajte proizvod pretragom.",
  className = "",
}: {
  lines: InvoiceFormLine[];
  lineErrors: Record<number, string>;
  highlightedProductID?: number | null;
  onQuantityChange: (productID: number, quantity: number) => void;
  onRemove: (productID: number) => void;
  emptyLabel?: string;
  className?: string;
}) {
  if (lines.length === 0) {
    return (
      <div
        className={`rounded-xl border border-dashed border-stone-200 bg-stone-50/80 px-3 py-8 text-center text-sm text-stone-500 ${className}`}
      >
        {emptyLabel}
      </div>
    );
  }

  return (
    <ul className={`space-y-2 ${className}`}>
      {lines.map((line) => (
        <li key={line.productID}>
          <InvoiceCartItem
            line={line}
            error={lineErrors[line.productID]}
            highlighted={highlightedProductID === line.productID}
            onQuantityChange={(quantity) =>
              onQuantityChange(line.productID, quantity)
            }
            onRemove={() => onRemove(line.productID)}
          />
        </li>
      ))}
    </ul>
  );
}
