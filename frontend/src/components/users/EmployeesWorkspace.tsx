"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

import { useAuth } from "@/components/auth/AuthProvider";
import { ConfirmDialog } from "@/components/ui/ConfirmDialog";
import {
  EmptyState,
  InlineError,
  ListSkeleton,
} from "@/components/ui/EmptyState";
import { EmployeeFormModal } from "@/components/users/EmployeeFormModal";
import { EmployeePasswordModal } from "@/components/users/EmployeePasswordModal";
import { EmployeeStatusBadge } from "@/components/users/EmployeeStatusBadge";
import {
  createEmployee,
  fetchEmployees,
  getApiBusinessMessage,
  updateEmployee,
  updateEmployeePassword,
  updateEmployeeStatus,
} from "@/lib/users-api";
import {
  AssignableEmployeeRole,
  Employee,
  employeeDisplayName,
  employeeRoleLabel,
  isAssignableEmployeeRole,
} from "@/types/user";

type FormMode = "create" | "edit" | null;

export function EmployeesWorkspace() {
  const { user } = useAuth();
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const [formMode, setFormMode] = useState<FormMode>(null);
  const [editing, setEditing] = useState<Employee | null>(null);
  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const [passwordTarget, setPasswordTarget] = useState<Employee | null>(null);
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [passwordError, setPasswordError] = useState<string | null>(null);

  const [statusTarget, setStatusTarget] = useState<Employee | null>(null);
  const [statusLoading, setStatusLoading] = useState(false);
  const [statusError, setStatusError] = useState<string | null>(null);

  const loadEmployees = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchEmployees();
      // Extra client guard — backend already excludes developer.
      setEmployees(data.filter((item) => item.role !== "developer"));
    } catch (err) {
      setError(
        getApiBusinessMessage(err, "Učitavanje zaposlenih nije uspjelo."),
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadEmployees();
    }, 0);
    return () => window.clearTimeout(timer);
  }, [loadEmployees]);

  useEffect(() => {
    if (!successMessage) {
      return;
    }
    const timer = window.setTimeout(() => setSuccessMessage(null), 4000);
    return () => window.clearTimeout(timer);
  }, [successMessage]);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) {
      return employees;
    }
    return employees.filter((employee) => {
      const haystack = [
        employee.username,
        employee.fullName,
        employeeRoleLabel(employee.role),
      ]
        .join(" ")
        .toLowerCase();
      return haystack.includes(q);
    });
  }, [employees, search]);

  function openCreate() {
    setFormMode("create");
    setEditing(null);
    setFormError(null);
  }

  function openEdit(employee: Employee) {
    if (!isAssignableEmployeeRole(employee.role)) {
      return;
    }
    setFormMode("edit");
    setEditing(employee);
    setFormError(null);
  }

  function closeForm() {
    if (formLoading) {
      return;
    }
    setFormMode(null);
    setEditing(null);
    setFormError(null);
  }

  async function handleFormSubmit(values: {
    username: string;
    fullName: string;
    role: AssignableEmployeeRole;
    password?: string;
  }) {
    setFormLoading(true);
    setFormError(null);
    try {
      if (formMode === "create") {
        if (!values.password) {
          setFormError("Lozinka je obavezna.");
          setFormLoading(false);
          return;
        }
        await createEmployee({
          username: values.username,
          password: values.password,
          role: values.role,
          fullName: values.fullName,
        });
        setSuccessMessage("Zaposleni je uspešno kreiran.");
      } else if (formMode === "edit" && editing) {
        await updateEmployee(editing.id, {
          username: values.username,
          role: values.role,
          fullName: values.fullName,
        });
        setSuccessMessage("Podaci zaposlenog su sačuvani.");
      }
      setFormMode(null);
      setEditing(null);
      await loadEmployees();
    } catch (err) {
      setFormError(
        getApiBusinessMessage(err, "Čuvanje zaposlenog nije uspjelo."),
      );
    } finally {
      setFormLoading(false);
    }
  }

  async function handlePasswordSubmit(password: string) {
    if (!passwordTarget) {
      return;
    }
    setPasswordLoading(true);
    setPasswordError(null);
    try {
      await updateEmployeePassword(passwordTarget.id, { password });
      setPasswordTarget(null);
      setSuccessMessage("Lozinka je uspešno promenjena.");
    } catch (err) {
      setPasswordError(
        getApiBusinessMessage(err, "Promjena lozinke nije uspjela."),
      );
    } finally {
      setPasswordLoading(false);
    }
  }

  async function handleStatusConfirm() {
    if (!statusTarget) {
      return;
    }
    setStatusLoading(true);
    setStatusError(null);
    try {
      const nextActive = !statusTarget.isActive;
      await updateEmployeeStatus(statusTarget.id, { isActive: nextActive });
      setStatusTarget(null);
      setSuccessMessage(
        nextActive
          ? "Zaposleni je aktiviran."
          : "Zaposleni je deaktiviran.",
      );
      await loadEmployees();
    } catch (err) {
      setStatusError(
        getApiBusinessMessage(err, "Izmjena statusa nije uspjela."),
      );
    } finally {
      setStatusLoading(false);
    }
  }

  function canManage(employee: Employee): boolean {
    if (employee.role === "developer") {
      return false;
    }
    return isAssignableEmployeeRole(employee.role);
  }

  function canDeactivate(employee: Employee): boolean {
    if (!canManage(employee)) {
      return false;
    }
    if (!employee.isActive) {
      return true;
    }
    // Hide self-deactivate — backend also blocks it.
    if (user && employee.id === user.id) {
      return false;
    }
    return true;
  }

  const actionBtn =
    "inline-flex min-h-10 items-center justify-center rounded-xl border border-stone-200 bg-white px-3 text-sm font-medium text-stone-700 transition hover:bg-stone-50 disabled:cursor-not-allowed disabled:opacity-50";

  return (
    <div className="min-w-0 space-y-4">
      <header className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div className="min-w-0">
          <h1 className="text-2xl font-semibold tracking-tight text-stone-900">
            Zaposleni
          </h1>
          <p className="mt-1 text-sm text-stone-500">
            Upravljanje nalozima zaposlenih i njihovim ovlašćenjima.
          </p>
        </div>
        <button
          type="button"
          onClick={openCreate}
          className="inline-flex min-h-11 shrink-0 items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white transition hover:bg-stone-800"
        >
          Dodaj zaposlenog
        </button>
      </header>

      {successMessage ? (
        <div className="rounded-xl border border-emerald-100 bg-emerald-50 px-4 py-3 text-sm text-emerald-900">
          {successMessage}
        </div>
      ) : null}

      <div className="rounded-2xl border border-stone-200 bg-white p-3 sm:p-4">
        <label className="block max-w-md">
          <span className="mb-1.5 block text-sm font-medium text-stone-700">
            Pretraga
          </span>
          <input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Ime, korisničko ime ili uloga…"
            className="min-h-11 w-full rounded-xl border border-stone-200 bg-white px-3 text-sm text-stone-900 outline-none focus:border-stone-400"
          />
        </label>
      </div>

      {loading ? <ListSkeleton rows={4} /> : null}

      {!loading && error ? (
        <InlineError message={error} onRetry={() => void loadEmployees()} />
      ) : null}

      {!loading && !error && employees.length === 0 ? (
        <EmptyState
          title="Još nema kreiranih zaposlenih."
          description="Dodajte prvog zaposlenog da bi mogao da se prijavi u sistem."
          action={
            <button
              type="button"
              onClick={openCreate}
              className="inline-flex min-h-11 items-center justify-center rounded-xl bg-stone-900 px-4 text-sm font-semibold text-white hover:bg-stone-800"
            >
              Dodaj prvog zaposlenog
            </button>
          }
        />
      ) : null}

      {!loading && !error && employees.length > 0 && filtered.length === 0 ? (
        <EmptyState
          title="Nema rezultata za pretragu."
          description="Pokušajte drugim pojmom ili očistite pretragu."
        />
      ) : null}

      {!loading && !error && filtered.length > 0 ? (
        <>
          {/* Desktop table */}
          <div className="hidden overflow-hidden rounded-2xl border border-stone-200 bg-white md:block">
            <table className="min-w-full text-left text-sm">
              <thead className="border-b border-stone-100 bg-stone-50/80 text-xs font-medium uppercase tracking-wide text-stone-500">
                <tr>
                  <th className="px-4 py-3">Zaposleni</th>
                  <th className="px-4 py-3">Uloga</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3 text-right">Akcije</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-stone-100">
                {filtered.map((employee) => (
                  <tr key={employee.id} className="align-middle">
                    <td className="px-4 py-3">
                      <p className="font-medium text-stone-900">
                        {employeeDisplayName(employee)}
                      </p>
                      <p className="text-xs text-stone-500">
                        @{employee.username}
                      </p>
                    </td>
                    <td className="px-4 py-3 text-stone-700">
                      {employeeRoleLabel(employee.role)}
                    </td>
                    <td className="px-4 py-3">
                      <EmployeeStatusBadge isActive={employee.isActive} />
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap justify-end gap-2">
                        {canManage(employee) ? (
                          <>
                            <button
                              type="button"
                              className={actionBtn}
                              onClick={() => openEdit(employee)}
                            >
                              Izmeni
                            </button>
                            <button
                              type="button"
                              className={actionBtn}
                              onClick={() => {
                                setPasswordError(null);
                                setPasswordTarget(employee);
                              }}
                            >
                              Promeni lozinku
                            </button>
                            {canDeactivate(employee) ? (
                              <button
                                type="button"
                                className={actionBtn}
                                onClick={() => {
                                  setStatusError(null);
                                  setStatusTarget(employee);
                                }}
                              >
                                {employee.isActive
                                  ? "Deaktiviraj"
                                  : "Aktiviraj"}
                              </button>
                            ) : null}
                          </>
                        ) : null}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Mobile cards */}
          <ul className="space-y-3 md:hidden">
            {filtered.map((employee) => (
              <li
                key={employee.id}
                className="rounded-2xl border border-stone-200 bg-white p-4"
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="truncate text-base font-semibold text-stone-900">
                      {employeeDisplayName(employee)}
                    </p>
                    <p className="truncate text-sm text-stone-500">
                      @{employee.username}
                    </p>
                    <p className="mt-1 text-sm text-stone-700">
                      {employeeRoleLabel(employee.role)}
                    </p>
                  </div>
                  <EmployeeStatusBadge isActive={employee.isActive} />
                </div>
                {canManage(employee) ? (
                  <div className="mt-3 grid grid-cols-1 gap-2 min-[400px]:grid-cols-2">
                    <button
                      type="button"
                      className={actionBtn}
                      onClick={() => openEdit(employee)}
                    >
                      Izmeni
                    </button>
                    <button
                      type="button"
                      className={actionBtn}
                      onClick={() => {
                        setPasswordError(null);
                        setPasswordTarget(employee);
                      }}
                    >
                      Promeni lozinku
                    </button>
                    {canDeactivate(employee) ? (
                      <button
                        type="button"
                        className={`${actionBtn} min-[400px]:col-span-2`}
                        onClick={() => {
                          setStatusError(null);
                          setStatusTarget(employee);
                        }}
                      >
                        {employee.isActive ? "Deaktiviraj" : "Aktiviraj"}
                      </button>
                    ) : null}
                  </div>
                ) : null}
              </li>
            ))}
          </ul>
        </>
      ) : null}

      <EmployeeFormModal
        key={
          formMode === "edit"
            ? `edit-${editing?.id ?? "none"}`
            : formMode === "create"
              ? "create"
              : "closed"
        }
        open={formMode != null}
        mode={formMode === "edit" ? "edit" : "create"}
        employee={editing}
        loading={formLoading}
        error={formError}
        onClose={closeForm}
        onSubmit={handleFormSubmit}
      />

      <EmployeePasswordModal
        key={passwordTarget ? `pw-${passwordTarget.id}` : "pw-closed"}
        open={passwordTarget != null}
        employee={passwordTarget}
        loading={passwordLoading}
        error={passwordError}
        onClose={() => {
          if (!passwordLoading) {
            setPasswordTarget(null);
            setPasswordError(null);
          }
        }}
        onSubmit={handlePasswordSubmit}
      />

      <ConfirmDialog
        open={statusTarget != null && statusTarget.isActive}
        title="Deaktivacija zaposlenog"
        message={
          statusTarget
            ? `Da li ste sigurni da želite da deaktivirate nalog zaposlenog ${statusTarget.username}? Zaposleni više neće moći da se prijavi dok nalog ponovo ne aktivirate.`
            : ""
        }
        confirmLabel="Deaktiviraj"
        cancelLabel="Odustani"
        loading={statusLoading}
        error={statusError}
        tone="danger"
        onConfirm={() => void handleStatusConfirm()}
        onClose={() => {
          if (!statusLoading) {
            setStatusTarget(null);
            setStatusError(null);
          }
        }}
      />

      <ConfirmDialog
        open={statusTarget != null && !statusTarget.isActive}
        title="Aktivacija zaposlenog"
        message={
          statusTarget
            ? `Aktivirati nalog zaposlenog ${statusTarget.username}?`
            : ""
        }
        confirmLabel="Aktiviraj"
        cancelLabel="Odustani"
        loading={statusLoading}
        error={statusError}
        tone="neutral"
        onConfirm={() => void handleStatusConfirm()}
        onClose={() => {
          if (!statusLoading) {
            setStatusTarget(null);
            setStatusError(null);
          }
        }}
      />
    </div>
  );
}
