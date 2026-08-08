"use client";

import Image from "next/image";
import Link from "next/link";
import { FormEvent, useEffect, useId, useState } from "react";
import { createPortal } from "react-dom";
import { usePathname, useRouter } from "next/navigation";

import { STOREFRONT_LOGO_SRC, companyConfig } from "@/config/company";
import { CartButton } from "@/components/storefront/cart/CartButton";
import type { PublicCategory } from "@/types/public-catalog";

function MenuToggleIcon({ open }: { open: boolean }) {
  return (
    <span className="relative block h-3.5 w-5" aria-hidden>
      <span
        className={`absolute left-0 block h-[2px] w-5 rounded-full bg-[#d7b896] transition-all duration-300 ease-out ${
          open ? "top-[6px] rotate-45" : "top-0"
        }`}
      />
      <span
        className={`absolute left-0 top-[6px] block h-[2px] w-5 rounded-full bg-[#d7b896] transition-all duration-300 ease-out ${
          open ? "scale-x-0 opacity-0" : "scale-x-100 opacity-100"
        }`}
      />
      <span
        className={`absolute left-0 block h-[2px] w-5 rounded-full bg-[#d7b896] transition-all duration-300 ease-out ${
          open ? "top-[6px] -rotate-45" : "top-[12px]"
        }`}
      />
    </span>
  );
}

export function StorefrontHeader({
  categories,
}: {
  categories: PublicCategory[];
}) {
  const pathname = usePathname();
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [menuVisible, setMenuVisible] = useState(false);
  const [menuMounted, setMenuMounted] = useState(false);
  const [portalReady, setPortalReady] = useState(false);
  const [catsOpen, setCatsOpen] = useState(false);
  const [search, setSearch] = useState("");
  const menuId = useId();

  function closeMenus() {
    setOpen(false);
    setCatsOpen(false);
  }

  useEffect(() => {
    setPortalReady(true);
  }, []);

  useEffect(() => {
    if (open) {
      setMenuMounted(true);
      const frame = window.requestAnimationFrame(() => {
        window.requestAnimationFrame(() => setMenuVisible(true));
      });
      return () => window.cancelAnimationFrame(frame);
    }

    setMenuVisible(false);
    const timeout = window.setTimeout(() => setMenuMounted(false), 300);
    return () => window.clearTimeout(timeout);
  }, [open]);

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

  const mobileMenu =
    portalReady && menuMounted
      ? createPortal(
          <div
            id={menuId}
            className="fixed inset-0 z-[200] lg:hidden"
            role="dialog"
            aria-modal="true"
            aria-label="Navigacija"
          >
            <button
              type="button"
              className={`absolute inset-0 bg-[#1c1917]/75 transition-opacity duration-300 ease-out ${
                menuVisible ? "opacity-100" : "opacity-0"
              }`}
              aria-label="Zatvori meni"
              onClick={() => setOpen(false)}
            />
            <div
              className={`absolute inset-y-0 right-0 flex w-[min(100%,20rem)] flex-col bg-[#1c1917] text-[#f3e6d4] shadow-[-12px_0_40px_rgba(28,25,23,0.55)] transition-transform duration-300 ease-out ${
                menuVisible ? "translate-x-0" : "translate-x-full"
              }`}
            >
              <div className="flex items-center justify-between gap-3 border-b border-[#c4a484]/25 bg-[#14110e] px-4 py-3.5">
                <div className="rounded-lg bg-[#f6f4f1] px-2.5 py-1.5">
                  <Image
                    src={STOREFRONT_LOGO_SRC}
                    alt={companyConfig.name}
                    width={120}
                    height={40}
                    className="h-8 w-auto object-contain"
                  />
                </div>
                <button
                  type="button"
                  className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-[#c4a484]/40 bg-[#2a2420] text-[#e8d4b8] transition hover:border-[#c4a484]/65 hover:bg-[#3a3028] hover:text-white"
                  aria-label="Zatvori meni"
                  onClick={() => setOpen(false)}
                >
                  <MenuToggleIcon open />
                </button>
              </div>

              <nav className="flex flex-1 flex-col gap-1 overflow-y-auto p-4">
                <Link
                  href="/"
                  onClick={closeMenus}
                  className={`rounded-xl px-3 py-3 text-base font-medium tracking-wide transition ${
                    pathname === "/"
                      ? "bg-[#2a2420] text-[#f3e6d4] ring-1 ring-[#c4a484]/40"
                      : "text-[#f0e8dc] hover:bg-[#2a2420]/90 hover:text-white"
                  }`}
                >
                  Početna
                </Link>
                <Link
                  href="/proizvodi"
                  onClick={closeMenus}
                  className={`rounded-xl px-3 py-3 text-base font-medium tracking-wide transition ${
                    pathname.startsWith("/proizvodi")
                      ? "bg-[#2a2420] text-[#f3e6d4] ring-1 ring-[#c4a484]/40"
                      : "text-[#f0e8dc] hover:bg-[#2a2420]/90 hover:text-white"
                  }`}
                >
                  Proizvodi
                </Link>

                <p className="mt-5 px-3 text-[11px] font-semibold uppercase tracking-[0.18em] text-[#d7b896]">
                  Kategorije
                </p>
                {categories.map((category) => {
                  const href = `/kategorije/${category.slug}`;
                  const active = pathname === href;
                  return (
                    <Link
                      key={category.id}
                      href={href}
                      onClick={closeMenus}
                      className={`rounded-xl px-3 py-2.5 text-[15px] font-medium transition ${
                        active
                          ? "bg-[#2a2420] text-[#f3e6d4] ring-1 ring-[#c4a484]/35"
                          : "text-[#e8dfd2] hover:bg-[#2a2420]/90 hover:text-white"
                      }`}
                    >
                      {category.name}
                    </Link>
                  );
                })}

                <div className="mt-auto space-y-3 border-t border-[#c4a484]/25 pt-4">
                  <Link
                    href="/korpa"
                    onClick={closeMenus}
                    className="flex min-h-11 items-center justify-center rounded-full border border-[#c4a484]/35 bg-[#2a2420] text-sm font-medium text-[#f3e6d4] transition hover:border-[#c4a484]/60 hover:bg-[#3a3028]"
                  >
                    Korpa
                  </Link>
                  <Link
                    href="/login"
                    onClick={closeMenus}
                    className="flex min-h-11 items-center justify-center rounded-full bg-[#d7b896] text-sm font-semibold tracking-[0.06em] text-[#1c1917] transition hover:bg-[#e8d4b8]"
                  >
                    Login
                  </Link>
                </div>
              </nav>
            </div>
          </div>,
          document.body,
        )
      : null;

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
                      className="block rounded-lg px-3 py-2.5 text-sm text-stone-700 transition hover:bg-white hover:text-stone-900"
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
          className="hidden min-w-0 max-w-xs flex-1 md:block lg:max-w-sm"
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

        <div className="ml-auto flex items-center gap-2">
          <CartButton onBeforeOpen={() => setOpen(false)} />
          <Link
            href="/login"
            className="hidden items-center rounded-full border border-stone-800/80 px-4 py-2 text-[13px] font-medium tracking-[0.06em] text-stone-900 transition hover:bg-stone-900 hover:text-white lg:inline-flex"
          >
            Login
          </Link>
          <button
            type="button"
            className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-[#1c1917] ring-1 ring-[#c4a484]/45 transition hover:bg-[#2a2420] lg:hidden"
            aria-expanded={open}
            aria-controls={menuId}
            aria-label={open ? "Zatvori meni" : "Otvori meni"}
            onClick={() => setOpen((v) => !v)}
          >
            <MenuToggleIcon open={open} />
          </button>
        </div>
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

      {mobileMenu}
    </header>
  );
}
