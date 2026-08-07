"use client";

import { useState } from "react";

export function PasswordField({
  id,
  label,
  value,
  onChange,
  autoComplete = "new-password",
  disabled,
  placeholder,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  autoComplete?: string;
  disabled?: boolean;
  placeholder?: string;
}) {
  const [visible, setVisible] = useState(false);

  return (
    <label className="block">
      <span className="mb-1.5 block text-sm font-medium text-stone-700">
        {label}
      </span>
      <div className="flex min-h-11 overflow-hidden rounded-xl border border-stone-200 bg-white focus-within:border-stone-400">
        <input
          id={id}
          type={visible ? "text" : "password"}
          value={value}
          disabled={disabled}
          autoComplete={autoComplete}
          placeholder={placeholder}
          onChange={(event) => onChange(event.target.value)}
          className="min-w-0 flex-1 bg-transparent px-3 text-sm text-stone-900 outline-none disabled:opacity-60"
        />
        <button
          type="button"
          disabled={disabled}
          onClick={() => setVisible((current) => !current)}
          className="shrink-0 border-l border-stone-200 px-3 text-xs font-medium text-stone-600 hover:bg-stone-50 disabled:opacity-60"
          aria-label={visible ? "Sakrij lozinku" : "Prikaži lozinku"}
        >
          {visible ? "Sakrij" : "Prikaži"}
        </button>
      </div>
    </label>
  );
}
