export class PublicAvailabilityError extends Error {
  status: number;

  constructor(message: string, status = 500) {
    super(message);
    this.name = "PublicAvailabilityError";
    this.status = status;
  }
}

export interface PublicAvailabilityCheckResult {
  available: boolean;
  reason?: string;
}

const API_URL = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ?? "";

const NETWORK_MESSAGE =
  "Trenutno nije moguće proveriti dostupnost. Pokušajte ponovo.";

export async function checkPublicProductAvailability(
  productId: number,
  quantity: number,
): Promise<PublicAvailabilityCheckResult> {
  if (!API_URL) {
    throw new PublicAvailabilityError(NETWORK_MESSAGE, 503);
  }

  let response: Response;
  try {
    response = await fetch(
      `${API_URL}/public/products/${productId}/check-availability`,
      {
        method: "POST",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ quantity }),
        cache: "no-store",
      },
    );
  } catch {
    throw new PublicAvailabilityError(NETWORK_MESSAGE, 503);
  }

  if (!response.ok) {
    if (response.status === 404) {
      return { available: false, reason: "unavailable" };
    }
    if (response.status === 400) {
      throw new PublicAvailabilityError(
        "Unesite ispravnu količinu veću od 0.",
        400,
      );
    }
    throw new PublicAvailabilityError(NETWORK_MESSAGE, response.status);
  }

  const body = (await response.json()) as PublicAvailabilityCheckResult;
  return {
    available: Boolean(body.available),
    reason: body.reason,
  };
}
