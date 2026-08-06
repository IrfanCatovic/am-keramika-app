"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";

import { PaymentDetails } from "@/components/payments/PaymentDetails";
import { InlineError, ListSkeleton } from "@/components/ui/EmptyState";
import {
  fetchPayment,
  getApiBusinessMessage,
} from "@/lib/payments-api";
import { Payment } from "@/types/payment";

export default function PaymentDetailPage() {
  const params = useParams();
  const raw = params?.id;
  const id = typeof raw === "string" ? Number(raw) : Number(raw?.[0]);
  const [payment, setPayment] = useState<Payment | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reloadToken, setReloadToken] = useState(0);

  useEffect(() => {
    if (!Number.isFinite(id) || id <= 0) {
      const timer = window.setTimeout(() => {
        setLoading(false);
        setError("Neispravan ID uplate.");
      }, 0);
      return () => window.clearTimeout(timer);
    }
    let cancelled = false;
    void (async () => {
      setLoading(true);
      try {
        const data = await fetchPayment(id);
        if (!cancelled) {
          setPayment(data);
          setError(null);
        }
      } catch (err) {
        if (!cancelled) {
          setPayment(null);
          setError(getApiBusinessMessage(err, "Uplata nije pronađena."));
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [id, reloadToken]);

  if (loading) {
    return <ListSkeleton rows={4} />;
  }
  if (error || !payment) {
    return (
      <div className="space-y-4">
        <InlineError
          message={error ?? "Uplata nije pronađena."}
          onRetry={() => setReloadToken((value) => value + 1)}
        />
        <Link href="/payments" className="text-sm font-medium text-[#8a6a45]">
          Lista uplata
        </Link>
      </div>
    );
  }
  return <PaymentDetails payment={payment} />;
}
