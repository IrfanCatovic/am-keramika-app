import Link from "next/link";

import { PublicProductGrid } from "@/components/storefront/PublicProductCard";
import type { PublicCategory, PublicProduct } from "@/types/public-catalog";

export function StorefrontHero() {
  return (
    <section className="relative overflow-hidden border-b border-stone-200 bg-[#111110] text-white">
      <div className="pointer-events-none absolute inset-0 marble-veil opacity-80" />
      <div className="pointer-events-none absolute -right-24 top-10 h-72 w-72 rounded-full bg-[#c4a484]/15 blur-3xl" />
      <div className="relative mx-auto flex min-h-[68vh] max-w-7xl flex-col justify-end px-4 pb-16 pt-24 sm:px-6 lg:px-8 lg:pb-20">
        <p className="text-xs uppercase tracking-[0.22em] text-[#c4a484]">
          AM Keramika
        </p>
        <h1 className="mt-4 max-w-3xl font-[family-name:var(--font-storefront-display)] text-4xl leading-[1.05] tracking-tight sm:text-5xl lg:text-6xl">
          Sve za vaš dom na jednom mjestu.
        </h1>
        <p className="mt-5 max-w-xl text-base text-stone-300 sm:text-lg">
          Keramika, sanitarije, grijanje i oprema.
        </p>
        <div className="mt-8 flex flex-wrap gap-3">
          <Link
            href="/proizvodi"
            className="inline-flex min-h-11 items-center rounded-full bg-white px-5 text-sm font-medium text-stone-900 transition hover:bg-stone-100"
          >
            Pogledajte proizvode
          </Link>
          <Link
            href="#kategorije"
            className="inline-flex min-h-11 items-center rounded-full border border-white/25 px-5 text-sm font-medium text-white transition hover:border-white/50 hover:bg-white/5"
          >
            Istražite kategorije
          </Link>
        </div>
      </div>
    </section>
  );
}

export function CategoryShowcase({
  categories,
}: {
  categories: PublicCategory[];
}) {
  if (categories.length === 0) return null;
  return (
    <section id="kategorije" className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
      <div className="mb-8 flex items-end justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.16em] text-stone-400">
            Asortiman
          </p>
          <h2 className="mt-2 font-[family-name:var(--font-storefront-display)] text-3xl text-stone-900">
            Kategorije
          </h2>
        </div>
        <Link
          href="/proizvodi"
          className="hidden text-sm text-stone-500 underline-offset-4 hover:text-stone-800 hover:underline sm:inline"
        >
          Svi proizvodi
        </Link>
      </div>
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
        {categories.map((category, index) => (
          <Link
            key={category.id}
            href={`/kategorije/${category.slug}`}
            className="group relative overflow-hidden rounded-2xl border border-stone-200 bg-white px-6 py-8 transition duration-300 hover:-translate-y-0.5 hover:border-stone-300 hover:shadow-[0_16px_40px_rgba(28,25,23,0.08)]"
            style={{ animationDelay: `${index * 40}ms` }}
          >
            <span className="text-xs uppercase tracking-[0.16em] text-[#8a6a45]">
              Kategorija
            </span>
            <span className="mt-3 block font-[family-name:var(--font-storefront-display)] text-2xl text-stone-900 transition group-hover:text-[#6b4f32]">
              {category.name}
            </span>
            <span className="mt-6 inline-flex text-sm text-stone-500 transition group-hover:text-stone-800">
              Pogledajte →
            </span>
          </Link>
        ))}
      </div>
    </section>
  );
}

export function ProductSection({
  title,
  eyebrow,
  products,
  href,
}: {
  title: string;
  eyebrow?: string;
  products: PublicProduct[];
  href?: string;
}) {
  if (products.length === 0) return null;
  return (
    <section className="mx-auto max-w-7xl px-4 py-14 sm:px-6 lg:px-8">
      <div className="mb-8 flex items-end justify-between gap-4">
        <div>
          {eyebrow ? (
            <p className="text-xs uppercase tracking-[0.16em] text-stone-400">
              {eyebrow}
            </p>
          ) : null}
          <h2 className="mt-2 font-[family-name:var(--font-storefront-display)] text-3xl text-stone-900">
            {title}
          </h2>
        </div>
        {href ? (
          <Link
            href={href}
            className="text-sm text-stone-500 underline-offset-4 hover:text-stone-800 hover:underline"
          >
            Pogledajte sve
          </Link>
        ) : null}
      </div>
      <PublicProductGrid products={products} />
    </section>
  );
}

export function StorefrontEmpty({
  title,
  description,
  actionHref,
  actionLabel,
}: {
  title: string;
  description: string;
  actionHref?: string;
  actionLabel?: string;
}) {
  return (
    <div className="rounded-3xl border border-dashed border-stone-300 bg-white/60 px-6 py-16 text-center">
      <h2 className="font-[family-name:var(--font-storefront-display)] text-2xl text-stone-900">
        {title}
      </h2>
      <p className="mx-auto mt-3 max-w-md text-sm text-stone-500">{description}</p>
      {actionHref && actionLabel ? (
        <Link
          href={actionHref}
          className="mt-6 inline-flex min-h-10 items-center rounded-full bg-stone-900 px-5 text-sm text-white"
        >
          {actionLabel}
        </Link>
      ) : null}
    </div>
  );
}

export function StorefrontBreadcrumb({
  items,
}: {
  items: { label: string; href?: string }[];
}) {
  return (
    <nav aria-label="Breadcrumb" className="mb-6 text-sm text-stone-500">
      <ol className="flex flex-wrap items-center gap-2">
        {items.map((item, index) => (
          <li key={`${item.label}-${index}`} className="inline-flex items-center gap-2">
            {index > 0 ? <span className="text-stone-300">/</span> : null}
            {item.href ? (
              <Link href={item.href} className="hover:text-stone-800">
                {item.label}
              </Link>
            ) : (
              <span className="text-stone-800">{item.label}</span>
            )}
          </li>
        ))}
      </ol>
    </nav>
  );
}
