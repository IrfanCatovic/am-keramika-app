"use client";

import { StatusBadge } from "@/components/ui/EmptyState";
import { Category } from "@/types/category";

export function CategoryCard({
  category,
  selected,
  busy,
  onSelect,
  onEdit,
  onToggleStatus,
  onDelete,
}: {
  category: Category;
  selected: boolean;
  busy: boolean;
  onSelect: () => void;
  onEdit: () => void;
  onToggleStatus: () => void;
  onDelete: () => void;
}) {
  return (
    <article
      className={`dash-enter min-w-0 rounded-2xl border p-3.5 transition sm:p-4 ${
        selected
          ? "border-[#c4a484]/70 bg-[#faf6f1] shadow-[0_8px_24px_rgba(28,25,23,0.06)] ring-1 ring-[#c4a484]/35"
          : "border-stone-200 bg-white hover:border-stone-300"
      } ${category.isActive ? "" : "opacity-70"}`}
    >
      <button
        type="button"
        onClick={onSelect}
        className="w-full min-w-0 text-left"
        aria-pressed={selected}
      >
        <div className="flex flex-wrap items-start justify-between gap-2">
          <div className="min-w-0">
            <p className="break-words text-sm font-semibold text-stone-900 sm:text-base">
              {category.name}
            </p>
            <p className="mt-1 break-all text-xs text-stone-500">
              {category.slug}
            </p>
          </div>
          <StatusBadge active={category.isActive} />
        </div>
      </button>

      <div className="mt-3 flex flex-wrap gap-2">
        <button
          type="button"
          disabled={busy}
          onClick={onEdit}
          className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 bg-white px-3 text-sm font-medium text-stone-700 transition hover:bg-stone-50 disabled:opacity-60"
        >
          Uredi
        </button>
        <button
          type="button"
          disabled={busy}
          onClick={onToggleStatus}
          className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 bg-white px-3 text-sm font-medium text-stone-700 transition hover:bg-stone-50 disabled:opacity-60"
        >
          {category.isActive ? "Deaktiviraj" : "Aktiviraj"}
        </button>
        <button
          type="button"
          disabled={busy}
          onClick={onDelete}
          className="inline-flex min-h-10 items-center rounded-xl border border-red-200 bg-white px-3 text-sm font-medium text-red-700 transition hover:bg-red-50 disabled:opacity-60"
        >
          Obriši
        </button>
      </div>
    </article>
  );
}
