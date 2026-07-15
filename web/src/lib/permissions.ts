import { useGetV1Me } from "@/api/generated/endpoints/users/users";
import { unwrap } from "@/api/unwrap";
import type { UserResponse } from "@/api/generated/model";
import { useOrgContext } from "@/lib/org-context";

// Single source of truth for what a role or is_super_admin unlocks --
// check here before adding another `role === "owner"` or `is_super_admin`
// check anywhere else in the app. Mirrors the backend's own gates
// (internal/provisioning.RequireOrgOwner/RequireSuperAdmin) so the UI
// hides what the API would 403 on anyway.
export type Ability = "inviteMember" | "viewAdminPanel";

const OWNER_ABILITIES: Ability[] = ["inviteMember"];
const SUPER_ADMIN_ABILITIES: Ability[] = ["viewAdminPanel"];

// usePermissions is the one place the app asks "can I do X" -- backed by
// react-query's cache (a single shared fetch of GET /v1/me, not
// re-requested per call site) plus the currently selected org's role from
// OrgContext.
export function usePermissions() {
  const meQuery = useGetV1Me();
  const { selectedOrg } = useOrgContext();

  const me = meQuery.data ? unwrap<UserResponse>(meQuery.data).data : undefined;
  const isSuperAdmin = me?.is_super_admin ?? false;
  const isOwner = selectedOrg?.role === "owner";

  function can(ability: Ability): boolean {
    if (isSuperAdmin && SUPER_ADMIN_ABILITIES.includes(ability)) return true;
    if (isOwner && OWNER_ABILITIES.includes(ability)) return true;
    return false;
  }

  return { isSuperAdmin, isOwner, isLoading: meQuery.isLoading, can };
}
