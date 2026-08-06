import { UserRole } from "@/types/auth";

export interface NavItem {
  label: string;
  href: string;
  roles: UserRole[];
  enabled: boolean;
}

export const NAV_ITEMS: NavItem[] = [
  {
    label: "Dashboard",
    href: "/dashboard",
    roles: ["developer", "sef", "menadzer", "radnik"],
    enabled: true,
  },
  {
    label: "Proizvodi",
    href: "/products",
    roles: ["developer", "sef", "menadzer", "radnik"],
    enabled: true,
  },
  {
    label: "Kategorije i grupe",
    href: "/categories",
    roles: ["developer", "sef", "menadzer", "radnik"],
    enabled: true,
  },
  {
    label: "Lager",
    href: "/inventory",
    roles: ["developer", "sef", "menadzer", "radnik"],
    enabled: true,
  },
  {
    label: "Kupci",
    href: "/customers",
    roles: ["developer", "sef", "menadzer", "radnik"],
    enabled: true,
  },
  {
    label: "Računi",
    href: "/invoices",
    roles: ["developer", "sef", "menadzer", "radnik"],
    enabled: true,
  },
  {
    label: "Uplate",
    href: "/payments",
    roles: ["developer", "sef", "menadzer", "radnik"],
    enabled: true,
  },
  {
    label: "Izvještaji",
    href: "/reports",
    roles: ["developer", "sef", "menadzer"],
    enabled: true,
  },
  {
    label: "Korisnici",
    href: "/users",
    roles: ["developer", "sef"],
    enabled: true,
  },
];

export function getNavItemsForRole(role: UserRole): NavItem[] {
  return NAV_ITEMS.filter((item) => item.roles.includes(role));
}
