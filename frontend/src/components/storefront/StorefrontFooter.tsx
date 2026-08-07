import Image from "next/image";
import Link from "next/link";

import {
  STOREFRONT_LOGO_SRC,
  companyAddressLines,
  companyConfig,
  companyContactLines,
} from "@/config/company";
import type { PublicCategory } from "@/types/public-catalog";

export function StorefrontFooter({
  categories,
}: {
  categories: PublicCategory[];
}) {
  const address = companyAddressLines();
  const contact = companyContactLines();
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
            className="h-12 w-auto object-contain brightness-110 contrast-105"
          />
          <p className="mt-5 max-w-sm text-sm leading-relaxed text-stone-400">
            Keramika, sanitarije, grijanje i oprema za vaš dom.
          </p>
          {(address.length > 0 || contact.length > 0) && (
            <div className="mt-5 space-y-1 text-sm text-stone-400">
              {address.map((line) => (
                <p key={line}>{line}</p>
              ))}
              {contact.map((line) => (
                <p key={line}>{line}</p>
              ))}
            </div>
          )}
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
        © {year} {companyConfig.name}
      </div>
    </footer>
  );
}
