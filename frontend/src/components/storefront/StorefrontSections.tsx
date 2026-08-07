import Link from "next/link";

import { PublicProductGrid } from "@/components/storefront/PublicProductCard";
import {
  STOREFRONT_HERO_SRC,
  STOREFRONT_SALON_SRC,
  companyAddressLines,
  companyConfig,
  companyContactLines,
} from "@/config/company";
import type { PublicCategory, PublicProduct } from "@/types/public-catalog";

export function StorefrontHero() {
  return (
    <section className="relative isolate min-h-[72vh] overflow-hidden text-white sm:min-h-[82vh]">
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={STOREFRONT_HERO_SRC}
        alt={`${companyConfig.name} — poslovnica`}
        className="absolute inset-0 h-full w-full object-cover object-[center_40%]"
        fetchPriority="high"
      />
      <div
        className="absolute inset-0 bg-gradient-to-t from-[#0c0b0a]/90 via-[#0c0b0a]/40 to-[#0c0b0a]/25"
        aria-hidden
      />
      <div
        className="absolute inset-0 bg-gradient-to-r from-[#0c0b0a]/55 via-transparent to-transparent"
        aria-hidden
      />
      <div
        className="pointer-events-none absolute inset-x-0 bottom-0 h-px bg-gradient-to-r from-transparent via-[#8a6a45]/40 to-transparent"
        aria-hidden
      />

      <div className="relative mx-auto flex min-h-[72vh] max-w-7xl flex-col justify-end px-4 pb-14 pt-28 sm:min-h-[82vh] sm:px-6 lg:px-8 lg:pb-20">
        <p className="text-[11px] uppercase tracking-[0.22em] text-[#d4b896]">
          {companyConfig.name}
        </p>
        <h1 className="mt-4 max-w-3xl font-[family-name:var(--font-storefront-display)] text-4xl leading-[1.05] tracking-tight sm:text-5xl lg:text-[3.5rem]">
          Sve za vaš dom na jednom mjestu.
        </h1>
        <p className="mt-5 max-w-xl text-base text-stone-200/90 sm:text-lg">
          Keramika, sanitarije, grijanje i oprema.
        </p>
        <div className="mt-8 flex flex-wrap gap-3">
          <Link
            href="/proizvodi"
            className="inline-flex min-h-11 items-center rounded-full bg-[#141311] px-6 text-sm font-medium text-white transition hover:bg-[#2a2420] hover:shadow-[0_0_0_1px_rgba(138,106,69,0.45)]"
          >
            Pogledajte proizvode
          </Link>
          <Link
            href="#kategorije"
            className="inline-flex min-h-11 items-center rounded-full border border-white/30 px-6 text-sm font-medium text-white transition hover:border-[#d4b896]/60 hover:bg-white/5"
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
    <section
      id="kategorije"
      className="border-b border-stone-200/80 bg-[#f6f4f1] px-4 py-16 sm:px-6 lg:px-8"
    >
      <div className="mx-auto max-w-7xl">
        <div className="mb-10 flex items-end justify-between gap-4">
          <div>
            <p className="text-[11px] uppercase tracking-[0.2em] text-[#8a6a45]">
              Asortiman
            </p>
            <h2 className="mt-2 font-[family-name:var(--font-storefront-display)] text-3xl text-stone-900 sm:text-4xl">
              Kategorije
            </h2>
          </div>
          <Link
            href="/proizvodi"
            className="hidden text-sm text-stone-500 transition hover:text-stone-800 sm:inline"
          >
            Svi proizvodi
          </Link>
        </div>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {categories.map((category) => (
            <Link
              key={category.id}
              href={`/kategorije/${category.slug}`}
              className="group relative overflow-hidden rounded-xl border border-stone-300/70 bg-white px-6 py-8 transition duration-300 hover:-translate-y-0.5 hover:border-stone-400 hover:shadow-[0_18px_40px_rgba(28,25,23,0.07)]"
            >
              <div
                className="pointer-events-none absolute inset-y-0 left-0 w-px bg-gradient-to-b from-transparent via-[#8a6a45]/50 to-transparent opacity-0 transition group-hover:opacity-100"
                aria-hidden
              />
              <p className="text-[11px] uppercase tracking-[0.2em] text-stone-400">
                Kategorija
              </p>
              <p className="mt-3 font-[family-name:var(--font-storefront-display)] text-2xl tracking-tight text-stone-900 transition group-hover:text-[#5c4630]">
                {category.name}
              </p>
              <span className="mt-8 inline-flex text-sm text-stone-500 transition group-hover:text-stone-800">
                Pogledajte
              </span>
            </Link>
          ))}
        </div>
      </div>
    </section>
  );
}

export function SalonSection() {
  const address = companyAddressLines();
  const contact = companyContactLines();

  return (
    <section className="border-y border-stone-200 bg-white">
      <div className="mx-auto grid max-w-7xl lg:grid-cols-[1.15fr_0.85fr]">
        <div className="relative min-h-[320px] overflow-hidden sm:min-h-[420px] lg:min-h-[520px]">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={STOREFRONT_SALON_SRC}
            alt={`${companyConfig.name} salon`}
            className="absolute inset-0 h-full w-full object-cover object-center"
            loading="lazy"
          />
          <div
            className="absolute inset-0 bg-[#141311]/10 mix-blend-multiply"
            aria-hidden
          />
        </div>
        <div className="flex flex-col justify-center px-6 py-12 sm:px-10 lg:px-12 lg:py-16">
          <p className="text-[11px] uppercase tracking-[0.2em] text-[#8a6a45]">
            {companyConfig.name}
          </p>
          <h2 className="mt-3 font-[family-name:var(--font-storefront-display)] text-3xl leading-tight text-stone-900 sm:text-4xl">
            Posjetite naš salon
          </h2>
          <div className="mt-4 h-px w-12 bg-gradient-to-r from-[#8a6a45] to-transparent" />
          <p className="mt-6 text-sm leading-relaxed text-stone-600 sm:text-base">
            Pogledajte našu ponudu i pronađite rješenja za vaš prostor.
          </p>
          <p className="mt-4 text-sm leading-relaxed text-stone-600 sm:text-base">
            Naš tim vam može pomoći pri izboru keramike, sanitarija, grijanja i
            ostale opreme.
          </p>
          {(address.length > 0 || contact.length > 0) && (
            <div className="mt-8 space-y-1 text-sm text-stone-500">
              {address.map((line) => (
                <p key={line}>{line}</p>
              ))}
              {contact.map((line) => (
                <p key={line}>{line}</p>
              ))}
            </div>
          )}
          <Link
            href="/proizvodi"
            className="mt-10 inline-flex min-h-11 w-fit items-center rounded-full bg-[#141311] px-6 text-sm font-medium text-white transition hover:bg-[#2a2420] hover:shadow-[0_0_0_1px_rgba(138,106,69,0.4)]"
          >
            Pregledajte proizvode
          </Link>
        </div>
      </div>
    </section>
  );
}

export function ProductSection({
  title,
  eyebrow,
  products,
  href,
  tone = "default",
}: {
  title: string;
  eyebrow?: string;
  products: PublicProduct[];
  href?: string;
  tone?: "default" | "muted" | "dark";
}) {
  if (products.length === 0) return null;

  const sectionClass =
    tone === "dark"
      ? "bg-[#141311] text-white"
      : tone === "muted"
        ? "bg-[#f6f4f1]"
        : "bg-white";

  const eyebrowClass =
    tone === "dark" ? "text-[#d4b896]" : "text-[#8a6a45]";
  const titleClass = tone === "dark" ? "text-white" : "text-stone-900";
  const linkClass =
    tone === "dark"
      ? "text-stone-400 hover:text-white"
      : "text-stone-500 hover:text-stone-800";

  return (
    <section className={`${sectionClass} px-4 py-16 sm:px-6 lg:px-8`}>
      <div className="mx-auto max-w-7xl">
        <div className="mb-10 flex items-end justify-between gap-4">
          <div>
            {eyebrow ? (
              <p className={`text-[11px] uppercase tracking-[0.2em] ${eyebrowClass}`}>
                {eyebrow}
              </p>
            ) : null}
            <h2
              className={`mt-2 font-[family-name:var(--font-storefront-display)] text-3xl sm:text-4xl ${titleClass}`}
            >
              {title}
            </h2>
          </div>
          {href ? (
            <Link
              href={href}
              className={`hidden text-sm sm:inline ${linkClass}`}
            >
              Pogledajte sve
            </Link>
          ) : null}
        </div>
        <PublicProductGrid products={products} />
      </div>
    </section>
  );
}

export function FinalCtaSection() {
  return (
    <section className="relative overflow-hidden bg-[#141311] px-4 py-16 text-center sm:px-6 lg:px-8">
      <div
        className="pointer-events-none absolute inset-0 marble-veil opacity-40"
        aria-hidden
      />
      <div className="relative mx-auto max-w-2xl">
        <p className="text-[11px] uppercase tracking-[0.2em] text-[#d4b896]">
          Katalog
        </p>
        <h2 className="mt-3 font-[family-name:var(--font-storefront-display)] text-3xl text-white sm:text-4xl">
          Pronađite rješenje za svoj prostor
        </h2>
        <p className="mt-4 text-sm leading-relaxed text-stone-400 sm:text-base">
          Pregledajte asortiman keramike, sanitarija, grijanja i opreme.
        </p>
        <Link
          href="/proizvodi"
          className="mt-8 inline-flex min-h-11 items-center rounded-full border border-[#d4b896]/50 px-6 text-sm font-medium text-white transition hover:border-[#d4b896] hover:bg-white/5"
        >
          Otvorite katalog
        </Link>
      </div>
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
    <div className="rounded-xl border border-dashed border-stone-300 bg-white/70 px-6 py-16 text-center">
      <h2 className="font-[family-name:var(--font-storefront-display)] text-2xl text-stone-900">
        {title}
      </h2>
      <p className="mx-auto mt-3 max-w-md text-sm text-stone-500">{description}</p>
      {actionHref && actionLabel ? (
        <Link
          href={actionHref}
          className="mt-6 inline-flex min-h-10 items-center rounded-full bg-[#141311] px-5 text-sm text-white"
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
          <li
            key={`${item.label}-${index}`}
            className="inline-flex items-center gap-2"
          >
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
