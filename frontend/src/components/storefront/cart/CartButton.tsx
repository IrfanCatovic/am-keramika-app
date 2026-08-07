"use client";

import { useCart } from "@/components/storefront/cart/CartProvider";

function BagIcon({ className = "" }: { className?: string }) {
  return (
    <svg
      className={className}
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.6"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M6 8h12l-1 12H7L6 8Z" />
      <path d="M9 8V7a3 3 0 0 1 6 0v1" />
    </svg>
  );
}

export function CartButton({
  className = "",
  onBeforeOpen,
}: {
  className?: string;
  onBeforeOpen?: () => void;
}) {
  const { itemCount, hydrated, openDrawer } = useCart();
  const showBadge = hydrated && itemCount > 0;

  return (
    <button
      type="button"
      onClick={() => {
        onBeforeOpen?.();
        openDrawer();
      }}
      className={`relative inline-flex h-10 w-10 items-center justify-center rounded-full border border-stone-800/80 text-stone-900 transition hover:bg-stone-900 hover:text-white ${className}`}
      aria-label={
        showBadge ? `Korpa, ${itemCount} proizvoda` : "Korpa"
      }
    >
      <BagIcon />
      {showBadge ? (
        <span className="absolute -right-1 -top-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-[#141311] px-1 text-[10px] font-medium tabular-nums text-[#e8d4b8]">
          {itemCount > 99 ? "99+" : itemCount}
        </span>
      ) : null}
    </button>
  );
}
