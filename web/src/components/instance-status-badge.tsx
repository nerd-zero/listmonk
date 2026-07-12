import { Badge } from "@/components/ui/badge";

const STAGE_LABEL: Record<string, string> = {
  created: "Queued",
  listmonk_tenant_provisioning: "Creating your workspace",
  active: "Live",
  failed: "Needs attention",
};

const STAGE_CLASS: Record<string, string> = {
  created: "bg-secondary text-muted-foreground",
  listmonk_tenant_provisioning: "bg-status-amber-soft text-status-amber",
  active: "bg-status-green-soft text-status-green",
  failed: "bg-status-red-soft text-status-red",
};

export function InstanceStatusBadge({ status }: { status?: string }) {
  const key = status ?? "created";
  return (
    <Badge variant="outline" className={STAGE_CLASS[key] ?? STAGE_CLASS.created}>
      {STAGE_LABEL[key] ?? key}
    </Badge>
  );
}
