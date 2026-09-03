import Image from "next/image";
import Link from "next/link";

import {
  STOREFRONT_LOGO_SRC,
  companyAddressLines,
  companyConfig,
  companyContactLines,
  companyIdLines,
} from "@/config/company";
import type { PublicCategory } from "@/types/public-catalog";

export function StorefrontFooter({
  categories,
}: {
  categories: PublicCategory[];
}) {
  const address = companyAddressLines();
  const contact = companyContactLines();
  const legal = companyIdLines();
  const year = new Date().getFullYear();

  return (
    <footer className="relative mt-auto overflow-hidden bg-[#121110] text-stone-300">
      <div
        className="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-[#8a6a45]/50 to-transparent"
        aria-hidden
      />
      <div
        className="pointer-events-none absolute -left-20 bottom-0 h-48 w-48 rounded-full bg-[#8a6a45]/10 blur-3xl"
        aria-hidden
      />

      <div className="relative mx-auto grid max-w-7xl gap-10 px-4 py-14 sm:px-6 lg:grid-cols-[1.3fr_1fr_1fr] lg:px-8">
        <div>
          <Image
            src={STOREFRONT_LOGO_SRC}
            alt={companyConfig.name}
            width={160}
            height={52}
            className="h-11 w-auto object-contain sm:h-12"
          />
          <p className="mt-5 max-w-sm text-sm leading-relaxed text-stone-400">
            Keramika, sanitarije, grejanje i oprema za vaš dom.
          </p>
          <div className="mt-5 flex items-center gap-3">
            <a
              href="https://www.facebook.com/profile.php?id=100068140516753"
              target="_blank"
              rel="noreferrer"
              aria-label="AM Keramika na Facebooku"
              className="flex h-10 w-10 items-center justify-center rounded-full border border-white/15 text-stone-400 transition duration-200 hover:-translate-y-0.5 hover:border-[#8a6a45] hover:text-white"
            >
              <svg
                viewBox="0 0 24 24"
                className="h-5 w-5 fill-current"
                aria-hidden="true"
              >
                <path d="M13.5 21v-8h2.75l.5-3h-3.25V8.05c0-.87.29-1.55 1.58-1.55H17V3.82c-.34-.05-1.2-.12-2.28-.12-2.26 0-3.8 1.38-3.8 3.92V10H8.5v3h2.42v8h2.58Z" />
              </svg>
            </a>
            <a
              href="https://www.instagram.com/amkeramika/"
              target="_blank"
              rel="noreferrer"
              aria-label="AM Keramika na Instagramu"
              className="flex h-10 w-10 items-center justify-center rounded-full border border-white/15 text-stone-400 transition duration-200 hover:-translate-y-0.5 hover:border-[#8a6a45] hover:text-white"
            >
              <svg
                viewBox="0 0 24 24"
                className="h-5 w-5 fill-none stroke-current"
                strokeWidth="1.8"
                aria-hidden="true"
              >
                <rect x="3.5" y="3.5" width="17" height="17" rx="4.5" />
                <circle cx="12" cy="12" r="4" />
                <circle cx="17.5" cy="6.5" r="1" className="fill-current stroke-none" />
              </svg>
            </a>
          </div>
          <div className="mt-5 space-y-1 text-sm text-stone-400">
            <p className="font-medium text-stone-300">{companyConfig.name}</p>
            <div className="mt-2 space-y-1">
              {address.length > 0 ? <p>{address.join(", ")}</p> : null}
              {contact.map((line) => (
                <p key={line}>{line}</p>
              ))}
            </div>
            {legal.length > 0 && (
              <div className="mt-3 space-y-1 text-xs text-stone-500">
                {legal.map((line) => (
                  <p key={line}>{line}</p>
                ))}
              </div>
            )}
            <div className="mt-3 space-y-1 text-xs text-stone-500">
              <p>Tekući račun: 160-6000002347524-65 — Banca Intesa</p>
              <p>Tekući račun: 155-0000000082232-82 — Halkbank</p>
            </div>
          </div>
        </div>

        <div>
          <p className="text-[11px] uppercase tracking-[0.18em] text-stone-500">
            Navigacija
          </p>
          <ul className="mt-4 space-y-2.5 text-sm">
            <li>
              <Link href="/" className="transition hover:text-white">
                Početna
              </Link>
            </li>
            <li>
              <Link href="/proizvodi" className="transition hover:text-white">
                Proizvodi
              </Link>
            </li>
            <li>
              <Link href="/#kategorije" className="transition hover:text-white">
                Kategorije
              </Link>
            </li>
            {categories.slice(0, 5).map((category) => (
              <li key={category.id}>
                <Link
                  href={`/kategorije/${category.slug}`}
                  prefetch={false}
                  className="transition hover:text-white"
                >
                  {category.name}
                </Link>
              </li>
            ))}
          </ul>
        </div>

        <div>
          <p className="text-[11px] uppercase tracking-[0.18em] text-stone-500">
            Nalog
          </p>
          <ul className="mt-4 space-y-2.5 text-sm">
            <li>
              <Link href="/login" className="transition hover:text-white">
                Login
              </Link>
            </li>
          </ul>
        </div>
      </div>

      <div className="relative border-t border-white/10 px-4 py-4 text-center text-xs text-stone-500 sm:px-6 lg:px-8">
        <p>© {year} {companyConfig.name}</p>
        <p className="mt-1 text-[10px] text-stone-600">
          Developer: Irfan Ćatović
        </p>
      </div>
    </footer>
  );
}
