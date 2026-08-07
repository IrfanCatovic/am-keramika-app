import { apiRequest } from "@/lib/api";
import { getApiBusinessMessage } from "@/lib/categories-api";
import {
  CreateEmployeePayload,
  Employee,
  EmployeeMutationResponse,
  EmployeePasswordResponse,
  EmployeesListResponse,
  UpdateEmployeePayload,
  UpdateEmployeePasswordPayload,
  UpdateEmployeeStatusPayload,
} from "@/types/user";

export { getApiBusinessMessage };

export async function fetchEmployees(): Promise<Employee[]> {
  const response = await apiRequest<EmployeesListResponse>("/users");
  return Array.isArray(response.data) ? response.data : [];
}

export async function createEmployee(
  payload: CreateEmployeePayload,
): Promise<Employee> {
  const response = await apiRequest<EmployeeMutationResponse>("/users", {
    method: "POST",
    body: {
      username: payload.username.trim(),
      password: payload.password,
      role: payload.role,
      fullName: payload.fullName?.trim() ?? "",
    },
  });
  return response.data;
}

export async function updateEmployee(
  id: number,
  payload: UpdateEmployeePayload,
): Promise<Employee> {
  const response = await apiRequest<EmployeeMutationResponse>(`/users/${id}`, {
    method: "PUT",
    body: {
      username: payload.username.trim(),
      role: payload.role,
      fullName: payload.fullName?.trim() ?? "",
    },
  });
  return response.data;
}

export async function updateEmployeePassword(
  id: number,
  payload: UpdateEmployeePasswordPayload,
): Promise<void> {
  await apiRequest<EmployeePasswordResponse>(`/users/${id}/password`, {
    method: "PUT",
    body: { password: payload.password },
  });
}

export async function updateEmployeeStatus(
  id: number,
  payload: UpdateEmployeeStatusPayload,
): Promise<Employee> {
  const response = await apiRequest<EmployeeMutationResponse>(
    `/users/${id}/status`,
    {
      method: "PUT",
      body: { isActive: payload.isActive },
    },
  );
  return response.data;
}
