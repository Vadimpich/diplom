import type { RoleSlug } from "@/lib/api/types";

export function getRoleHome(role?: RoleSlug | string | null): "/admin" | "/operator" {
  return role === "admin" ? "/admin" : "/operator";
}
