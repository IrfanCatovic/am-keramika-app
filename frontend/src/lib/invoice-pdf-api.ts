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
      "API adresa nije podešena. Proverite .env.local datoteku.",
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
            : "Preuzimanje PDF-a nije uspelo.";
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

export async function downloadInvoicePdf(invoiceId: number): Promise<void> {
  const blob = await fetchInvoicePdfBlob(invoiceId, { download: true });
  downloadBlobAsFile(blob, invoicePdfFilename(invoiceId));
}
