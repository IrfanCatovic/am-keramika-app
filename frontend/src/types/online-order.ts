export interface PublicCreateOrderItem {
  productID: number;
  quantity: number;
}

export interface PublicCreateOrderPayload {
  firstName: string;
  lastName: string;
  phone: string;
  city: string;
  address: string;
  email?: string;
  note?: string;
  /** Honeypot — must stay empty. */
  website?: string;
  items: PublicCreateOrderItem[];
}

export interface PublicOrderResponse {
  id: number;
  status: string;
  totalAmount: number;
  createdAt: string;
}

export interface PublicOrderErrorBody {
  message?: string;
  productID?: number;
  code?: string;
}

export interface CheckoutDraft {
  firstName: string;
  lastName: string;
  phone: string;
  city: string;
  address: string;
  email: string;
  note: string;
}
