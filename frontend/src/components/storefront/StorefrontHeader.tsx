"use client";

import Image from "next/image";
import Link from "next/link";
import { FormEvent, useEffect, useId, useState } from "react";
import { usePathname, useRouter } from "next/navigation";

import { STOREFRONT_LOGO_SRC, companyConfig } from "@/config/company";
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
    `text-[13px] tracking-[0.04em] transition ${
      pathname === href || (href !== "/" && pathname.startsWith(href))
        ? "text-stone-900"
        : "text-stone-500 hover:text-stone-900"
    }`;

  return (
    <header className="sticky top-0 z-40 border-b border-stone-200/70 bg-[#f6f4f1]/92 backdrop-blur-md">
      <div className="mx-auto flex max-w-7xl items-center gap-3 px-4 py-3 sm:gap-4 sm:px-6 lg:px-8">
        <Link
          href="/"
          className="flex shrink-0 items-center"
          aria-label={companyConfig.name}
        >
          <Image
            src={STOREFRONT_LOGO_SRC}
            alt={companyConfig.name}
            width={148}
            height={48}
            className="h-10 w-auto object-contain sm:h-11"
            priority
          />
        </Link>

        <nav className="ml-2 hidden items-center gap-7 lg:flex">
          <Link href="/" className={linkClass("/")} onClick={closeMenus}>
            Početna
          </Link>
          <Link
            href="/proizvodi"
            className={linkClass("/proizvodi")}
            onClick={closeMenus}
          >
            Proizvodi
          </Link>
          <div
            className="relative"
            onMouseEnter={() => setCatsOpen(true)}
            onMouseLeave={() => setCatsOpen(false)}
          >
            <button
              type="button"
              className={`text-[13px] tracking-[0.04em] ${
                catsOpen || pathname.startsWith("/kategorije")
                  ? "text-stone-900"
                  : "text-stone-500 hover:text-stone-900"
              }`}
              aria-expanded={catsOpen}
            >
              Kategorije
            </button>
            {catsOpen && categories.length > 0 ? (
              <div className="absolute left-0 top-full z-50 min-w-[230px] pt-3">
                <div className="overflow-hidden rounded-xl border border-stone-200 bg-white p-1.5 shadow-[0_16px_40px_rgba(28,25,23,0.1)]">
                  {categories.map((category) => (
                    <Link
                      key={category.id}
                      href={`/kategorije/${category.slug}`}
                      onClick={closeMenus}
                      className="block rounded-lg px-3 py-2.5 text-sm text-stone-700 transition hover:bg-stone-50 hover:text-stone-900"
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
          className="ml-auto hidden min-w-0 max-w-xs flex-1 md:block lg:max-w-sm"
        >
          <label className="sr-only" htmlFor="storefront-search">
            Pretraga proizvoda
          </label>
          <input
            id="storefront-search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Pretražite proizvode..."
            className="w-full rounded-full border border-stone-300/80 bg-white/90 px-4 py-2 text-sm text-stone-900 outline-none transition placeholder:text-stone-400 focus:border-stone-400 focus:ring-2 focus:ring-[rgba(138,106,69,0.18)]"
          />
        </form>

        <Link
          href="/login"
          className="ml-1 hidden items-center rounded-full border border-stone-800/80 px-4 py-2 text-[13px] font-medium tracking-[0.06em] text-stone-900 transition hover:bg-stone-900 hover:text-white lg:inline-flex"
        >
          Login
        </Link>

        <button
          type="button"
          className="ml-auto inline-flex h-10 w-10 items-center justify-center rounded-full border border-stone-300 bg-white text-stone-800 lg:hidden"
          aria-expanded={open}
          aria-controls={menuId}
          onClick={() => setOpen((v) => !v)}
        >
          <span className="sr-only">Meni</span>
          <span className="flex flex-col gap-1.5">
            <span className="block h-px w-4 bg-current" />
            <span className="block h-px w-4 bg-current" />
            <span className="block h-px w-4 bg-current" />
          </span>
        </button>
      </div>

      <div className="border-t border-stone-200/50 px-4 py-2 md:hidden">
        <form onSubmit={onSearch}>
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Pretražite proizvode..."
            className="w-full rounded-full border border-stone-300/80 bg-white px-4 py-2.5 text-sm outline-none focus:ring-2 focus:ring-[rgba(138,106,69,0.18)]"
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
            className="absolute inset-0 bg-stone-950/45"
            aria-label="Zatvori meni"
            onClick={() => setOpen(false)}
          />
          <div className="absolute inset-y-0 right-0 flex w-[min(100%,20rem)] flex-col border-l border-stone-200 bg-[#f6f4f1] shadow-2xl">
            <div className="flex items-center justify-between border-b border-stone-200 px-4 py-3">
              <Image
                src={STOREFRONT_LOGO_SRC}
                alt={companyConfig.name}
                width={120}
                height={40}
                className="h-9 w-auto object-contain"
              />
              <button
                type="button"
                className="rounded-full border border-stone-300 bg-white px-3 py-1.5 text-sm"
                onClick={() => setOpen(false)}
              >
                Zatvori
              </button>
            </div>
            <nav className="flex flex-1 flex-col gap-1 p-4">
              <Link
                href="/"
                onClick={closeMenus}
                className="rounded-lg px-3 py-3 text-base text-stone-800 hover:bg-white"
              >
                Početna
              </Link>
              <Link
                href="/proizvodi"
                onClick={closeMenus}
                className="rounded-lg px-3 py-3 text-base text-stone-800 hover:bg-white"
              >
                Proizvodi
              </Link>
              <p className="mt-4 px-3 text-[11px] uppercase tracking-[0.18em] text-stone-400">
                Kategorije
              </p>
              {categories.map((category) => (
                <Link
                  key={category.id}
                  href={`/kategorije/${category.slug}`}
                  onClick={closeMenus}
                  className="rounded-lg px-3 py-2.5 text-sm text-stone-700 hover:bg-white"
                >
                  {category.name}
                </Link>
              ))}
              <div className="mt-auto border-t border-stone-200 pt-4">
                <Link
                  href="/login"
                  onClick={closeMenus}
                  className="flex min-h-11 items-center justify-center rounded-full border border-stone-800 text-sm font-medium tracking-[0.06em] text-stone-900 transition hover:bg-stone-900 hover:text-white"
                >
                  Login
                </Link>
              </div>
            </nav>
          </div>
        </div>
      ) : null}
    </header>
  );
}
