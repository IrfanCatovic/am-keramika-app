import Link from "next/link";

export function CartEmptyState() {
  return (
    <div className="mx-auto max-w-xl px-4 py-20 text-center sm:px-6">
      <p className="text-[11px] uppercase tracking-[0.18em] text-[#8a6a45]">
        Korpa
      </p>
      <h1 className="mt-3 font-[family-name:var(--font-storefront-display)] text-3xl text-stone-900 sm:text-4xl">
        Vaša korpa je prazna.
      </h1>
      <p className="mt-4 text-sm leading-relaxed text-stone-500 sm:text-base">
        Istražite našu ponudu i pronađite proizvode za vaš prostor.
      </p>
      <Link
        href="/proizvodi"
        className="mt-8 inline-flex min-h-11 items-center justify-center rounded-full bg-[#141311] px-6 text-sm font-medium text-white transition hover:bg-[#2a2420]"
      >
        Pogledajte proizvode
      </Link>
    </div>
  );
}
