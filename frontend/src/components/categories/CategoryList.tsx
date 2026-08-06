"use client";

import { CategoryCard } from "@/components/categories/CategoryCard";
import {
  EmptyState,
  InlineError,
  ListSkeleton,
} from "@/components/ui/EmptyState";
import { Category } from "@/types/category";

export function CategoryList({
  categories,
  selectedId,
  loading,
  error,
  busyId,
  onRetry,
  onSelect,
  onCreate,
  onEdit,
  onToggleStatus,
  onDelete,
}: {
  categories: Category[];
  selectedId: number | null;
  loading: boolean;
  error: string | null;
  busyId: number | null;
  onRetry: () => void;
  onSelect: (category: Category) => void;
  onCreate: () => void;
  onEdit: (category: Category) => void;
  onToggleStatus: (category: Category) => void;
  onDelete: (category: Category) => void;
}) {
  return (
    <section className="min-w-0 rounded-2xl border border-stone-200/90 bg-white shadow-[0_1px_2px_rgba(28,25,23,0.04)]">
      <div className="flex flex-wrap items-start justify-between gap-3 border-b border-stone-100 px-4 py-3.5 sm:px-5 sm:py-4">
        <div className="min-w-0">
          <h2 className="text-base font-semibold tracking-tight text-stone-900">
            Kategorije
          </h2>
          <p className="mt-0.5 text-sm text-stone-500">
            Izaberite kategoriju za pregled grupa
          </p>
        </div>
        <button
          type="button"
          onClick={onCreate}
          className="inline-flex min-h-11 shrink-0 items-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white transition hover:bg-stone-800"
        >
          Nova kategorija
        </button>
      </div>

      <div className="space-y-3 px-4 py-4 sm:px-5">
        {loading ? <ListSkeleton rows={4} /> : null}

        {!loading && error ? (
          <InlineError message={error} onRetry={onRetry} />
        ) : null}

        {!loading && !error && categories.length === 0 ? (
          <EmptyState
            title="Nema kategorija"
            description="Dodajte prvu kategoriju da biste organizovali proizvode."
            action={
              <button
                type="button"
                onClick={onCreate}
                className="inline-flex min-h-11 items-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white"
              >
                Dodaj kategoriju
              </button>
            }
          />
        ) : null}

        {!loading && !error
          ? categories.map((category) => (
              <CategoryCard
                key={category.id}
                category={category}
                selected={selectedId === category.id}
                busy={busyId === category.id}
                onSelect={() => onSelect(category)}
                onEdit={() => onEdit(category)}
                onToggleStatus={() => onToggleStatus(category)}
                onDelete={() => onDelete(category)}
              />
            ))
          : null}
      </div>
    </section>
  );
}
