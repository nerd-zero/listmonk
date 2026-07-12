import { createContext, useContext, useEffect, useState } from "react";
import type { ReactNode } from "react";
import { useGetV1Orgs } from "@/api/generated/endpoints/orgs/orgs";
import type { DbOrg, OrgListResponse } from "@/api/generated/model";
import { unwrap } from "@/api/unwrap";

const SELECTED_ORG_STORAGE_KEY = "listnun.selectedOrgId";

// A person can belong to (and create) many orgs, each owning its own
// instances -- switching between them is a regular, self-service action.
// Cross-org visibility and management (suspend/reactivate/disable a
// tenant) is exclusively the super admin panel's job
// (internal/provisioning's AdminListOrgs/AdminSetTenantStatus) and never
// exposed here.
interface OrgContextValue {
  orgs: DbOrg[];
  isLoading: boolean;
  selectedOrgId: string | undefined;
  selectedOrg: DbOrg | undefined;
  setSelectedOrgId: (orgId: string) => void;
}

const OrgContext = createContext<OrgContextValue | undefined>(undefined);

export function OrgProvider({ children }: { children: ReactNode }) {
  const query = useGetV1Orgs();
  const orgs = query.data ? (unwrap<OrgListResponse>(query.data).data ?? []) : [];

  const [selectedOrgId, setSelectedOrgIdState] = useState<string | undefined>(
    () => localStorage.getItem(SELECTED_ORG_STORAGE_KEY) ?? undefined,
  );

  // Once orgs load, make sure the selection actually points at a real one
  // (first login, stale localStorage from a removed org, etc.) --
  // defaults to the first org.
  useEffect(() => {
    if (orgs.length === 0) return;
    if (selectedOrgId && orgs.some((org) => org.id === selectedOrgId)) return;
    setSelectedOrgIdState(orgs[0].id);
  }, [orgs, selectedOrgId]);

  const setSelectedOrgId = (orgId: string) => {
    localStorage.setItem(SELECTED_ORG_STORAGE_KEY, orgId);
    setSelectedOrgIdState(orgId);
  };

  const selectedOrg = orgs.find((org) => org.id === selectedOrgId);

  return (
    <OrgContext.Provider
      value={{
        orgs,
        isLoading: query.isLoading,
        selectedOrgId,
        selectedOrg,
        setSelectedOrgId,
      }}
    >
      {children}
    </OrgContext.Provider>
  );
}

export function useOrgContext() {
  const ctx = useContext(OrgContext);
  if (!ctx) {
    throw new Error("useOrgContext must be used within an OrgProvider");
  }
  return ctx;
}
