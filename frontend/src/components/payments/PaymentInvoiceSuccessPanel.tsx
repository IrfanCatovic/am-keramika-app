"use client";

import Link from "next/link";

import { InvoiceSuccessPanel } from "@/components/invoices/InvoiceSuccessPanel";
import { InvoiceDetails } from "@/types/invoice";
import { Payment } from "@/types/payment";

export function PaymentInvoiceSuccessPanel({
  invoice,
  payment,
  customerLabel,
}: {
  invoice: InvoiceDetails;
  payment: Payment;
  customerLabel: string;
}) {
  return (
    <InvoiceSuccessPanel
      invoice={invoice}
      customerLabel={customerLabel}
      title="Uplata je uspješno evidentirana"
      extraActions={
        <Link
          href={`/payments/${payment.id}`}
          className="inline-flex min-h-11 w-full items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white hover:bg-stone-800"
        >
          Otvori uplatu
        </Link>
      }
    />
  );
}
