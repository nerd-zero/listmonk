import { useState } from "react";
import { useNavigate } from "react-router";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { InstanceStatusBadge } from "@/components/instance-status-badge";
import { CreateInstanceWizard } from "@/components/create-instance-wizard";
import { useOrgContext } from "@/lib/org-context";
import { unwrap } from "@/api/unwrap";
import { useGetV1OrgsOrgIDInstances } from "@/api/generated/endpoints/instances/instances";
import type { InstanceListResponse } from "@/api/generated/model";

function LoadingSkeleton() {
  return (
    <div className="flex flex-col gap-3">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-16 w-full" />
      <Skeleton className="h-16 w-full" />
    </div>
  );
}

export function InstancesPage() {
  const navigate = useNavigate();
  const { selectedOrg, isLoading: orgLoading } = useOrgContext();
  const [createOpen, setCreateOpen] = useState(false);

  const instancesQuery = useGetV1OrgsOrgIDInstances(selectedOrg?.id ?? "", {
    query: { enabled: !!selectedOrg?.id },
  });

  if (orgLoading) {
    return <LoadingSkeleton />;
  }

  // No org selected (none exist yet, or none loaded) -- never show the
  // instances list/empty-state/create-instance flow without one, since
  // every one of those is meaningless without an org to scope them to.
  if (!selectedOrg) {
    return (
      <div className="flex flex-col items-center gap-3 rounded-md border border-dashed border-border py-16 text-center">
        <h2 className="font-mono text-lg font-semibold">No org yet</h2>
        <p className="max-w-sm text-sm text-muted-foreground">
          Create an org from the sidebar to start setting up workspaces.
        </p>
      </div>
    );
  }

  if (instancesQuery.isLoading) {
    return <LoadingSkeleton />;
  }

  const instances = instancesQuery.data
    ? (unwrap<InstanceListResponse>(instancesQuery.data).data ?? [])
    : [];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-baseline justify-between">
        <div>
          <h1 className="font-mono text-2xl font-bold tracking-tight">
            Instances
          </h1>
          <p className="text-sm text-muted-foreground">
            Every workspace <strong>{selectedOrg?.name}</strong> owns.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>New instance</Button>
      </div>

      {instances.length === 0 ? (
        <EmptyState onCreate={() => setCreateOpen(true)} orgName={selectedOrg?.name} />
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Workspace</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {instances.map((instance) => (
              <TableRow
                key={instance.id}
                className="cursor-pointer"
                onClick={() => navigate(`/instances/${instance.id}`)}
              >
                <TableCell>
                  <div className="font-mono text-sm font-medium">
                    {instance.slug}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {instance.name}
                  </div>
                </TableCell>
                <TableCell>
                  <InstanceStatusBadge status={instance.status} />
                </TableCell>
                <TableCell className="font-mono text-xs text-muted-foreground">
                  {instance.created_at
                    ? new Date(instance.created_at).toLocaleDateString()
                    : "—"}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <CreateInstanceWizard
        open={createOpen}
        onOpenChange={setCreateOpen}
        orgId={selectedOrg?.id}
        orgName={selectedOrg?.name}
      />
    </div>
  );
}

function EmptyState({
  onCreate,
  orgName,
}: {
  onCreate: () => void;
  orgName?: string;
}) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-md border border-dashed border-border py-16 text-center">
      <h2 className="font-mono text-lg font-semibold">
        Let's set up {orgName ?? "your"} first workspace
      </h2>
      <p className="max-w-sm text-sm text-muted-foreground">
        This is where {orgName ?? "your org"}'s sending domains show up.
        Create one to get started -- it's usually ready in under a minute.
      </p>
      <Button onClick={onCreate}>Create instance</Button>
    </div>
  );
}
