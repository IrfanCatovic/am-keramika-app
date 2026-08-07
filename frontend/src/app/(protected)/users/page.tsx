"use client";

import { RequireRoles } from "@/components/auth/RequireRoles";
import { EmployeesWorkspace } from "@/components/users/EmployeesWorkspace";

export default function UsersPage() {
  return (
    <RequireRoles roles={["developer", "sef"]}>
      <EmployeesWorkspace />
    </RequireRoles>
  );
}
