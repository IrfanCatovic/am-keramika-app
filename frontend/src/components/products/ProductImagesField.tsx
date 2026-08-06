"use client";

import Image from "next/image";
import { useRef, useState } from "react";

import { ConfirmDialog } from "@/components/ui/ConfirmDialog";
import { ProductImage } from "@/types/product";

export type PendingImage = {
  key: string;
  file: File;
  previewUrl: string;
};

const MAX_IMAGES = 8;

export function ProductImagesField({
  mode,
  existingImages = [],
  pendingImages,
  onPendingChange,
  busy = false,
  onSetPrimary,
  onMove,
  onDeleteExisting,
  onUploadFiles,
}: {
  mode: "create" | "edit";
  existingImages?: ProductImage[];
  pendingImages: PendingImage[];
  onPendingChange: (images: PendingImage[]) => void;
  busy?: boolean;
  onSetPrimary?: (imageId: number) => Promise<void> | void;
  onMove?: (imageId: number, direction: "up" | "down") => Promise<void> | void;
  onDeleteExisting?: (imageId: number) => Promise<void> | void;
  onUploadFiles?: (files: File[]) => Promise<void> | void;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [confirmDeleteId, setConfirmDeleteId] = useState<number | null>(null);
  const [confirmLoading, setConfirmLoading] = useState(false);

  const existingCount = existingImages.length;
  const pendingCount = pendingImages.length;
  const totalCount = existingCount + pendingCount;
  const slotsLeft = Math.max(0, MAX_IMAGES - totalCount);

  function revokePending(images: PendingImage[]) {
    for (const image of images) {
      URL.revokeObjectURL(image.previewUrl);
    }
  }

  function handlePickFiles(fileList: FileList | null) {
    if (!fileList || fileList.length === 0) {
      return;
    }
    setActionError(null);
    const selected = Array.from(fileList).filter((file) =>
      file.type.startsWith("image/"),
    );
    if (selected.length === 0) {
      setActionError("Izaberite slike (JPEG, PNG, WebP...).");
      return;
    }

    const allowed = selected.slice(0, slotsLeft);
    if (allowed.length < selected.length) {
      setActionError(`Maksimalno ${MAX_IMAGES} slika po proizvodu.`);
    }

    if (mode === "edit" && onUploadFiles) {
      void onUploadFiles(allowed);
      if (inputRef.current) {
        inputRef.current.value = "";
      }
      return;
    }

    const next: PendingImage[] = allowed.map((file) => ({
      key: `${file.name}-${file.size}-${file.lastModified}-${Math.random()}`,
      file,
      previewUrl: URL.createObjectURL(file),
    }));
    onPendingChange([...pendingImages, ...next]);
    if (inputRef.current) {
      inputRef.current.value = "";
    }
  }

  function removePending(key: string) {
    const remaining = pendingImages.filter((image) => image.key !== key);
    const removed = pendingImages.filter((image) => image.key === key);
    revokePending(removed);
    onPendingChange(remaining);
  }

  async function handleConfirmDelete() {
    if (confirmDeleteId == null || !onDeleteExisting) {
      return;
    }
    setConfirmLoading(true);
    setActionError(null);
    try {
      await onDeleteExisting(confirmDeleteId);
      setConfirmDeleteId(null);
    } catch (error) {
      setActionError(
        error instanceof Error ? error.message : "Brisanje slike nije uspjelo.",
      );
    } finally {
      setConfirmLoading(false);
    }
  }

  const sortedExisting = [...existingImages].sort(
    (a, b) => a.sortOrder - b.sortOrder || a.id - b.id,
  );

  return (
    <section className="space-y-3">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h3 className="text-sm font-semibold text-stone-900">Slike</h3>
          <p className="mt-0.5 text-xs text-stone-500">
            Do {MAX_IMAGES} slika. Prva postaje primarna ako nije drugačije
            označeno.
          </p>
        </div>
        <button
          type="button"
          disabled={busy || slotsLeft === 0}
          onClick={() => inputRef.current?.click()}
          className="inline-flex min-h-11 items-center rounded-xl border border-stone-200 bg-white px-4 text-sm font-medium text-stone-700 transition hover:bg-stone-50 disabled:opacity-60"
        >
          {mode === "edit" ? "Dodaj slike" : "Izaberi slike"}
        </button>
        <input
          ref={inputRef}
          type="file"
          accept="image/*"
          multiple
          className="hidden"
          onChange={(event) => handlePickFiles(event.target.files)}
        />
      </div>

      {actionError ? (
        <p className="break-words rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
          {actionError}
        </p>
      ) : null}

      {sortedExisting.length === 0 && pendingImages.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-stone-200 bg-stone-50/80 px-4 py-6 text-center text-sm text-stone-500">
          Nema slika.
        </div>
      ) : null}

      <ul className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        {sortedExisting.map((image, index) => (
          <li
            key={image.id}
            className="flex gap-3 rounded-2xl border border-stone-200 bg-white p-3"
          >
            <div className="relative h-20 w-20 shrink-0 overflow-hidden rounded-xl bg-stone-100 ring-1 ring-stone-200">
              <Image
                src={image.url}
                alt=""
                fill
                className="object-cover"
                sizes="80px"
                unoptimized
              />
            </div>
            <div className="min-w-0 flex-1 space-y-2">
              <div className="flex flex-wrap items-center gap-2">
                {image.isPrimary ? (
                  <span className="inline-flex rounded-md bg-[#faf6f1] px-2 py-0.5 text-xs font-medium text-[#8a6a45] ring-1 ring-inset ring-[#c4a484]/50">
                    Primarna
                  </span>
                ) : (
                  <button
                    type="button"
                    disabled={busy}
                    onClick={() => void onSetPrimary?.(image.id)}
                    className="text-xs font-medium text-stone-600 underline-offset-2 hover:underline disabled:opacity-60"
                  >
                    Postavi kao primarnu
                  </button>
                )}
              </div>
              <div className="flex flex-wrap gap-2">
                <button
                  type="button"
                  disabled={busy || index === 0}
                  onClick={() => void onMove?.(image.id, "up")}
                  className="inline-flex min-h-9 items-center rounded-lg border border-stone-200 px-2.5 text-xs font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-50"
                >
                  Gore
                </button>
                <button
                  type="button"
                  disabled={busy || index >= sortedExisting.length - 1}
                  onClick={() => void onMove?.(image.id, "down")}
                  className="inline-flex min-h-9 items-center rounded-lg border border-stone-200 px-2.5 text-xs font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-50"
                >
                  Dole
                </button>
                <button
                  type="button"
                  disabled={busy}
                  onClick={() => setConfirmDeleteId(image.id)}
                  className="inline-flex min-h-9 items-center rounded-lg border border-red-200 px-2.5 text-xs font-medium text-red-700 hover:bg-red-50 disabled:opacity-50"
                >
                  Obriši
                </button>
              </div>
            </div>
          </li>
        ))}

        {pendingImages.map((image) => (
          <li
            key={image.key}
            className="flex gap-3 rounded-2xl border border-dashed border-stone-300 bg-stone-50 p-3"
          >
            <div className="relative h-20 w-20 shrink-0 overflow-hidden rounded-xl bg-stone-100 ring-1 ring-stone-200">
              <Image
                src={image.previewUrl}
                alt=""
                fill
                className="object-cover"
                sizes="80px"
                unoptimized
              />
            </div>
            <div className="min-w-0 flex-1 space-y-2">
              <p className="truncate text-xs text-stone-500">{image.file.name}</p>
              <p className="text-xs text-stone-400">Čeka upload nakon čuvanja</p>
              <button
                type="button"
                disabled={busy}
                onClick={() => removePending(image.key)}
                className="inline-flex min-h-9 items-center rounded-lg border border-red-200 px-2.5 text-xs font-medium text-red-700 hover:bg-red-50 disabled:opacity-50"
              >
                Ukloni
              </button>
            </div>
          </li>
        ))}
      </ul>

      <ConfirmDialog
        open={confirmDeleteId != null}
        title="Obriši sliku"
        message="Da li ste sigurni da želite obrisati ovu sliku?"
        confirmLabel="Obriši"
        tone="danger"
        loading={confirmLoading}
        onClose={() => {
          if (!confirmLoading) {
            setConfirmDeleteId(null);
          }
        }}
        onConfirm={() => void handleConfirmDelete()}
      />
    </section>
  );
}
