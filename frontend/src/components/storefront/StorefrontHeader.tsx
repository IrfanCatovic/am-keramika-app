"use client";

import Image from "next/image";
import Link from "next/link";
import { FormEvent, useEffect, useId, useState } from "react";
import { usePathname, useRouter } from "next/navigation";

import { COMPANY_LOGO_SRC, companyConfig } from "@/config/company";
import type { PublicCategory } from "@/types/public-catalog";

export function StorefrontHeader({
  categories,
}: {
  categories: PublicCategory[];
}) {
  const pathname = usePathname();
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [catsOpen, setCatsOpen] = useState(false);
  const [search, setSearch] = useState("");
  const menuId = useId();

  function closeMenus() {
    setOpen(false);
    setCatsOpen(false);
  }

  useEffect(() => {
    if (!open) return;
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    document.addEventListener("keydown", onKey);
    document.body.style.overflow = "hidden";
    return () => {
      document.removeEventListener("keydown", onKey);
      document.body.style.overflow = "";
    };
  }, [open]);

  function onSearch(event: FormEvent) {
    event.preventDefault();
    const q = search.trim();
    router.push(q ? `/proizvodi?search=${encodeURIComponent(q)}` : "/proizvodi");
    closeMenus();
  }

  const linkClass = (href: string) =>
    `text-sm transition ${
      pathname === href || (href !== "/" && pathname.startsWith(href))
        ? "text-stone-900"
        : "text-stone-500 hover:text-stone-900"
    }`;

  return (
    <header className="sticky top-0 z-40 border-b border-stone-200/80 bg-[#f7f5f2]/90 backdrop-blur-md">
      <div className="mx-auto flex max-w-7xl items-center gap-4 px-4 py-3 sm:px-6 lg:px-8">
        <Link href="/" className="flex min-w-0 items-center gap-3">
          <Image
            src={COMPANY_LOGO_SRC}
            alt={companyConfig.name}
            width={40}
            height={40}
            className="h-10 w-10 object-contain"
            priority
          />
          <span className="truncate font-[family-name:var(--font-storefront-display)] text-xl tracking-wide text-stone-900 sm:text-2xl">
            {companyConfig.name}
          </span>
        </Link>

        <nav className="ml-6 hidden items-center gap-6 lg:flex">
          <Link href="/" className={linkClass("/")} onClick={closeMenus}>
            Početna
          </Link>
          <Link href="/proizvodi" className={linkClass("/proizvodi")} onClick={closeMenus}>
            Proizvodi
          </Link>
          <div
            className="relative"
            onMouseEnter={() => setCatsOpen(true)}
            onMouseLeave={() => setCatsOpen(false)}
          >
            <button
              type="button"
              className={`text-sm ${catsOpen || pathname.startsWith("/kategorije") ? "text-stone-900" : "text-stone-500 hover:text-stone-900"}`}
              aria-expanded={catsOpen}
            >
              Kategorije
            </button>
            {catsOpen && categories.length > 0 ? (
              <div className="absolute left-0 top-full z-50 min-w-[220px] pt-2">
                <div className="rounded-2xl border border-stone-200 bg-white p-2 shadow-lg">
                  {categories.map((category) => (
                    <Link
                      key={category.id}
                      href={`/kategorije/${category.slug}`}
                      onClick={closeMenus}
                      className="block rounded-xl px-3 py-2 text-sm text-stone-700 transition hover:bg-stone-50 hover:text-stone-900"
                    >
                      {category.name}
                    </Link>
                  ))}
                </div>
              </div>
            ) : null}
          </div>
        </nav>

        <form
          onSubmit={onSearch}
          className="ml-auto hidden min-w-0 flex-1 max-w-sm md:block"
        >
          <label className="sr-only" htmlFor="storefront-search">
            Pretraga proizvoda
          </label>
          <input
            id="storefront-search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Pretražite proizvode..."
            className="w-full rounded-full border border-stone-200 bg-white px-4 py-2 text-sm text-stone-900 outline-none ring-[#c4a484]/30 transition placeholder:text-stone-400 focus:ring-2"
          />
        </form>

        {/* Cart slot reserved for KORAK 3 — intentionally empty */}
        <div className="hidden w-10 shrink-0 lg:block" aria-hidden />

        <button
          type="button"
          className="ml-auto inline-flex h-10 w-10 items-center justify-center rounded-full border border-stone-200 bg-white text-stone-800 lg:hidden"
          aria-expanded={open}
          aria-controls={menuId}
          onClick={() => setOpen((v) => !v)}
        >
          <span className="sr-only">Meni</span>
          <span className="flex flex-col gap-1.5">
            <span className="block h-0.5 w-4 bg-current" />
            <span className="block h-0.5 w-4 bg-current" />
            <span className="block h-0.5 w-4 bg-current" />
          </span>
        </button>
      </div>

      <div className="border-t border-stone-200/60 px-4 py-2 md:hidden">
        <form onSubmit={onSearch}>
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Pretražite proizvode..."
            className="w-full rounded-full border border-stone-200 bg-white px-4 py-2.5 text-sm outline-none ring-[#c4a484]/30 focus:ring-2"
          />
        </form>
      </div>

      {open ? (
        <div
          id={menuId}
          className="fixed inset-0 z-50 lg:hidden"
          role="dialog"
          aria-modal="true"
        >
          <button
            type="button"
            className="absolute inset-0 bg-stone-900/40"
            aria-label="Zatvori meni"
            onClick={() => setOpen(false)}
          />
          <div className="absolute inset-y-0 right-0 flex w-[min(100%,20rem)] flex-col bg-[#f7f5f2] shadow-xl">
            <div className="flex items-center justify-between border-b border-stone-200 px-4 py-3">
              <span className="font-[family-name:var(--font-storefront-display)] text-lg text-stone-900">
                Meni
              </span>
              <button
                type="button"
                className="rounded-full border border-stone-200 bg-white px-3 py-1.5 text-sm"
                onClick={() => setOpen(false)}
              >
                Zatvori
              </button>
            </div>
            <nav className="flex flex-col gap-1 p-4">
              <Link href="/" onClick={closeMenus} className="rounded-xl px-3 py-3 text-base text-stone-800 hover:bg-white">
                Početna
              </Link>
              <Link
                href="/proizvodi"
                onClick={closeMenus}
                className="rounded-xl px-3 py-3 text-base text-stone-800 hover:bg-white"
              >
                Proizvodi
              </Link>
              <p className="mt-3 px-3 text-xs uppercase tracking-[0.16em] text-stone-400">
                Kategorije
              </p>
              {categories.map((category) => (
                <Link
                  key={category.id}
                  href={`/kategorije/${category.slug}`}
                  onClick={closeMenus}
                  className="rounded-xl px-3 py-2.5 text-sm text-stone-700 hover:bg-white"
                >
                  {category.name}
                </Link>
              ))}
            </nav>
          </div>
        </div>
      ) : null}
    </header>
  );
}
