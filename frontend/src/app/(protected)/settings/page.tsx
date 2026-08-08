"use client";

import { ChangePasswordForm } from "@/components/settings/ChangePasswordForm";

export default function SettingsPage() {
  return (
    <div className="mx-auto w-full max-w-xl space-y-5 pb-4">
      <header className="space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight text-stone-900">
          Podešavanja
        </h1>
      </header>

      <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm sm:p-6">
        <h2 className="text-base font-semibold text-stone-900">
          Promjena lozinke
        </h2>
        <p className="mt-1 text-sm text-stone-500">
          Unesite trenutnu lozinku i zatim novu lozinku.
        </p>
        <div className="mt-5">
          <ChangePasswordForm />
        </div>
      </section>
    </div>
  );
}
