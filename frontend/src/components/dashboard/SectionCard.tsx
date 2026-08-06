"use client";

import { SectionStatus } from "@/hooks/useAsyncSection";

export function SectionCard({
  title,
  description,
  action,
  children,
}: {
  title: string;
  description?: string;
  action?: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <section className="dash-enter rounded-2xl border border-stone-200/80 bg-white shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
      <div className="flex flex-wrap items-start justify-between gap-3 border-b border-stone-100 px-5 py-4">
        <div>
          <h2 className="text-base font-semibold tracking-tight text-stone-900">
            {title}
          </h2>
          {description ? (
            <p className="mt-0.5 text-sm text-stone-500">{description}</p>
          ) : null}
        </div>
        {action}
      </div>
      <div className="px-5 py-4">{children}</div>
    </section>
  );
}

export function SectionRetry({
  message,
  onRetry,
}: {
  message: string;
  onRetry: () => void;
}) {
  return (
    <div className="flex flex-col items-start gap-3 rounded-xl border border-red-100 bg-red-50/70 px-4 py-4">
      <p className="text-sm text-red-700">{message}</p>
      <button
        type="button"
        onClick={onRetry}
        className="rounded-lg border border-red-200 bg-white px-3 py-1.5 text-sm font-medium text-red-800 transition hover:bg-red-50"
      >
        Pokušaj ponovo
      </button>
    </div>
  );
}

export function SectionEmpty({ message }: { message: string }) {
  return (
    <div className="rounded-xl border border-dashed border-stone-200 bg-stone-50/80 px-4 py-8 text-center text-sm text-stone-500">
      {message}
    </div>
  );
}

export function SkeletonBlock({ className = "" }: { className?: string }) {
  return (
    <div
      className={`animate-pulse rounded-xl bg-stone-100 ${className}`}
      aria-hidden
    />
  );
}

export function SectionBody({
  status,
  error,
  onRetry,
  loadingFallback,
  children,
}: {
  status: SectionStatus;
  error: string | null;
  onRetry: () => void;
  loadingFallback: React.ReactNode;
  children: React.ReactNode;
}) {
  if (status === "loading") {
    return loadingFallback;
  }
  if (status === "error") {
    return <SectionRetry message={error ?? "Došlo je do greške."} onRetry={onRetry} />;
  }
  return children;
}
