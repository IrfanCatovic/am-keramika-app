"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { CustomerForm } from "@/components/customers/CustomerForm";
import { InlineError, ListSkeleton } from "@/components/ui/EmptyState";
import {
  fetchCustomer,
  getApiBusinessMessage,
  updateCustomer,
} from "@/lib/customers-api";
import { CustomerDetails } from "@/types/customer";

export default function EditCustomerPage() {
  const params = useParams();
  const router = useRouter();
  const id = Number(params.id);
  const invalidId = !Number.isFinite(id) || id <= 0;

  const [customer, setCustomer] = useState<CustomerDetails | null>(null);
  const [loading, setLoading] = useState(!invalidId);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loadError, setLoadError] = useState<string | null>(
    invalidId ? "Neispravan ID kupca." : null,
  );

  useEffect(() => {
    if (invalidId) {
      return;
    }

    let cancelled = false;
    async function run() {
      try {
        const data = await fetchCustomer(id);
        if (!cancelled) {
          setCustomer(data);
          setLoadError(null);
        }
      } catch (err) {
        if (!cancelled) {
          setCustomer(null);
          setLoadError(getApiBusinessMessage(err, "Kupac nije pronađen."));
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }
    void run();
    return () => {
      cancelled = true;
    };
  }, [id, invalidId]);

  if (loading) {
    return <ListSkeleton rows={3} />;
  }

  if (loadError || !customer) {
    return (
      <div className="space-y-3">
        <InlineError message={loadError ?? "Kupac nije pronađen."} />
        <Link href="/customers" className="text-sm font-medium text-[#8a6a45]">
          Nazad na listu
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto min-w-0 max-w-xl space-y-4">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight text-stone-900">
          Uredi kupca
        </h1>
      </header>

      <CustomerForm
        key={customer.id}
        mode="edit"
        initialName={customer.name}
        initialPhone={customer.phone}
        loading={saving}
        error={error}
        cancelHref={`/customers/${customer.id}`}
        onSubmit={async ({ name, phone }) => {
          setSaving(true);
          setError(null);
          try {
            await updateCustomer(customer.id, { name, phone });
            router.replace(`/customers/${customer.id}`);
          } catch (err) {
            setError(
              getApiBusinessMessage(err, "Ažuriranje kupca nije uspelo."),
            );
          } finally {
            setSaving(false);
          }
        }}
      />
    </div>
  );
}
