import type {
  CheckoutDraft,
  PublicCreateOrderPayload,
  PublicOrderErrorBody,
  PublicOrderResponse,
} from "@/types/online-order";

export const CHECKOUT_DRAFT_KEY = "am-keramika-checkout-draft-v1";

export class PublicOrderError extends Error {
  status: number;
  code?: string;
  productID?: number;

  constructor(
    message: string,
    status = 500,
    extras?: { code?: string; productID?: number },
  ) {
    super(message);
    this.name = "PublicOrderError";
    this.status = status;
    this.code = extras?.code;
    this.productID = extras?.productID;
  }
}

const API_URL = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ?? "";

const NETWORK_MESSAGE =
  "Narudžbinu trenutno nije moguće poslati. Pokušajte ponovo.";

export async function createPublicOrder(
  payload: PublicCreateOrderPayload,
): Promise<PublicOrderResponse> {
  if (!API_URL) {
    throw new PublicOrderError(NETWORK_MESSAGE, 503);
  }

  let response: Response;
  try {
    response = await fetch(`${API_URL}/public/orders`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
      cache: "no-store",
    });
  } catch {
    throw new PublicOrderError(NETWORK_MESSAGE, 503);
  }

  if (!response.ok) {
    let message = NETWORK_MESSAGE;
    let code: string | undefined;
    let productID: number | undefined;
    try {
      const body = (await response.json()) as PublicOrderErrorBody;
      if (body.message?.trim()) message = body.message.trim();
      code = body.code;
      productID = body.productID;
    } catch {
      /* ignore */
    }
    if (response.status >= 500) {
      message = NETWORK_MESSAGE;
    }
    throw new PublicOrderError(message, response.status, { code, productID });
  }

  return (await response.json()) as PublicOrderResponse;
}

export function readCheckoutDraft(): CheckoutDraft | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = window.sessionStorage.getItem(CHECKOUT_DRAFT_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as Partial<CheckoutDraft>;
    return {
      firstName: String(parsed.firstName ?? ""),
      lastName: String(parsed.lastName ?? ""),
      phone: String(parsed.phone ?? ""),
      city: String(parsed.city ?? ""),
      address: String(parsed.address ?? ""),
      email: String(parsed.email ?? ""),
      note: String(parsed.note ?? ""),
    };
  } catch {
    return null;
  }
}

export function writeCheckoutDraft(draft: CheckoutDraft): void {
  if (typeof window === "undefined") return;
  try {
    window.sessionStorage.setItem(CHECKOUT_DRAFT_KEY, JSON.stringify(draft));
  } catch {
    /* ignore */
  }
}

export function clearCheckoutDraft(): void {
  if (typeof window === "undefined") return;
  try {
    window.sessionStorage.removeItem(CHECKOUT_DRAFT_KEY);
  } catch {
    /* ignore */
  }
}
