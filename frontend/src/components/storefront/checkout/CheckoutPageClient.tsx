"use client";

import Link from "next/link";
import { FormEvent, useEffect, useMemo, useRef, useState } from "react";

import { CheckoutSuccess } from "@/components/storefront/checkout/CheckoutSuccess";
import { CheckoutSummary } from "@/components/storefront/checkout/CheckoutSummary";
import { useCart } from "@/components/storefront/cart/CartProvider";
import { checkPublicProductAvailability } from "@/lib/public-availability-api";
import { fetchPublicProductBySlug } from "@/lib/public-catalog-api";
import {
  clearCheckoutDraft,
  createPublicOrder,
  PublicOrderError,
  readCheckoutDraft,
  writeCheckoutDraft,
} from "@/lib/public-order-api";
import type { CartItem } from "@/types/cart";
import type { CheckoutDraft } from "@/types/online-order";

type LineMeta = {
  unavailable: boolean;
  insufficient: boolean;
};

const emptyDraft: CheckoutDraft = {
  firstName: "",
  lastName: "",
  phone: "",
  city: "",
  address: "",
  email: "",
  note: "",
};

function validateDraft(draft: CheckoutDraft): string | null {
  if (!draft.firstName.trim()) return "Unesite ime.";
  if (!draft.lastName.trim()) return "Unesite prezime.";
  const digits = draft.phone.replace(/\D/g, "");
  if (digits.length < 6) return "Unesite ispravan broj telefona.";
  if (!draft.city.trim()) return "Unesite grad.";
  if (!draft.address.trim()) return "Unesite adresu.";
  if (draft.email.trim()) {
    const email = draft.email.trim();
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      return "Email adresa nije ispravna.";
    }
  }
  return null;
}

export function CheckoutPageClient() {
  const { items, hydrated, clearCart, updateItemSnapshot } = useCart();
  const [draft, setDraft] = useState<CheckoutDraft>(emptyDraft);
  const [draftReady, setDraftReady] = useState(false);
  const [honeypot, setHoneypot] = useState("");
  const [metaById, setMetaById] = useState<Record<number, LineMeta>>({});
  const [refreshing, setRefreshing] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [success, setSuccess] = useState<{
    id: number;
    totalAmount: number;
  } | null>(null);
  const syncedKey = useRef("");
  const submitLock = useRef(false);

  const productKey = items.map((i) => `${i.productId}:${i.quantity}`).join("|");

  useEffect(() => {
    const frame = window.requestAnimationFrame(() => {
      setDraft(readCheckoutDraft() ?? emptyDraft);
      setDraftReady(true);
    });
    return () => window.cancelAnimationFrame(frame);
  }, []);

  useEffect(() => {
    if (!draftReady) return;
    writeCheckoutDraft(draft);
  }, [draft, draftReady]);

  useEffect(() => {
    if (!hydrated) return;
    if (!productKey) {
      syncedKey.current = "";
      return;
    }
    if (syncedKey.current === productKey) return;
    syncedKey.current = productKey;

    let cancelled = false;

    async function refresh() {
      setRefreshing(true);
      const snapshot = items;
      const nextMeta: Record<number, LineMeta> = {};

      await Promise.all(
        snapshot.map(async (item) => {
          try {
            const product = await fetchPublicProductBySlug(item.slug);
            if (product.id !== item.productId) {
              nextMeta[item.productId] = {
                unavailable: true,
                insufficient: false,
              };
              return;
            }

            const patch: Partial<CartItem> = {
              name: product.name,
              slug: product.slug,
              unit: product.unit,
              imageUrl: product.primaryImage?.url ?? item.imageUrl,
              salePrice: product.salePrice,
              effectiveSalePrice: product.effectiveSalePrice,
              isOnSale: product.isOnSale,
              discountPercent: product.discountPercent,
              categoryName: product.category?.name,
              groupName: product.group?.name,
            };
            updateItemSnapshot(item.productId, patch);

            let insufficient = false;
            if (!product.inStock) {
              insufficient = true;
            } else {
              try {
                const avail = await checkPublicProductAvailability(
                  product.id,
                  item.quantity,
                );
                insufficient = !avail.available;
              } catch {
                insufficient = false;
              }
            }

            nextMeta[item.productId] = {
              unavailable: false,
              insufficient,
            };
          } catch {
            nextMeta[item.productId] = {
              unavailable: true,
              insufficient: false,
            };
          }
        }),
      );

      if (!cancelled) {
        setMetaById(nextMeta);
        setRefreshing(false);
      }
    }

    void refresh();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [hydrated, productKey]);

  const hasBlockingItems = useMemo(
    () =>
      items.some(
        (item) =>
          metaById[item.productId]?.unavailable ||
          metaById[item.productId]?.insufficient,
      ),
    [items, metaById],
  );

  const subtotal = useMemo(
    () =>
      items.reduce(
        (sum, item) => sum + item.effectiveSalePrice * item.quantity,
        0,
      ),
    [items],
  );

  function patchDraft(field: keyof CheckoutDraft, value: string) {
    setDraft((prev) => ({ ...prev, [field]: value }));
  }

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    if (submitLock.current || submitting || success) return;

    setFormError(null);
    const validation = validateDraft(draft);
    if (validation) {
      setFormError(validation);
      return;
    }
    if (!hydrated || items.length === 0) {
      setFormError("Korpa je prazna.");
      return;
    }
    if (hasBlockingItems) {
      setFormError(
        "Jedan od proizvoda iz korpe više nije dostupan ili nema dovoljno na stanju.",
      );
      return;
    }

    submitLock.current = true;
    setSubmitting(true);
    try {
      const response = await createPublicOrder({
        firstName: draft.firstName.trim(),
        lastName: draft.lastName.trim(),
        phone: draft.phone.trim(),
        city: draft.city.trim(),
        address: draft.address.trim(),
        email: draft.email.trim() || undefined,
        note: draft.note.trim() || undefined,
        website: honeypot,
        items: items.map((item) => ({
          productID: item.productId,
          quantity: item.quantity,
        })),
      });

      clearCart();
      clearCheckoutDraft();
      setSuccess({ id: response.id, totalAmount: response.totalAmount });
    } catch (err) {
      const message =
        err instanceof PublicOrderError
          ? err.message
          : "Narudžbinu trenutno nije moguće poslati. Pokušajte ponovo.";
      setFormError(message);
      if (err instanceof PublicOrderError && err.productID) {
        setMetaById((prev) => ({
          ...prev,
          [err.productID!]: {
            unavailable: err.code === "unavailable",
            insufficient: err.code === "insufficient_stock",
          },
        }));
      }
    } finally {
      submitLock.current = false;
      setSubmitting(false);
    }
  }

  if (success) {
    return (
      <CheckoutSuccess orderId={success.id} totalAmount={success.totalAmount} />
    );
  }

  if (!hydrated || !draftReady) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
        <div className="h-8 w-56 animate-pulse rounded bg-stone-200" />
        <div className="mt-8 h-64 animate-pulse rounded-xl bg-stone-100" />
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className="mx-auto max-w-xl px-4 py-20 text-center sm:px-6">
        <p className="text-[11px] uppercase tracking-[0.18em] text-[#8a6a45]">
          Narudžbina
        </p>
        <h1 className="mt-3 font-[family-name:var(--font-storefront-display)] text-3xl text-stone-900">
          Korpa je prazna
        </h1>
        <p className="mt-3 text-sm text-stone-500">
          Dodajte proizvode u korpu da biste nastavili sa narudžbinom.
        </p>
        <Link
          href="/proizvodi"
          className="mt-8 inline-flex min-h-11 items-center rounded-full bg-[#141311] px-6 text-sm text-white"
        >
          Pogledajte proizvode
        </Link>
      </div>
    );
  }

  const fieldClass =
    "mt-1.5 w-full rounded-xl border border-stone-300/80 bg-white px-3.5 py-2.5 text-sm text-stone-900 outline-none transition placeholder:text-stone-400 focus:border-stone-400 focus:ring-2 focus:ring-[rgba(138,106,69,0.18)]";

  return (
    <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
      <div className="mb-8 max-w-2xl">
        <p className="text-[11px] uppercase tracking-[0.2em] text-[#8a6a45]">
          Završetak narudžbine
        </p>
        <h1 className="mt-2 font-[family-name:var(--font-storefront-display)] text-3xl text-stone-900 sm:text-4xl">
          Podaci za narudžbinu
        </h1>
        <p className="mt-3 text-sm text-stone-500 sm:text-base">
          Unesite podatke kako bismo vas kontaktirali i dogovorili isporuku.
        </p>
        {refreshing ? (
          <p className="mt-2 text-xs text-stone-400">Ažuriranje proizvoda…</p>
        ) : null}
      </div>

      {hasBlockingItems ? (
        <div className="mb-6 rounded-xl border border-stone-200 bg-white px-4 py-4 text-sm text-stone-600">
          <p>
            Jedan ili više proizvoda u korpi trenutno nije dostupan za
            narudžbinu.
          </p>
          <Link
            href="/korpa"
            className="mt-2 inline-flex font-medium text-stone-900 underline-offset-4 hover:underline"
          >
            Vrati se u korpu
          </Link>
        </div>
      ) : null}

      <div className="grid gap-10 lg:grid-cols-[minmax(0,1.05fr)_360px] xl:grid-cols-[minmax(0,1.1fr)_380px]">
        <form onSubmit={onSubmit} className="relative space-y-5" noValidate>
          <div className="grid gap-5 sm:grid-cols-2">
            <div>
              <label htmlFor="checkout-firstName" className="text-sm text-stone-700">
                Ime *
              </label>
              <input
                id="checkout-firstName"
                name="given-name"
                autoComplete="given-name"
                required
                value={draft.firstName}
                onChange={(e) => patchDraft("firstName", e.target.value)}
                className={fieldClass}
              />
            </div>
            <div>
              <label htmlFor="checkout-lastName" className="text-sm text-stone-700">
                Prezime *
              </label>
              <input
                id="checkout-lastName"
                name="family-name"
                autoComplete="family-name"
                required
                value={draft.lastName}
                onChange={(e) => patchDraft("lastName", e.target.value)}
                className={fieldClass}
              />
            </div>
          </div>

          <div>
            <label htmlFor="checkout-phone" className="text-sm text-stone-700">
              Broj telefona *
            </label>
            <input
              id="checkout-phone"
              name="tel"
              type="tel"
              autoComplete="tel"
              required
              value={draft.phone}
              onChange={(e) => patchDraft("phone", e.target.value)}
              className={fieldClass}
            />
          </div>

          <div>
            <label htmlFor="checkout-city" className="text-sm text-stone-700">
              Grad *
            </label>
            <input
              id="checkout-city"
              name="address-level2"
              autoComplete="address-level2"
              required
              value={draft.city}
              onChange={(e) => patchDraft("city", e.target.value)}
              className={fieldClass}
            />
          </div>

          <div>
            <label htmlFor="checkout-address" className="text-sm text-stone-700">
              Adresa *
            </label>
            <input
              id="checkout-address"
              name="street-address"
              autoComplete="street-address"
              required
              value={draft.address}
              onChange={(e) => patchDraft("address", e.target.value)}
              className={fieldClass}
            />
          </div>

          <div>
            <label htmlFor="checkout-email" className="text-sm text-stone-700">
              Email
            </label>
            <input
              id="checkout-email"
              name="email"
              type="email"
              autoComplete="email"
              value={draft.email}
              onChange={(e) => patchDraft("email", e.target.value)}
              className={fieldClass}
            />
          </div>

          <div>
            <label htmlFor="checkout-note" className="text-sm text-stone-700">
              Napomena
            </label>
            <textarea
              id="checkout-note"
              name="note"
              rows={4}
              value={draft.note}
              onChange={(e) => patchDraft("note", e.target.value)}
              className={`${fieldClass} resize-y`}
            />
          </div>

          {/* Honeypot */}
          <div className="absolute -left-[9999px] h-0 w-0 overflow-hidden opacity-0" aria-hidden>
            <label htmlFor="checkout-website">Website</label>
            <input
              id="checkout-website"
              name="website"
              tabIndex={-1}
              autoComplete="off"
              value={honeypot}
              onChange={(e) => setHoneypot(e.target.value)}
            />
          </div>

          <div className="space-y-3 rounded-xl border border-stone-200/80 bg-[#f6f4f1] px-4 py-4 text-sm leading-relaxed text-stone-600">
            <p>Troškovi transporta nisu uračunati u cenu.</p>
            <p>
              Nakon prijema narudžbine kontaktiraćemo vas u najkraćem roku radi
              dogovora o načinu i ceni dostave.
            </p>
          </div>

          {formError ? (
            <p className="text-sm text-stone-700" role="alert">
              {formError}
            </p>
          ) : null}

          <button
            type="submit"
            disabled={submitting || hasBlockingItems || refreshing}
            className="inline-flex min-h-12 w-full items-center justify-center rounded-full bg-[#141311] px-6 text-sm font-medium text-white transition hover:bg-[#2a2420] disabled:cursor-not-allowed disabled:opacity-45 sm:w-auto sm:min-w-[220px]"
          >
            {submitting ? "Slanje…" : "Pošalji narudžbinu"}
          </button>
        </form>

        <CheckoutSummary items={items} subtotal={subtotal} />
      </div>
    </div>
  );
}
