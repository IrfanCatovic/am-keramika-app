"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { CustomerForm } from "@/components/customers/CustomerForm";
import {
  createCustomer,
  getApiBusinessMessage,
} from "@/lib/customers-api";

export default function NewCustomerPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  return (
    <div className="mx-auto min-w-0 max-w-xl space-y-4">
      <header>
        <Link
          href="/customers"
          className="text-sm font-medium text-[#8a6a45] hover:text-stone-900"
        >
          ← Nazad na kupce
        </Link>
        <h1 className="mt-2 text-2xl font-semibold tracking-tight text-stone-900">
          Novi kupac
        </h1>
        <p className="mt-1 text-sm text-stone-500">
          Unesite ime/naziv i opciono telefon. Email i adresa nisu dio backend
          DTO-a.
        </p>
      </header>

      <CustomerForm
        mode="create"
        loading={loading}
        error={error}
        cancelHref="/customers"
        onSubmit={async ({ name, phone }) => {
          setLoading(true);
          setError(null);
          try {
            const response = await createCustomer({ name, phone });
            router.replace(`/customers/${response.customer.id}`);
          } catch (err) {
            setError(
              getApiBusinessMessage(err, "Kreiranje kupca nije uspjelo."),
            );
          } finally {
            setLoading(false);
          }
        }}
      />
    </div>
  );
}
