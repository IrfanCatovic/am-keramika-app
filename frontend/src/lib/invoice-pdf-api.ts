import { ApiError } from "@/lib/api";
import { clearToken, getToken } from "@/lib/auth-token";

const API_URL = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ?? "";

function extractErrorMessage(payload: unknown, fallback: string): string {
  if (!payload || typeof payload !== "object") {
    return fallback;
  }
  const data = payload as Record<string, unknown>;
  if (typeof data.message === "string" && data.message.trim()) {
    return data.message;
  }
  if (typeof data.error === "string" && data.error.trim()) {
    return data.error;
  }
  return fallback;
}

export function invoicePdfFilename(invoiceId: number): string {
  return `AM-Keramika-Racun-${invoiceId}.pdf`;
}

/** Authenticated PDF fetch — returns Blob, never parses JSON success body. */
export async function fetchInvoicePdfBlob(
  invoiceId: number,
  options: { download?: boolean } = {},
): Promise<Blob> {
  if (!API_URL) {
    throw new ApiError(
      "API adresa nije podešena. Provjerite .env.local datoteku.",
      500,
    );
  }

  const token = getToken();
  if (!token) {
    throw new ApiError("Sesija je istekla. Prijavite se ponovo.", 401);
  }

  const query = options.download === false ? "" : "?download=true";
  const response = await fetch(
    `${API_URL}/invoices/${invoiceId}/pdf${query}`,
    {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
        Accept: "application/pdf",
      },
    },
  );

  if (!response.ok) {
    if (response.status === 401) {
      clearToken();
    }
    let payload: unknown = null;
    const rawText = await response.text();
    if (rawText) {
      try {
        payload = JSON.parse(rawText);
      } catch {
        payload = rawText;
      }
    }
    const fallback =
      response.status === 401
        ? "Sesija je istekla. Prijavite se ponovo."
        : response.status === 403
          ? "Nemate dozvolu za ovu akciju."
          : response.status === 404
            ? "Račun nije pronađen."
            : "Preuzimanje PDF-a nije uspjelo.";
    throw new ApiError(extractErrorMessage(payload, fallback), response.status, payload);
  }

  const blob = await response.blob();
  if (!blob || blob.size === 0) {
    throw new ApiError("PDF je prazan ili neispravan.", 500);
  }
  return blob;
}

export function downloadBlobAsFile(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  try {
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    anchor.rel = "noopener";
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
  } finally {
    window.setTimeout(() => URL.revokeObjectURL(url), 1000);
  }
}

export function isShareAbortError(err: unknown): boolean {
  if (!err || typeof err !== "object") {
    return false;
  }
  const name = "name" in err ? String((err as { name?: unknown }).name) : "";
  return name === "AbortError" || name === "NotAllowedError";
}

export async function canSharePdfFile(file: File): Promise<boolean> {
  if (typeof navigator === "undefined" || typeof navigator.share !== "function") {
    return false;
  }
  const data = { files: [file] };
  if (typeof navigator.canShare === "function") {
    try {
      return navigator.canShare(data);
    } catch {
      return false;
    }
  }
  return true;
}

export type ShareInvoicePdfResult =
  | { mode: "shared" }
  | { mode: "downloaded"; message: string }
  | { mode: "cancelled" };

export async function shareOrDownloadInvoicePdf(
  invoiceId: number,
): Promise<ShareInvoicePdfResult> {
  const blob = await fetchInvoicePdfBlob(invoiceId, { download: true });
  const filename = invoicePdfFilename(invoiceId);
  const file = new File([blob], filename, { type: "application/pdf" });

  if (await canSharePdfFile(file)) {
    try {
      await navigator.share({
        files: [file],
        title: `Račun #${invoiceId}`,
        text: `AM Keramika — račun #${invoiceId}`,
      });
      return { mode: "shared" };
    } catch (err) {
      if (isShareAbortError(err)) {
        return { mode: "cancelled" };
      }
      // Fall through to download fallback.
    }
  }

  downloadBlobAsFile(blob, filename);
  return {
    mode: "downloaded",
    message: "PDF je preuzet. Možete ga poslati iz aplikacije po izboru.",
  };
}

export async function downloadInvoicePdf(invoiceId: number): Promise<void> {
  const blob = await fetchInvoicePdfBlob(invoiceId, { download: true });
  downloadBlobAsFile(blob, invoicePdfFilename(invoiceId));
}
