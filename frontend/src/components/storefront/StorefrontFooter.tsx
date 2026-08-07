import Link from "next/link";

import {
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
    <footer className="mt-auto border-t border-stone-200 bg-stone-950 text-stone-300">
      <div className="mx-auto grid max-w-7xl gap-10 px-4 py-12 sm:px-6 lg:grid-cols-[1.2fr_1fr_1fr] lg:px-8">
        <div>
          <p className="font-[family-name:var(--font-storefront-display)] text-2xl text-white">
            {companyConfig.name}
          </p>
          <p className="mt-3 max-w-sm text-sm leading-relaxed text-stone-400">
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
          <p className="text-xs uppercase tracking-[0.16em] text-stone-500">
            Navigacija
          </p>
          <ul className="mt-4 space-y-2 text-sm">
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
            {categories.slice(0, 6).map((category) => (
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
          <p className="text-xs uppercase tracking-[0.16em] text-stone-500">
            Informacije
          </p>
          <ul className="mt-4 space-y-2 text-sm">
            <li>
              <Link
                href="/login"
                className="text-stone-500 transition hover:text-stone-300"
              >
                Prijava za zaposlene
              </Link>
            </li>
          </ul>
        </div>
      </div>
      <div className="border-t border-white/10 px-4 py-4 text-center text-xs text-stone-500 sm:px-6 lg:px-8">
        © {year} {companyConfig.name}
      </div>
    </footer>
  );
}
