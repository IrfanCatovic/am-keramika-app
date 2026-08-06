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
    roles: ["sef", "menadzer", "radnik"],
    enabled: true,
  },
  {
    label: "Proizvodi",
    href: "/products",
    roles: ["sef", "menadzer", "radnik"],
    enabled: true,
  },
  {
    label: "Kategorije i grupe",
    href: "/categories",
    roles: ["sef", "menadzer", "radnik"],
    enabled: true,
  },
  {
    label: "Lager",
    href: "/inventory",
    roles: ["sef", "menadzer", "radnik"],
    enabled: true,
  },
  {
    label: "Kupci",
    href: "/customers",
    roles: ["sef", "menadzer", "radnik"],
    enabled: true,
  },
  {
    label: "Računi",
    href: "/invoices",
    roles: ["sef", "menadzer", "radnik"],
    enabled: true,
  },
  {
    label: "Uplate",
    href: "/payments",
    roles: ["sef", "menadzer", "radnik"],
    enabled: true,
  },
  {
    label: "Izvještaji",
    href: "/reports",
    roles: ["sef", "menadzer"],
    enabled: true,
  },
  {
    label: "Korisnici",
    href: "/users",
    roles: ["sef"],
    enabled: true,
  },
];

export function getNavItemsForRole(role: UserRole): NavItem[] {
  return NAV_ITEMS.filter((item) => item.roles.includes(role));
}
