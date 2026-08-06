"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";

export function CustomerForm({
  mode,
  initialName = "",
  initialPhone = "",
  loading,
  error,
  onSubmit,
  cancelHref,
}: {
  mode: "create" | "edit";
  initialName?: string;
  initialPhone?: string;
  loading: boolean;
  error: string | null;
  onSubmit: (values: { name: string; phone: string }) => Promise<void> | void;
  cancelHref: string;
}) {
  const [name, setName] = useState(initialName);
  const [phone, setPhone] = useState(initialPhone);
  const [localError, setLocalError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (loading) {
      return;
    }
    const trimmedName = name.trim();
    if (!trimmedName) {
      setLocalError("Naziv kupca je obavezan.");
      return;
    }
    setLocalError(null);
    await onSubmit({ name: trimmedName, phone: phone.trim() });
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="dash-enter space-y-4 rounded-2xl border border-stone-200 bg-white p-4 shadow-[0_1px_2px_rgba(28,25,23,0.04)] sm:p-5"
    >
      <div>
        <label
          htmlFor="customer-name"
          className="mb-1.5 block text-sm font-medium text-stone-700"
        >
          Ime / naziv *
        </label>
        <input
          id="customer-name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          disabled={loading}
          className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
          placeholder="npr. Amir Softić ili Keramika d.o.o."
        />
      </div>

      <div>
        <label
          htmlFor="customer-phone"
          className="mb-1.5 block text-sm font-medium text-stone-700"
        >
          Telefon
        </label>
        <input
          id="customer-phone"
          value={phone}
          onChange={(event) => setPhone(event.target.value)}
          disabled={loading}
          className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
          placeholder="opciono"
        />
        <p className="mt-1 text-xs text-stone-500">
          Telefon nije unique — više kupaca može imati isti broj.
        </p>
      </div>

      {(localError || error) && (
        <p className="break-words rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
          {localError || error}
        </p>
      )}

      <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <Link
          href={cancelHref}
          className="inline-flex min-h-11 items-center justify-center rounded-xl border border-stone-200 px-4 text-sm font-medium text-stone-700 hover:bg-stone-50"
        >
          Otkaži
        </Link>
        <button
          type="submit"
          disabled={loading}
          className="inline-flex min-h-11 items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white hover:bg-stone-800 disabled:opacity-60"
        >
          {loading
            ? "Čuvanje..."
            : mode === "create"
              ? "Sačuvaj kupca"
              : "Sačuvaj izmjene"}
        </button>
      </div>
    </form>
  );
}
