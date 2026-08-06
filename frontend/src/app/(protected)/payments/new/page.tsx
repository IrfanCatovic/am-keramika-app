"use client";

import { Suspense } from "react";
import { useSearchParams } from "next/navigation";

import { PaymentForm } from "@/components/payments/PaymentForm";
import { ListSkeleton } from "@/components/ui/EmptyState";

function NewPaymentContent() {
  const searchParams = useSearchParams();
  const invoiceRaw = searchParams.get("invoiceID");
  const customerRaw = searchParams.get("customerID");
  const invoiceParsed = invoiceRaw ? Number(invoiceRaw) : null;
  const customerParsed = customerRaw ? Number(customerRaw) : null;
  const initialInvoiceID =
    invoiceParsed && Number.isFinite(invoiceParsed) && invoiceParsed > 0
      ? invoiceParsed
      : null;
  const initialCustomerID =
    !initialInvoiceID &&
    customerParsed &&
    Number.isFinite(customerParsed) &&
    customerParsed > 0
      ? customerParsed
      : null;

  return (
    <PaymentForm
      initialInvoiceID={initialInvoiceID}
      initialCustomerID={initialCustomerID}
    />
  );
}

export default function NewPaymentPage() {
  return (
    <Suspense fallback={<ListSkeleton rows={4} />}>
      <NewPaymentContent />
    </Suspense>
  );
}
