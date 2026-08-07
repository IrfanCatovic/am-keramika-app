import Link from "next/link";

export default function StorefrontNotFound() {
  return (
    <div className="mx-auto flex min-h-[50vh] max-w-xl flex-col items-center justify-center px-4 py-20 text-center">
      <p className="text-xs uppercase tracking-[0.16em] text-stone-400">404</p>
      <h1 className="mt-3 font-[family-name:var(--font-storefront-display)] text-3xl text-stone-900">
        Stranica nije pronađena
      </h1>
      <p className="mt-3 text-sm text-stone-500">
        Traženi proizvod ili kategorija ne postoji ili više nije aktivna.
      </p>
      <div className="mt-8 flex flex-wrap justify-center gap-3">
        <Link
          href="/"
          className="inline-flex min-h-10 items-center rounded-full bg-stone-900 px-5 text-sm text-white"
        >
          Početna
        </Link>
        <Link
          href="/proizvodi"
          className="inline-flex min-h-10 items-center rounded-full border border-stone-200 bg-white px-5 text-sm text-stone-800"
        >
          Proizvodi
        </Link>
      </div>
    </div>
  );
}
