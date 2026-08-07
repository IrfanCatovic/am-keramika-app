"use client";

import { useState } from "react";

import type { PublicProductImage } from "@/types/public-catalog";

export function ProductGallery({
  images,
  productName,
}: {
  images: PublicProductImage[];
  productName: string;
}) {
  const sorted = [...images].sort((a, b) => {
    if (a.isPrimary !== b.isPrimary) return a.isPrimary ? -1 : 1;
    return a.sortOrder - b.sortOrder || a.id - b.id;
  });
  const [activeIdx, setActiveIdx] = useState(0);
  const active = sorted[activeIdx] ?? sorted[0];

  if (!active) {
    return (
      <div className="flex aspect-square items-center justify-center rounded-3xl bg-gradient-to-br from-stone-100 to-stone-200/70">
        <span className="font-[family-name:var(--font-storefront-display)] text-5xl text-stone-300">
          AM
        </span>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="relative aspect-square overflow-hidden rounded-xl border border-stone-200 bg-white">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          src={active.url}
          alt={productName}
          className="h-full w-full object-contain p-6"
        />
      </div>
      {sorted.length > 1 ? (
        <div className="flex gap-2 overflow-x-auto pb-1">
          {sorted.map((image, index) => (
            <button
              key={image.id}
              type="button"
              onClick={() => setActiveIdx(index)}
              className={`relative h-20 w-20 shrink-0 overflow-hidden rounded-xl border bg-white transition ${
                index === activeIdx
                  ? "border-stone-900 ring-1 ring-stone-900"
                  : "border-stone-200 hover:border-stone-400"
              }`}
            >
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={image.url}
                alt=""
                className="h-full w-full object-contain p-1.5"
              />
            </button>
          ))}
        </div>
      ) : null}
    </div>
  );
}
