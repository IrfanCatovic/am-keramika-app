"use client";

import { FormEvent, useState } from "react";

import { PasswordField } from "@/components/users/PasswordField";
import { Modal } from "@/components/ui/Modal";
import { Employee, MIN_PASSWORD_LENGTH } from "@/types/user";

export function EmployeePasswordModal({
  open,
  employee,
  loading,
  error,
  onClose,
  onSubmit,
}: {
  open: boolean;
  employee: Employee | null;
  loading: boolean;
  error: string | null;
  onClose: () => void;
  onSubmit: (password: string) => Promise<void> | void;
}) {
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [localError, setLocalError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setLocalError(null);
    if (password.length < MIN_PASSWORD_LENGTH) {
      setLocalError(
        `Lozinka mora imati najmanje ${MIN_PASSWORD_LENGTH} karaktera.`,
      );
      return;
    }
    if (password !== confirmPassword) {
      setLocalError("Lozinke se ne poklapaju.");
      return;
    }
    await onSubmit(password);
  }

  return (
    <Modal
      open={open}
      title="Promeni lozinku"
      description={
        employee
          ? `Postavite novu lozinku za nalog „${employee.username}”.`
          : undefined
      }
      onClose={loading ? () => undefined : onClose}
    >
      <form className="space-y-3" onSubmit={(event) => void handleSubmit(event)}>
        <PasswordField
          id="employee-new-password"
          label="Nova lozinka"
          value={password}
          disabled={loading}
          onChange={setPassword}
          placeholder={`Najmanje ${MIN_PASSWORD_LENGTH} karaktera`}
        />
        <PasswordField
          id="employee-new-password-confirm"
          label="Potvrdi novu lozinku"
          value={confirmPassword}
          disabled={loading}
          onChange={setConfirmPassword}
        />

        {localError || error ? (
          <p className="break-words rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-700">
            {localError ?? error}
          </p>
        ) : null}

        <div className="flex flex-col-reverse gap-2 pt-1 sm:flex-row sm:justify-end">
          <button
            type="button"
            disabled={loading}
            onClick={onClose}
            className="inline-flex min-h-11 items-center justify-center rounded-xl border border-stone-200 px-4 text-sm font-medium text-stone-700 hover:bg-stone-50 disabled:opacity-60"
          >
            Odustani
          </button>
          <button
            type="submit"
            disabled={loading}
            className="inline-flex min-h-11 items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white hover:bg-stone-800 disabled:opacity-60"
          >
            {loading ? "Sačekajte…" : "Sačuvaj lozinku"}
          </button>
        </div>
      </form>
    </Modal>
  );
}
