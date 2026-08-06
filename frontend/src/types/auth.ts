export type UserRole = "sef" | "menadzer" | "radnik";

export interface AuthUser {
  id: number;
  username: string;
  role: UserRole;
  isActive?: boolean;
}

export interface LoginResponse {
  token: string;
  user: AuthUser;
}

export function isUserRole(value: string): value is UserRole {
  return value === "sef" || value === "menadzer" || value === "radnik";
}

export function roleLabel(role: UserRole): string {
  switch (role) {
    case "sef":
      return "Šef";
    case "menadzer":
      return "Menadžer";
    case "radnik":
      return "Radnik";
  }
}
