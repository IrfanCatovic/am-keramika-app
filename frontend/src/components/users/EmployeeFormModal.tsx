"use client";

import { FormEvent, useState } from "react";

import { PasswordField } from "@/components/users/PasswordField";
import { Modal } from "@/components/ui/Modal";
import {
  ASSIGNABLE_EMPLOYEE_ROLES,
  AssignableEmployeeRole,
  Employee,
  employeeRoleLabel,
  isAssignableEmployeeRole,
  MIN_PASSWORD_LENGTH,
} from "@/types/user";

type Mode = "create" | "edit";

function initialRole(employee?: Employee | null): AssignableEmployeeRole {
  if (employee && isAssignableEmployeeRole(employee.role)) {
    return employee.role;
  }
  return "radnik";
}

export function EmployeeFormModal({
  open,
  mode,
  employee,
  loading,
  error,
  onClose,
  onSubmit,
}: {
  open: boolean;
  mode: Mode;
  employee?: Employee | null;
  loading: boolean;
  error: string | null;
  onClose: () => void;
  onSubmit: (values: {
    username: string;
    fullName: string;
    role: AssignableEmployeeRole;
    password?: string;
  }) => Promise<void> | void;
}) {
  const [username, setUsername] = useState(
    mode === "edit" ? (employee?.username ?? "") : "",
  );
  const [fullName, setFullName] = useState(
    mode === "edit" ? (employee?.fullName ?? "") : "",
  );
  const [role, setRole] = useState<AssignableEmployeeRole>(
    initialRole(mode === "edit" ? employee : null),
  );
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [localError, setLocalError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setLocalError(null);

    const trimmedUsername = username.trim();
    if (!trimmedUsername) {
      setLocalError("Unesite korisničko ime.");
      return;
    }
    const trimmedFullName = fullName.trim();
    if (trimmedFullName.length < 2) {
      setLocalError("Unesite ime i prezime zaposlenog.");
      return;
    }
    if (!isAssignableEmployeeRole(role)) {
      setLocalError("Izaberite validnu ulogu.");
      return;
    }

    if (mode === "create") {
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
    }

    await onSubmit({
      username: trimmedUsername,
      fullName: trimmedFullName,
      role,
      password: mode === "create" ? password : undefined,
    });
  }

  return (
    <Modal
      open={open}
      title={mode === "create" ? "Dodaj zaposlenog" : "Izmeni zaposlenog"}
      description={
        mode === "create"
          ? "Kreirajte nalog sa ulogom Šef, Menadžer ili Radnik."
          : "Ažurirajte podatke i ovlašćenja zaposlenog."
      }
      onClose={loading ? () => undefined : onClose}
    >
      <form className="space-y-3" onSubmit={(event) => void handleSubmit(event)}>
        <label className="block">
          <span className="mb-1.5 block text-sm font-medium text-stone-700">
            Ime i prezime
          </span>
          <input
            value={fullName}
            disabled={loading}
            onChange={(event) => setFullName(event.target.value)}
            className="min-h-11 w-full rounded-xl border border-stone-200 bg-white px-3 text-sm text-stone-900 outline-none focus:border-stone-400 disabled:opacity-60"
            placeholder="npr. Marko Marković"
            required
          />
        </label>

        <label className="block">
          <span className="mb-1.5 block text-sm font-medium text-stone-700">
            Korisničko ime
          </span>
          <input
            value={username}
            disabled={loading}
            autoComplete="username"
            onChange={(event) => setUsername(event.target.value)}
            className="min-h-11 w-full rounded-xl border border-stone-200 bg-white px-3 text-sm text-stone-900 outline-none focus:border-stone-400 disabled:opacity-60"
            placeholder="npr. marko"
            required
          />
        </label>

        <label className="block">
          <span className="mb-1.5 block text-sm font-medium text-stone-700">
            Uloga
          </span>
          <select
            value={role}
            disabled={loading}
            onChange={(event) => {
              const next = event.target.value;
              if (isAssignableEmployeeRole(next)) {
                setRole(next);
              }
            }}
            className="min-h-11 w-full rounded-xl border border-stone-200 bg-white px-3 text-sm text-stone-900 outline-none focus:border-stone-400 disabled:opacity-60"
          >
            {ASSIGNABLE_EMPLOYEE_ROLES.map((option) => (
              <option key={option} value={option}>
                {employeeRoleLabel(option)}
              </option>
            ))}
          </select>
        </label>

        {mode === "create" ? (
          <>
            <PasswordField
              id="employee-password"
              label="Lozinka"
              value={password}
              disabled={loading}
              onChange={setPassword}
              placeholder={`Najmanje ${MIN_PASSWORD_LENGTH} karaktera`}
            />
            <PasswordField
              id="employee-password-confirm"
              label="Potvrdi lozinku"
              value={confirmPassword}
              disabled={loading}
              onChange={setConfirmPassword}
            />
          </>
        ) : null}

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
            {loading
              ? "Sačekajte…"
              : mode === "create"
                ? "Kreiraj zaposlenog"
                : "Sačuvaj izmjene"}
          </button>
        </div>
      </form>
    </Modal>
  );
}
