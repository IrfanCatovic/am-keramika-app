"use client";

import { FormEvent, useState } from "react";

import { Modal } from "@/components/ui/Modal";

export function CategoryForm({
  open,
  mode,
  initialName = "",
  loading,
  error,
  onClose,
  onSubmit,
}: {
  open: boolean;
  mode: "create" | "edit";
  initialName?: string;
  loading: boolean;
  error: string | null;
  onClose: () => void;
  onSubmit: (name: string) => Promise<void> | void;
}) {
  const [name, setName] = useState(initialName);
  const [localError, setLocalError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) {
      setLocalError("Naziv kategorije je obavezan.");
      return;
    }
    setLocalError(null);
    await onSubmit(trimmed);
  }

  return (
    <Modal
      open={open}
      title={mode === "create" ? "Nova kategorija" : "Uredi kategoriju"}
      description="Unesite naziv kategorije proizvoda."
      onClose={loading ? () => undefined : onClose}
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label
            htmlFor="category-name"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            Naziv
          </label>
          <input
            id="category-name"
            value={name}
            onChange={(event) => setName(event.target.value)}
            disabled={loading}
            autoFocus
            className="w-full rounded-xl border border-stone-200 bg-white px-3 py-2.5 text-sm text-stone-900 outline-none ring-[#c4a484]/40 transition focus:ring-2 disabled:opacity-60"
            placeholder="npr. Keramika"
          />
        </div>

        {(localError || error) && (
          <p className="break-words rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
            {localError || error}
          </p>
        )}

        <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <button
            type="button"
            disabled={loading}
            onClick={onClose}
            className="inline-flex min-h-11 items-center justify-center rounded-xl border border-stone-200 px-4 text-sm font-medium text-stone-700 transition hover:bg-stone-50 disabled:opacity-60"
          >
            Otkaži
          </button>
          <button
            type="submit"
            disabled={loading}
            className="inline-flex min-h-11 items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white transition hover:bg-stone-800 disabled:opacity-60"
          >
            {loading ? "Čuvanje..." : mode === "create" ? "Dodaj" : "Sačuvaj"}
          </button>
        </div>
      </form>
    </Modal>
  );
}
