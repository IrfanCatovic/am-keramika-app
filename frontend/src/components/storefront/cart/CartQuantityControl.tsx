"use client";

import { useState } from "react";

export function CartQuantityControl({
  value,
  onChange,
  disabled = false,
  id,
  unit,
}: {
  value: number;
  onChange: (next: number) => void;
  disabled?: boolean;
  id?: string;
  unit?: string;
}) {
  const [draft, setDraft] = useState<string | null>(null);
  const display = draft ?? formatQty(value);

  function commit(raw: string) {
    const parsed = parseInput(raw);
    setDraft(null);
    if (parsed == null || parsed <= 0) {
      if (!(value > 0)) onChange(1);
      return;
    }
    if (parsed !== value) onChange(parsed);
  }

  return (
    <div className="inline-flex items-stretch overflow-hidden rounded-full border border-stone-300 bg-white">
      <button
        type="button"
        disabled={disabled || value <= 1}
        aria-label="Smanji količinu"
        className="flex h-11 w-11 items-center justify-center text-stone-700 transition hover:bg-stone-50 disabled:cursor-not-allowed disabled:opacity-40"
        onClick={() => {
          if (value > 1) onChange(Math.round((value - 1) * 100) / 100);
        }}
      >
        −
      </button>
      <input
        id={id}
        type="text"
        inputMode="decimal"
        disabled={disabled}
        value={display}
        aria-label={unit ? `Količina (${unit})` : "Količina"}
        className="w-16 border-x border-stone-200 bg-transparent text-center text-sm tabular-nums text-stone-900 outline-none disabled:opacity-50"
        onFocus={() => setDraft(formatQty(value))}
        onChange={(e) => setDraft(e.target.value)}
        onBlur={(e) => commit(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            e.currentTarget.blur();
          }
        }}
      />
      <button
        type="button"
        disabled={disabled}
        aria-label="Povećaj količinu"
        className="flex h-11 w-11 items-center justify-center text-stone-700 transition hover:bg-stone-50 disabled:cursor-not-allowed disabled:opacity-40"
        onClick={() => onChange(Math.round((value + 1) * 100) / 100)}
      >
        +
      </button>
    </div>
  );
}

function formatQty(n: number): string {
  if (!Number.isFinite(n)) return "1";
  return String(Math.round(n * 100) / 100);
}

function parseInput(raw: string): number | null {
  const normalized = raw.replace(",", ".").trim();
  if (!normalized) return null;
  const n = Number(normalized);
  if (!Number.isFinite(n)) return null;
  return Math.round(n * 100) / 100;
}
