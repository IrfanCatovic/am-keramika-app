"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/components/auth/AuthProvider";
import { PasswordField } from "@/components/users/PasswordField";
import { ApiError } from "@/lib/api";
import { changePassword } from "@/lib/auth-api";
import { MIN_PASSWORD_LENGTH } from "@/types/user";

export function ChangePasswordForm() {
  const { logout } = useAuth();
  const router = useRouter();
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setSuccess(null);

    if (!currentPassword || !newPassword || !confirmPassword) {
      setError("Unesite sva tri polja.");
      return;
    }
    if (newPassword.length < MIN_PASSWORD_LENGTH) {
      setError(`Nova lozinka mora imati najmanje ${MIN_PASSWORD_LENGTH} karaktera.`);
      return;
    }
    if (newPassword !== confirmPassword) {
      setError("Nove lozinke se ne podudaraju.");
      return;
    }
    if (newPassword === currentPassword) {
      setError("Nova lozinka mora biti drugačija od trenutne.");
      return;
    }

    setSubmitting(true);
    try {
      const response = await changePassword({
        currentPassword,
        newPassword,
      });
      setSuccess(
        response.message?.trim() ||
          "Lozinka je uspješno promijenjena. Prijavite se ponovo.",
      );
      window.setTimeout(() => {
        logout();
        router.replace("/login");
      }, 700);
    } catch (err) {
      if (err instanceof ApiError && err.message.trim()) {
        setError(err.message);
      } else {
        setError("Lozinku trenutno nije moguće promijeniti.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={(event) => void onSubmit(event)} className="space-y-4">
      <PasswordField
        id="settings-current-password"
        label="Trenutna lozinka"
        value={currentPassword}
        disabled={submitting}
        autoComplete="current-password"
        onChange={setCurrentPassword}
      />
      <PasswordField
        id="settings-new-password"
        label="Nova lozinka"
        value={newPassword}
        disabled={submitting}
        autoComplete="new-password"
        placeholder={`Najmanje ${MIN_PASSWORD_LENGTH} karaktera`}
        onChange={setNewPassword}
      />
      <PasswordField
        id="settings-confirm-password"
        label="Potvrdite novu lozinku"
        value={confirmPassword}
        disabled={submitting}
        autoComplete="new-password"
        onChange={setConfirmPassword}
      />

      {error ? (
        <p className="break-words rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
          {error}
        </p>
      ) : null}
      {success ? (
        <p className="break-words rounded-xl border border-emerald-100 bg-emerald-50 px-3 py-2 text-sm text-emerald-800">
          {success}
        </p>
      ) : null}

      <button
        type="submit"
        disabled={submitting}
        className="inline-flex min-h-11 items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-medium text-white transition hover:bg-stone-800 disabled:cursor-not-allowed disabled:opacity-60"
      >
        {submitting ? "Promjena..." : "Promijeni lozinku"}
      </button>
    </form>
  );
}
