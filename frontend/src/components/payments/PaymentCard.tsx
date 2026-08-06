"use client";

import Link from "next/link";

import { formatMoney } from "@/lib/format";
import { paymentCustomerLabel } from "@/lib/payments-api";
import { Payment } from "@/types/payment";

export function PaymentCard({ payment }: { payment: Payment }) {
  const allocationCount = payment.allocations?.length ?? 0;

  return (
    <Link
      href={`/payments/${payment.id}`}
      className="block rounded-2xl border border-stone-200 bg-white p-4 transition hover:border-[#c4a484]/60 hover:bg-[#faf7f3]"
    >
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div className="min-w-0">
          <p className="font-semibold text-stone-900">Uplata #{payment.id}</p>
          <p className="mt-0.5 truncate text-sm text-stone-500">
            {paymentCustomerLabel(payment)}
          </p>
          <p className="mt-1 text-xs text-stone-400">{payment.createdAt}</p>
        </div>
        <div className="text-right">
          <p className="text-base font-semibold tabular-nums text-stone-900">
            {formatMoney(payment.totalAmount)}
          </p>
          <p className="mt-1 text-xs text-stone-500">
            {allocationCount}{" "}
            {allocationCount === 1 ? "račun" : "računa"}
          </p>
        </div>
      </div>
    </Link>
  );
}
