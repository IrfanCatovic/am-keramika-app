"use client";

import { FormEvent, useState } from "react";

import { Modal } from "@/components/ui/Modal";

export function CancelInvoiceDialog({
  open,
  invoiceId,
  loading,
  error,
  onClose,
  onConfirm,
}: {
  open: boolean;
  invoiceId?: number;
  loading: boolean;
  error: string | null;
  onClose: () => void;
  onConfirm: (reason: string) => void;
}) {
  const [reason, setReason] = useState("");
  const [localError, setLocalError] = useState<string | null>(null);

  const confirmMessage =
    invoiceId != null
      ? `Da li ste sigurni da želite da stornirate račun #${invoiceId}? Nakon potvrde račun će biti označen kao storniran.`
      : "Da li ste sigurni da želite da stornirate ovaj račun? Nakon potvrde račun će biti označen kao storniran.";

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    const trimmed = reason.trim();
    if (trimmed.length < 3) {
      setLocalError("Razlog mora imati najmanje 3 karaktera.");
      return;
    }
    setLocalError(null);
    onConfirm(trimmed);
  }

  return (
    <Modal
      open={open}
      title="Storniranje računa"
      onClose={loading ? () => undefined : onClose}
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <p className="text-sm leading-relaxed text-stone-600">
          {confirmMessage}
        </p>
        <div>
          <label
            htmlFor="cancel-reason"
            className="mb-1.5 block text-sm font-medium text-stone-700"
          >
            Razlog storniranja *
          </label>
          <textarea
            id="cancel-reason"
            value={reason}
            disabled={loading}
            onChange={(event) => setReason(event.target.value)}
            rows={3}
            className="w-full rounded-xl border border-stone-200 px-3 py-2.5 text-sm outline-none ring-[#c4a484]/40 focus:ring-2 disabled:opacity-60"
            placeholder="npr. Greška u količini"
          />
        </div>
        {(localError || error) && (
          <p className="break-words rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
            {localError || error}
          </p>
        )}
        <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <button
            type="button"
            disabled={loading}
            onClick={onClose}
            className="inline-flex min-h-11 items-center justify-center rounded-xl border border-stone-200 px-4 text-sm font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-60"
          >
            Odustani
          </button>
          <button
            type="submit"
            disabled={loading}
            className="inline-flex min-h-11 items-center justify-center rounded-xl bg-red-700 px-4 text-sm font-medium text-white hover:bg-red-800 disabled:opacity-60"
          >
            {loading ? "Storniranje…" : "Storniraj račun"}
          </button>
        </div>
      </form>
    </Modal>
  );
}
