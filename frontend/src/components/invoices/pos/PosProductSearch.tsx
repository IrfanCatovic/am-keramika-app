"use client";

import {
  forwardRef,
  useEffect,
  useImperativeHandle,
  useRef,
  type KeyboardEvent,
} from "react";

export type PosProductSearchHandle = {
  focus: () => void;
  clearAndFocus: () => void;
};

/**
 * Glavni POS search. Barcode podrška se može kasnije dodati
 * (npr. poseban input / scanner mode) bez mijenjanja invoice toka.
 */
export const PosProductSearch = forwardRef<
  PosProductSearchHandle,
  {
    value: string;
    onChange: (value: string) => void;
    onKeyDown?: (event: KeyboardEvent<HTMLInputElement>) => void;
    loading?: boolean;
    autoFocus?: boolean;
  }
>(function PosProductSearch(
  { value, onChange, onKeyDown, loading = false, autoFocus = true },
  ref,
) {
  const inputRef = useRef<HTMLInputElement>(null);

  useImperativeHandle(ref, () => ({
    focus() {
      inputRef.current?.focus();
    },
    clearAndFocus() {
      onChange("");
      window.requestAnimationFrame(() => {
        inputRef.current?.focus();
      });
    },
  }));

  useEffect(() => {
    if (!autoFocus) {
      return;
    }
    const timer = window.setTimeout(() => {
      inputRef.current?.focus();
    }, 40);
    return () => window.clearTimeout(timer);
  }, [autoFocus]);

  return (
    <div className="relative min-w-0">
      <label htmlFor="pos-product-search" className="sr-only">
        Pretraga proizvoda
      </label>
      <div className="pointer-events-none absolute inset-y-0 left-3 flex items-center text-stone-400">
        <svg viewBox="0 0 24 24" className="h-5 w-5" aria-hidden>
          <path
            d="M10.5 3a7.5 7.5 0 015.95 12.1l3.72 3.73a1 1 0 01-1.42 1.41l-3.72-3.72A7.5 7.5 0 1110.5 3zm0 2a5.5 5.5 0 100 11 5.5 5.5 0 000-11z"
            fill="currentColor"
          />
        </svg>
      </div>
      <input
        ref={inputRef}
        id="pos-product-search"
        type="search"
        autoComplete="off"
        spellCheck={false}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        onKeyDown={onKeyDown}
        placeholder="Traži proizvod po nazivu…"
        className="w-full rounded-2xl border border-stone-200 bg-white py-3.5 pl-11 pr-12 text-base text-stone-900 shadow-[0_1px_2px_rgba(28,25,23,0.04)] outline-none ring-[#c4a484]/35 transition placeholder:text-stone-400 focus:border-stone-400 focus:ring-2"
      />
      {loading ? (
        <span className="absolute inset-y-0 right-3 flex items-center text-xs text-stone-400">
          …
        </span>
      ) : null}
      {/* Slot za budući barcode scanner trigger — ne renderuje UI dok backend nema barcode. */}
      <span data-pos-barcode-slot className="hidden" aria-hidden />
    </div>
  );
});
