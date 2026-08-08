import { apiRequest } from "@/lib/api";

export interface ChangePasswordPayload {
  currentPassword: string;
  newPassword: string;
}

export interface ChangePasswordResponse {
  message?: string;
}

export async function changePassword(
  payload: ChangePasswordPayload,
): Promise<ChangePasswordResponse> {
  return apiRequest<ChangePasswordResponse>("/auth/change-password", {
    method: "PUT",
    body: {
      currentPassword: payload.currentPassword,
      newPassword: payload.newPassword,
    },
  });
}
