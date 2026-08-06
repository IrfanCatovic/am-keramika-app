"use client";

import { Suspense, use } from "react";
import { useSearchParams } from "next/navigation";

import { ProductForm } from "@/components/products/ProductForm";
import { ListSkeleton } from "@/components/ui/EmptyState";

function EditProductInner({ id }: { id: number }) {
  const searchParams = useSearchParams();
  const uploadWarning = searchParams.get("uploadWarning");

  return (
    <div className="space-y-3">
      {uploadWarning ? (
        <div className="rounded-xl border border-amber-100 bg-amber-50 px-4 py-3 text-sm text-amber-900">
          <p className="break-words">{uploadWarning}</p>
          <p className="mt-1 text-xs text-amber-800">
            Proizvod je sačuvan. Možete nastaviti uređivanje slika ispod.
          </p>
        </div>
      ) : null}
      <ProductForm mode="edit" productId={id} />
    </div>
  );
}

export default function EditProductPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const resolved = use(params);
  const id = Number(resolved.id);

  if (!Number.isFinite(id) || id <= 0) {
    return (
      <div className="rounded-xl border border-red-100 bg-red-50 px-4 py-3 text-sm text-red-700">
        Neispravan ID proizvoda.
      </div>
    );
  }

  return (
    <Suspense
      fallback={
        <div className="space-y-4">
          <div className="h-16 animate-pulse rounded-2xl bg-stone-100" />
          <ListSkeleton rows={4} />
        </div>
      }
    >
      <EditProductInner id={id} />
    </Suspense>
  );
}
