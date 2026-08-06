"use client";

import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/components/auth/AuthProvider";
import { ApiError } from "@/lib/api";

export default function LoginPage() {
  const { login, user, isLoading } = useAuth();
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!isLoading && user) {
      router.replace("/dashboard");
    }
  }, [isLoading, user, router]);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);

    const trimmedUsername = username.trim();
    if (!trimmedUsername || !password) {
      setError("Unesite korisničko ime i lozinku.");
      return;
    }

    setSubmitting(true);
    try {
      await login(trimmedUsername, password);
      router.replace("/dashboard");
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("Prijava nije uspjela. Pokušajte ponovo.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  if (isLoading || user) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-100">
        <div className="rounded-lg border border-slate-200 bg-white px-6 py-4 text-sm text-slate-600 shadow-sm">
          {user ? "Već ste prijavljeni..." : "Učitavanje..."}
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-100 via-slate-50 to-stone-100 px-4">
      <div className="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-8 shadow-sm">
        <div className="mb-8">
          <p className="text-sm font-medium uppercase tracking-[0.16em] text-slate-500">
            AM Keramika
          </p>
          <h1 className="mt-2 text-2xl font-semibold tracking-tight text-slate-900">
            Prijava
          </h1>
          <p className="mt-2 text-sm text-slate-500">
            Unesite pristupne podatke za internu aplikaciju.
          </p>
        </div>

        <form onSubmit={onSubmit} className="space-y-4">
          <div>
            <label
              htmlFor="username"
              className="mb-1.5 block text-sm font-medium text-slate-700"
            >
              Korisničko ime
            </label>
            <input
              id="username"
              name="username"
              autoComplete="username"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm text-slate-900 outline-none ring-slate-400 focus:ring-2"
              disabled={submitting}
            />
          </div>

          <div>
            <label
              htmlFor="password"
              className="mb-1.5 block text-sm font-medium text-slate-700"
            >
              Lozinka
            </label>
            <input
              id="password"
              name="password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm text-slate-900 outline-none ring-slate-400 focus:ring-2"
              disabled={submitting}
            />
          </div>

          {error ? (
            <div className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {error}
            </div>
          ) : null}

          <button
            type="submit"
            disabled={submitting}
            className="w-full rounded-lg bg-slate-900 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {submitting ? "Prijava..." : "Prijavi se"}
          </button>
        </form>
      </div>
    </div>
  );
}
