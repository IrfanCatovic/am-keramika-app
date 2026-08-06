export function EmptyState({
  title,
  description,
  action,
}: {
  title: string;
  description?: string;
  action?: React.ReactNode;
}) {
  return (
    <div className="rounded-2xl border border-dashed border-stone-200 bg-stone-50/80 px-4 py-8 text-center">
      <p className="text-sm font-medium text-stone-800">{title}</p>
      {description ? (
        <p className="mt-1 break-words text-sm text-stone-500">{description}</p>
      ) : null}
      {action ? <div className="mt-4 flex justify-center">{action}</div> : null}
    </div>
  );
}

export function StatusBadge({ active }: { active: boolean }) {
  if (active) {
    return (
      <span className="inline-flex items-center rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-800 ring-1 ring-inset ring-emerald-200">
        Aktivna
      </span>
    );
  }

  return (
    <span className="inline-flex items-center rounded-md bg-stone-100 px-2 py-0.5 text-xs font-medium text-stone-600 ring-1 ring-inset ring-stone-200">
      Neaktivna
    </span>
  );
}

export function InlineError({
  message,
  onRetry,
}: {
  message: string;
  onRetry?: () => void;
}) {
  return (
    <div className="rounded-xl border border-red-100 bg-red-50/80 px-4 py-4">
      <p className="break-words text-sm text-red-700">{message}</p>
      {onRetry ? (
        <button
          type="button"
          onClick={onRetry}
          className="mt-3 inline-flex min-h-10 items-center rounded-lg border border-red-200 bg-white px-3 text-sm font-medium text-red-800 transition hover:bg-red-50"
        >
          Pokušaj ponovo
        </button>
      ) : null}
    </div>
  );
}

export function ListSkeleton({ rows = 4 }: { rows?: number }) {
  return (
    <div className="space-y-3" aria-hidden>
      {Array.from({ length: rows }).map((_, index) => (
        <div
          key={index}
          className="h-24 animate-pulse rounded-2xl bg-stone-100"
        />
      ))}
    </div>
  );
}
