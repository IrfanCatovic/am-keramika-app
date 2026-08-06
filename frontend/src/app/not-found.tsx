import Link from "next/link";

export default function NotFound() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-100 px-4">
      <div className="w-full max-w-md rounded-xl border border-slate-200 bg-white p-8 text-center shadow-sm">
        <p className="text-sm font-medium uppercase tracking-[0.16em] text-slate-500">
          404
        </p>
        <h1 className="mt-2 text-2xl font-semibold text-slate-900">
          Stranica nije pronađena
        </h1>
        <p className="mt-2 text-sm text-slate-500">
          Tražena ruta ne postoji u AM Keramika aplikaciji.
        </p>
        <Link
          href="/dashboard"
          className="mt-6 inline-flex rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800"
        >
          Nazad na dashboard
        </Link>
      </div>
    </div>
  );
}
