"use client";

export function EmployeeStatusBadge({ isActive }: { isActive: boolean }) {
  if (isActive) {
    return (
      <span className="inline-flex items-center rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-800 ring-1 ring-inset ring-emerald-200">
        Aktivan
      </span>
    );
  }
  return (
    <span className="inline-flex items-center rounded-md bg-stone-100 px-2 py-0.5 text-xs font-medium text-stone-600 ring-1 ring-inset ring-stone-200">
      Neaktivan
    </span>
  );
}
