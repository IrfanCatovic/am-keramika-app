"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";

import { CustomerSelector } from "@/components/customers/CustomerSelector";
import { InvoiceItemRow } from "@/components/invoices/InvoiceItemRow";
import { InvoiceSummary } from "@/components/invoices/InvoiceSummary";
import { ProductSelector } from "@/components/invoices/ProductSelector";
import { fetchCustomer } from "@/lib/customers-api";
import {
  createInvoice,
  getApiBusinessMessage,
} from "@/lib/invoices-api";
import { CustomerListItem } from "@/types/customer";
import { InvoiceFormLine } from "@/types/invoice";
import { Product } from "@/types/product";

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

export function InvoiceForm({
  initialCustomerID,
}: {
  initialCustomerID?: number | null;
}) {
  const router = useRouter();
  const [customerMode, setCustomerMode] = useState<CustomerMode>(
    initialCustomerID ? "customer" : "cash",
  );
  const [customer, setCustomer] = useState<CustomerListItem | null>(null);
  const [customerPrefillError, setCustomerPrefillError] = useState<string | null>(
    null,
  );
  const [lines, setLines] = useState<InvoiceFormLine[]>([]);
  const [lineErrors, setLineErrors] = useState<Record<number, string>>({});
  const [selectorOpen, setSelectorOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const dirty = lines.length > 0 || (customerMode === "customer" && customer);

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

  const selectedQtyByProduct = useMemo(() => {
    const map = new Map<number, number>();
    for (const line of lines) {
      map.set(line.productID, line.quantity);
    }
    return map;
  }, [lines]);

  function addProduct(product: Product) {
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
            ? { ...line, quantity: nextQty, stockQuantity: product.stockQuantity }
            : line,
        );
      }
      if (product.stockQuantity <= 0) {
        return current;
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
  }

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

  async function handleSubmit() {
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
      router.replace(`/invoices/${invoice.id}`);
    } catch (err) {
      const message = getApiBusinessMessage(
        err,
        "Kreiranje računa nije uspjelo.",
      );
      setError(message);
      if (message.toLowerCase().includes("nema dovoljno")) {
        setLineErrors((current) => {
          const next = { ...current };
          for (const line of lines) {
            next[line.productID] =
              next[line.productID] ??
              "Provjerite lager — backend je odbio količinu.";
          }
          return next;
        });
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="min-w-0 space-y-4 sm:space-y-5">
      <header>
        <Link
          href="/invoices"
          className="text-sm font-medium text-[#8a6a45] hover:text-stone-900"
        >
          ← Nazad na račune
        </Link>
        <h1 className="mt-2 text-2xl font-semibold tracking-tight text-stone-900">
          Novi račun
        </h1>
        <p className="mt-1 text-sm text-stone-500">
          Preview cijena je informativan — server radi konačan obračun.
        </p>
      </header>

      <section className="rounded-2xl border border-stone-200 bg-white p-4 sm:p-5">
        <p className="text-sm font-medium text-stone-700">Tip računa</p>
        <div className="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-2">
          <button
            type="button"
            onClick={() => {
              setCustomerMode("cash");
              setCustomer(null);
              setCustomerPrefillError(null);
              setError(null);
            }}
            className={`rounded-xl border px-4 py-3 text-left text-sm transition ${
              customerMode === "cash"
                ? "border-stone-900 bg-stone-900 text-white"
                : "border-stone-200 bg-white text-stone-700 hover:bg-stone-50"
            }`}
          >
            <span className="font-medium">Gotovinska prodaja</span>
            <span className="mt-1 block text-xs opacity-80">Bez kupca</span>
          </button>
          <button
            type="button"
            onClick={() => {
              setCustomerMode("customer");
              setError(null);
            }}
            className={`rounded-xl border px-4 py-3 text-left text-sm transition ${
              customerMode === "customer"
                ? "border-stone-900 bg-stone-900 text-white"
                : "border-stone-200 bg-white text-stone-700 hover:bg-stone-50"
            }`}
          >
            <span className="font-medium">Račun za kupca</span>
            <span className="mt-1 block text-xs opacity-80">
              Samo aktivni kupci
            </span>
          </button>
        </div>

        {customerMode === "customer" ? (
          <div className="mt-4">
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
      </section>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1.4fr)_minmax(18rem,0.8fr)] lg:items-start">
        <div className="space-y-3">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <h2 className="text-base font-semibold text-stone-900">Stavke</h2>
            <button
              type="button"
              onClick={() => setSelectorOpen(true)}
              className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-800 hover:bg-stone-50"
            >
              Dodaj proizvod
            </button>
          </div>

          {lines.length === 0 ? (
            <div className="rounded-2xl border border-dashed border-stone-300 bg-white px-4 py-10 text-center text-sm text-stone-500">
              Još nema stavki. Dodajte proizvod da sastavite račun.
            </div>
          ) : (
            <ul className="space-y-3">
              {lines.map((line) => (
                <li key={line.productID}>
                  <InvoiceItemRow
                    line={line}
                    error={lineErrors[line.productID]}
                    onQuantityChange={(quantity) =>
                      updateQuantity(line.productID, quantity)
                    }
                    onRemove={() => removeLine(line.productID)}
                  />
                </li>
              ))}
            </ul>
          )}
        </div>

        <InvoiceSummary
          lines={lines}
          customerMode={customerMode}
          customerName={customer?.name ?? null}
          submitting={submitting}
          error={error}
          onSubmit={() => void handleSubmit()}
        />
      </div>

      <ProductSelector
        open={selectorOpen}
        onClose={() => setSelectorOpen(false)}
        onSelect={addProduct}
        excludeFullySelected={selectedQtyByProduct}
      />
    </div>
  );
}
