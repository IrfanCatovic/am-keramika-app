import Link from "next/link";

import { PublicProductGrid } from "@/components/storefront/PublicProductCard";
import {
  STOREFRONT_HERO_SRC,
  STOREFRONT_SALON_SRC,
  companyAddressLines,
  companyConfig,
  companyContactLines,
} from "@/config/company";
import type {
  PublicCategory,
  PublicProduct,
  PublicProductImage,
} from "@/types/public-catalog";

export type CategoryShowcaseItem = PublicCategory & {
  image: PublicProductImage | null;
};

export function StorefrontHero() {
  return (
    <section className="relative isolate min-h-[72vh] overflow-hidden text-white sm:min-h-[82vh]">
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={STOREFRONT_HERO_SRC}
        alt={`${companyConfig.name} — poslovnica`}
        className="absolute inset-0 h-full w-full object-cover object-[center_34%] sm:object-[center_40%]"
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
        <p className="text-xs font-semibold uppercase tracking-[0.28em] text-[#e8d4b8] sm:text-sm">
          {companyConfig.name}
        </p>
        <h1 className="mt-4 max-w-3xl font-[family-name:var(--font-storefront-display)] text-4xl leading-[1.05] tracking-tight sm:text-5xl lg:text-[3.5rem]">
          Sve za vaš dom na jednom mestu.
        </h1>
        <p className="mt-5 max-w-xl text-base text-stone-200/90 sm:text-lg">
          Keramika, sanitarije, grejanje i oprema.
        </p>
        <div className="mt-8 flex flex-col gap-3 sm:flex-row sm:flex-wrap">
          <Link
            href="/proizvodi"
            className="inline-flex min-h-12 items-center justify-center rounded-full bg-white px-7 text-sm font-medium text-[#141311] transition hover:bg-[#ece5db] sm:min-h-11"
          >
            Pogledajte proizvode
          </Link>
          <Link
            href="#salon"
            className="inline-flex min-h-12 items-center justify-center rounded-full border border-white/30 px-7 text-sm font-medium text-white transition hover:border-[#d4b896]/70 hover:bg-white/5 sm:min-h-11"
          >
            Posetite salon
          </Link>
        </div>
      </div>
    </section>
  );
}

export function CategoryShowcase({
  categories,
}: {
  categories: CategoryShowcaseItem[];
}) {
  if (categories.length === 0) return null;
  return (
    <section
      id="kategorije"
      className="scroll-mt-24 border-b border-stone-200/80 bg-[#f6f4f1] px-4 py-14 sm:px-6 sm:py-16 lg:px-8"
    >
      <div className="mx-auto max-w-7xl">
        <div className="mb-8 flex items-end justify-between gap-4 sm:mb-10">
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
        <div className="grid grid-cols-1 gap-3 min-[420px]:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {categories.map((category, index) => {
            const imageUrl = category.image?.url;

            return (
              <Link
                key={category.id}
                href={`/kategorije/${category.slug}`}
                prefetch={false}
                className={`group relative isolate min-h-52 overflow-hidden rounded-xl border border-stone-300/70 transition duration-300 hover:-translate-y-0.5 hover:border-stone-400 hover:shadow-[0_18px_40px_rgba(28,25,23,0.1)] ${
                  imageUrl ? "bg-stone-900 text-white" : "bg-white"
                }`}
              >
                {imageUrl ? (
                  <>
                    {/* API image URLs are runtime-configured, so Next/Image cannot
                        safely optimize them without broadening remotePatterns. */}
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img
                      src={imageUrl}
                      alt=""
                      className="absolute inset-0 h-full w-full object-cover transition duration-500 group-hover:scale-[1.03]"
                      loading="lazy"
                    />
                    <div
                      className="absolute inset-0 bg-gradient-to-t from-[#141311]/90 via-[#141311]/25 to-transparent"
                      aria-hidden
                    />
                  </>
                ) : null}
                <div
                  className={`relative flex min-h-52 flex-col justify-end p-5 ${
                    imageUrl ? "text-white" : "text-stone-900"
                  }`}
                >
                  <p
                    className={`font-[family-name:var(--font-storefront-display)] text-sm tabular-nums tracking-[0.18em] ${
                      imageUrl ? "text-[#e8d4b8]" : "text-[#b39a7c]"
                    }`}
                  >
                    {String(index + 1).padStart(2, "0")}
                  </p>
                  <p className="mt-2 max-w-[18rem] font-[family-name:var(--font-storefront-display)] text-2xl leading-tight tracking-tight">
                    {category.name}
                  </p>
                  <span
                    className={`mt-4 inline-flex text-sm transition ${
                      imageUrl
                        ? "text-stone-200 group-hover:text-white"
                        : "text-stone-500 group-hover:text-stone-800"
                    }`}
                  >
                    Pogledajte
                  </span>
                </div>
              </Link>
            );
          })}
        </div>
        <div className="mt-6 text-center sm:hidden">
          <Link
            href="/proizvodi"
            className="text-sm text-stone-500 underline-offset-4 hover:text-stone-800 hover:underline"
          >
            Svi proizvodi
          </Link>
        </div>
      </div>
    </section>
  );
}

export function SalonSection() {
  const address = companyAddressLines();
  const contact = companyContactLines();

  return (
    <section id="salon" className="relative scroll-mt-24 overflow-hidden bg-[#141311] text-white">
      <div
        className="pointer-events-none absolute inset-0 marble-veil opacity-50"
        aria-hidden
      />
      <div
        className="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-[#8a6a45]/50 to-transparent"
        aria-hidden
      />
      <div className="relative mx-auto grid max-w-7xl lg:grid-cols-[1.15fr_0.85fr]">
        <div className="relative min-h-[280px] overflow-hidden sm:min-h-[420px] lg:min-h-[540px]">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={STOREFRONT_SALON_SRC}
            alt={`${companyConfig.name} salon`}
            className="absolute inset-0 h-full w-full object-cover object-center"
            loading="lazy"
          />
          <div
            className="absolute inset-0 bg-gradient-to-r from-transparent via-transparent to-[#141311]/60 max-lg:bg-gradient-to-t max-lg:to-[#141311]/70"
            aria-hidden
          />
        </div>
        <div className="flex flex-col justify-center px-6 py-12 sm:px-10 lg:px-12 lg:py-16">
          <p className="text-[11px] uppercase tracking-[0.2em] text-[#d4b896]">
            {companyConfig.name}
          </p>
          <h2 className="mt-3 font-[family-name:var(--font-storefront-display)] text-3xl leading-tight text-white sm:text-4xl">
            Posetite naš salon
          </h2>
          <div className="mt-4 h-px w-12 bg-gradient-to-r from-[#d4b896] to-transparent" />
          <p className="mt-6 text-sm leading-relaxed text-stone-300 sm:text-base">
            Pogledajte našu ponudu i pronađite rešenja za vaš prostor.
          </p>
          <p className="mt-4 text-sm leading-relaxed text-stone-300 sm:text-base">
            Naš tim vam može pomoći pri izboru keramike, sanitarija, grejanja i
            ostale opreme.
          </p>
          {(address.length > 0 || contact.length > 0) && (
            <div className="mt-8 space-y-1 text-sm text-stone-400">
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
            className="mt-10 inline-flex min-h-12 w-full items-center justify-center rounded-full bg-white px-6 text-sm font-medium text-[#141311] transition hover:bg-[#ece5db] sm:min-h-11 sm:w-fit"
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
  homepage = false,
}: {
  title: string;
  eyebrow?: string;
  products: PublicProduct[];
  href?: string;
  tone?: "default" | "muted" | "dark";
  homepage?: boolean;
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
    <section className={`${sectionClass} px-4 py-14 sm:px-6 sm:py-16 lg:px-8`}>
      <div className="mx-auto max-w-7xl">
        <div className="mb-8 flex items-end justify-between gap-4 sm:mb-10">
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
        <PublicProductGrid products={products} homepage={homepage} />
        {href ? (
          <div className="mt-6 text-center sm:hidden">
            <Link
              href={href}
              className={`text-sm underline-offset-4 hover:underline ${linkClass}`}
            >
              Pogledajte sve
            </Link>
          </div>
        ) : null}
      </div>
    </section>
  );
}

export function FinalCtaSection() {
  return (
    <section className="relative overflow-hidden bg-[#141311] px-4 py-16 text-center sm:px-6 sm:py-20 lg:px-8">
      <div
        className="pointer-events-none absolute inset-0 marble-veil opacity-40"
        aria-hidden
      />
      <div
        className="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-[#8a6a45]/40 to-transparent"
        aria-hidden
      />
      <div className="relative mx-auto max-w-2xl">
        <p className="text-[11px] uppercase tracking-[0.2em] text-[#d4b896]">
          Katalog
        </p>
        <h2 className="mt-3 font-[family-name:var(--font-storefront-display)] text-3xl text-white sm:text-4xl">
          Pronađite rešenje za svoj prostor
        </h2>
        <p className="mt-4 text-sm leading-relaxed text-stone-400 sm:text-base">
          Pregledajte asortiman keramike, sanitarija, grejanja i opreme.
        </p>
        <Link
          href="/proizvodi"
          className="mt-8 inline-flex min-h-12 w-full max-w-xs items-center justify-center rounded-full border border-[#d4b896]/50 px-6 text-sm font-medium text-white transition hover:border-[#d4b896] hover:bg-white/5 sm:min-h-11 sm:w-auto sm:max-w-none"
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
