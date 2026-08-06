"use client";

import { Modal } from "@/components/ui/Modal";

export function ConfirmDialog({
  open,
  title,
  message,
  confirmLabel = "Potvrdi",
  cancelLabel = "Otkaži",
  loading = false,
  error = null,
  tone = "danger",
  onConfirm,
  onClose,
}: {
  open: boolean;
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  loading?: boolean;
  error?: string | null;
  tone?: "danger" | "neutral";
  onConfirm: () => void;
  onClose: () => void;
}) {
  const confirmClass =
    tone === "danger"
      ? "bg-red-700 text-white hover:bg-red-800"
      : "bg-stone-900 text-white hover:bg-stone-800";

  return (
    <Modal open={open} title={title} onClose={loading ? () => undefined : onClose}>
      <div className="space-y-4">
        <p className="break-words text-sm leading-relaxed text-stone-600">
          {message}
        </p>
        {error ? (
          <p className="break-words rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
            {error}
          </p>
        ) : null}
        <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <button
            type="button"
            disabled={loading}
            onClick={onClose}
            className="inline-flex min-h-11 items-center justify-center rounded-xl border border-stone-200 px-4 text-sm font-medium text-stone-700 transition hover:bg-stone-50 disabled:opacity-60"
          >
            {cancelLabel}
          </button>
          <button
            type="button"
            disabled={loading}
            onClick={onConfirm}
            className={`inline-flex min-h-11 items-center justify-center rounded-xl px-4 text-sm font-medium transition disabled:opacity-60 ${confirmClass}`}
          >
            {loading ? "Sačekajte..." : confirmLabel}
          </button>
        </div>
      </div>
    </Modal>
  );
}
