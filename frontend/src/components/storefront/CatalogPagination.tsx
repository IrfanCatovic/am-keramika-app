import Link from "next/link";

function buildPageHref(
  basePath: string,
  query: Record<string, string | undefined>,
  nextPage: number,
): string {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value) params.set(key, value);
  }
  if (nextPage > 1) {
    params.set("page", String(nextPage));
  }
  const qs = params.toString();
  return qs ? `${basePath}?${qs}` : basePath;
}

export function CatalogPagination({
  page,
  totalPages,
  basePath,
  query = {},
}: {
  page: number;
  totalPages: number;
  basePath: string;
  query?: Record<string, string | undefined>;
}) {
  if (totalPages <= 1) return null;

  return (
    <div className="mt-10 flex flex-wrap items-center justify-center gap-2">
      {page > 1 ? (
        <Link
          href={buildPageHref(basePath, query, page - 1)}
          className="rounded-full border border-stone-200 bg-white px-4 py-2 text-sm text-stone-700"
        >
          Prethodna
        </Link>
      ) : null}
      <span className="px-2 text-sm text-stone-500">
        Stranica {page} / {totalPages}
      </span>
      {page < totalPages ? (
        <Link
          href={buildPageHref(basePath, query, page + 1)}
          className="rounded-full border border-stone-200 bg-white px-4 py-2 text-sm text-stone-700"
        >
          Sledeća
        </Link>
      ) : null}
    </div>
  );
}
