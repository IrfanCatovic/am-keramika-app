export type UserRole = "developer" | "sef" | "menadzer" | "radnik";

export interface AuthUser {
  id: number;
  username: string;
  role: UserRole;
  fullName?: string;
  isActive?: boolean;
}

export interface LoginResponse {
  token: string;
  user: AuthUser;
}

export function isUserRole(value: string): value is UserRole {
  return (
    value === "developer" ||
    value === "sef" ||
    value === "menadzer" ||
    value === "radnik"
  );
}

export function roleLabel(role: UserRole): string {
  switch (role) {
    case "developer":
      return "Developer";
    case "sef":
      return "Šef";
    case "menadzer":
      return "Menadžer";
    case "radnik":
      return "Radnik";
  }
}
