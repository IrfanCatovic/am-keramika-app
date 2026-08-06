"use client";

import {
  EmptyState,
  InlineError,
  ListSkeleton,
} from "@/components/ui/EmptyState";
import { Category } from "@/types/category";
import { ProductGroup } from "@/types/product-group";

function GroupCard({
  group,
  busy,
  onEdit,
  onDelete,
}: {
  group: ProductGroup;
  busy: boolean;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <li className="min-w-0 rounded-2xl border border-stone-200 bg-white p-3.5 sm:p-4">
      <div className="min-w-0">
        <p className="break-words font-medium text-stone-900">{group.name}</p>
        {group.slug ? (
          <p className="mt-1 break-all text-xs text-stone-500">{group.slug}</p>
        ) : null}
      </div>
      <div className="mt-3 flex flex-wrap gap-2">
        <button
          type="button"
          disabled={busy}
          onClick={onEdit}
          className="inline-flex min-h-10 items-center rounded-xl border border-stone-200 px-3 text-sm font-medium text-stone-700 transition hover:bg-stone-50 disabled:opacity-60"
        >
          Uredi
        </button>
        <button
          type="button"
          disabled={busy}
          onClick={onDelete}
          className="inline-flex min-h-10 items-center rounded-xl border border-red-200 px-3 text-sm font-medium text-red-700 transition hover:bg-red-50 disabled:opacity-60"
        >
          Obriši
        </button>
      </div>
    </li>
  );
}

export function ProductGroupList({
  category,
  groups,
  loading,
  error,
  busyId,
  onRetry,
  onCreate,
  onEdit,
  onDelete,
}: {
  category: Category | null;
  groups: ProductGroup[];
  loading: boolean;
  error: string | null;
  busyId: number | null;
  onRetry: () => void;
  onCreate: () => void;
  onEdit: (group: ProductGroup) => void;
  onDelete: (group: ProductGroup) => void;
}) {
  if (!category) {
    return (
      <section className="min-w-0 rounded-2xl border border-stone-200/90 bg-white shadow-[0_1px_2px_rgba(28,25,23,0.04)]">
        <div className="border-b border-stone-100 px-4 py-3.5 sm:px-5 sm:py-4">
          <h2 className="text-base font-semibold tracking-tight text-stone-900">
            Grupe proizvoda
          </h2>
        </div>
        <div className="px-4 py-4 sm:px-5">
          <EmptyState
            title="Izaberite kategoriju"
            description="Odaberite kategoriju s lijeve strane da vidite njene grupe."
          />
        </div>
      </section>
    );
  }

  const canAddGroup = category.isActive;

  return (
    <section className="min-w-0 rounded-2xl border border-stone-200/90 bg-white shadow-[0_1px_2px_rgba(28,25,23,0.04)]">
      <div className="flex flex-wrap items-start justify-between gap-3 border-b border-stone-100 px-4 py-3.5 sm:px-5 sm:py-4">
        <div className="min-w-0">
          <h2 className="text-base font-semibold tracking-tight text-stone-900">
            Grupe proizvoda
          </h2>
          <p className="mt-0.5 break-words text-sm text-stone-500">
            {category.name}
          </p>
        </div>
        <button
          type="button"
          disabled={!canAddGroup}
          onClick={onCreate}
          title={
            canAddGroup
              ? undefined
              : "Aktivirajte kategoriju da biste dodali grupu"
          }
          className="inline-flex min-h-11 shrink-0 items-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white transition hover:bg-stone-800 disabled:cursor-not-allowed disabled:bg-stone-300 disabled:text-stone-600"
        >
          Nova grupa
        </button>
      </div>

      <div className="space-y-3 px-4 py-4 sm:px-5">
        {!canAddGroup ? (
          <p className="rounded-xl border border-amber-100 bg-amber-50/80 px-3 py-2.5 text-sm text-amber-900">
            Kategorija je neaktivna. Aktivirajte je prije dodavanja novih grupa.
            Pregled postojećih grupa je i dalje dostupan.
          </p>
        ) : null}

        {loading ? <ListSkeleton rows={3} /> : null}

        {!loading && error ? (
          <InlineError message={error} onRetry={onRetry} />
        ) : null}

        {!loading && !error && groups.length === 0 ? (
          <EmptyState
            title="Nema grupa"
            description={
              canAddGroup
                ? "Dodajte prvu grupu u ovoj kategoriji."
                : "Ova kategorija trenutno nema grupa."
            }
            action={
              canAddGroup ? (
                <button
                  type="button"
                  onClick={onCreate}
                  className="inline-flex min-h-11 items-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white"
                >
                  Dodaj grupu
                </button>
              ) : undefined
            }
          />
        ) : null}

        {!loading && !error && groups.length > 0 ? (
          <ul className="space-y-3">
            {groups.map((group) => (
              <GroupCard
                key={group.id}
                group={group}
                busy={busyId === group.id}
                onEdit={() => onEdit(group)}
                onDelete={() => onDelete(group)}
              />
            ))}
          </ul>
        ) : null}
      </div>
    </section>
  );
}
