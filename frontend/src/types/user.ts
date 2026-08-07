import { roleLabel, UserRole } from "@/types/auth";

/** Role koje se smiju dodijeliti kroz Users UI (nikad developer). */
export type AssignableEmployeeRole = Exclude<UserRole, "developer">;

export interface Employee {
  id: number;
  username: string;
  role: UserRole;
  fullName: string;
  isActive: boolean;
}

export interface CreateEmployeePayload {
  username: string;
  password: string;
  role: AssignableEmployeeRole;
  fullName?: string;
}

export interface UpdateEmployeePayload {
  username: string;
  role: AssignableEmployeeRole;
  fullName?: string;
}

export interface UpdateEmployeePasswordPayload {
  password: string;
}

export interface UpdateEmployeeStatusPayload {
  isActive: boolean;
}

export interface EmployeesListResponse {
  data: Employee[];
}

export interface EmployeeMutationResponse {
  message?: string;
  data: Employee;
}

export interface EmployeePasswordResponse {
  message?: string;
}

export const ASSIGNABLE_EMPLOYEE_ROLES: AssignableEmployeeRole[] = [
  "sef",
  "menadzer",
  "radnik",
];

export const MIN_PASSWORD_LENGTH = 8;

export function isAssignableEmployeeRole(
  value: string,
): value is AssignableEmployeeRole {
  return (
    value === "sef" || value === "menadzer" || value === "radnik"
  );
}

export function employeeRoleLabel(role: string): string {
  if (isAssignableEmployeeRole(role) || role === "developer") {
    return roleLabel(role);
  }
  return role;
}

export function employeeDisplayName(employee: Employee): string {
  const name = employee.fullName?.trim();
  if (name) {
    return name;
  }
  return employee.username;
}
