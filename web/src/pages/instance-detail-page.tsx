import { useState } from "react";
import { Link, useParams } from "react-router";
import { ArrowLeft, Copy } from "lucide-react";
import { toast } from "sonner";
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
import { SenderIdentityCard } from "@/components/sender-identity-card";
import { PostmarkServerCard } from "@/components/postmark-server-card";
import { useOrgContext } from "@/lib/org-context";
import { copyToClipboard } from "@/lib/utils";
import { unwrap } from "@/api/unwrap";
import {
  useGetV1OrgsOrgIDInstancesInstanceID,
  useGetV1OrgsOrgIDInstancesInstanceIDEvents,
  usePostV1OrgsOrgIDInstancesInstanceIDSetupLink,
} from "@/api/generated/endpoints/instances/instances";
import type {
  InstanceResponse,
  ProvisioningJobListResponse,
  SetupLinkResponse,
} from "@/api/generated/model";

const JOB_STATUS_LABEL: Record<string, string> = {
  created: "Queued",
  succeeded: "Succeeded",
  failed: "Failed",
};

const JOB_STATUS_CLASS: Record<string, string> = {
  created: "text-muted-foreground",
  succeeded: "text-status-green",
  failed: "text-status-red",
};

function formatTimestamp(value?: string) {
  if (!value) return "—";
  return new Date(value).toLocaleString();
}

function LoadingSkeleton() {
  return (
    <div className="flex flex-col gap-3">
      <Skeleton className="h-8 w-64" />
      <Skeleton className="h-24 w-full" />
      <Skeleton className="h-40 w-full" />
    </div>
  );
}

export function InstanceDetailPage() {
  const { instanceId = "" } = useParams<{ instanceId: string }>();
  const { selectedOrg, isLoading: orgLoading } = useOrgContext();
  const orgId = selectedOrg?.id ?? "";
  const [resentSetupUrl, setResentSetupUrl] = useState<string>();

  const instanceQuery = useGetV1OrgsOrgIDInstancesInstanceID(orgId, instanceId, {
    query: { enabled: !!orgId && !!instanceId },
  });
  const eventsQuery = useGetV1OrgsOrgIDInstancesInstanceIDEvents(
    orgId,
    instanceId,
    { query: { enabled: !!orgId && !!instanceId } },
  );

  const resendSetupLink = usePostV1OrgsOrgIDInstancesInstanceIDSetupLink({
    mutation: {
      onSuccess: (result) => {
        const url = unwrap<SetupLinkResponse>(result).data?.setup_url;
        if (url) setResentSetupUrl(url);
        toast.success("Reissued the setup link");
      },
      onError: (error) => {
        toast.error(error.error ?? "Couldn't reissue the setup link");
      },
    },
  });

  if (orgLoading || instanceQuery.isLoading) {
    return <LoadingSkeleton />;
  }

  const instance = instanceQuery.data
    ? unwrap<InstanceResponse>(instanceQuery.data).data
    : undefined;

  if (!instance) {
    return (
      <div className="flex flex-col items-center gap-3 rounded-md border border-dashed border-border py-16 text-center">
        <h2 className="font-mono text-lg font-semibold">
          Instance not found
        </h2>
        <Link
          to="/"
          className="text-sm text-muted-foreground underline underline-offset-4"
        >
          Back to instances
        </Link>
      </div>
    );
  }

  const setupUrl = resentSetupUrl ?? instance.admin_setup_url;
  const events = eventsQuery.data
    ? (unwrap<ProvisioningJobListResponse>(eventsQuery.data).data ?? [])
    : [];

  return (
    <div className="flex flex-col gap-6">
      <div>
        <Link
          to="/"
          className="mb-3 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="size-3.5" /> Instances
        </Link>
        <div className="flex items-center gap-3">
          <h1 className="font-mono text-2xl font-bold tracking-tight">
            {instance.name}
          </h1>
          <InstanceStatusBadge status={instance.status} />
        </div>
        <p className="text-sm text-muted-foreground">{instance.slug}</p>
      </div>

      <div className="rounded-md border border-border p-4">
        <h2 className="mb-1 text-sm font-semibold">Admin setup link</h2>
        <p className="mb-3 text-xs text-muted-foreground">
          One-time link for {instance.admin_email} to set their password on
          this workspace. It's spent after first use, and lost if the
          instance restarts before then — reissue a fresh one below if
          needed.
        </p>
        {setupUrl ? (
          <div className="flex items-center gap-2">
            <code className="flex-1 truncate rounded-md bg-secondary px-2 py-1.5 text-xs">
              {setupUrl}
            </code>
            <Button
              variant="outline"
              size="icon"
              onClick={() => copyToClipboard(setupUrl, "Setup link copied")}
              aria-label="Copy setup link"
            >
              <Copy className="size-3.5" />
            </Button>
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">
            No setup link yet — the workspace is still being provisioned.
          </p>
        )}
        <Button
          variant="ghost"
          size="sm"
          className="mt-3"
          disabled={
            resendSetupLink.isPending || instance.status !== "active"
          }
          onClick={() =>
            resendSetupLink.mutate({ orgID: orgId, instanceID: instanceId })
          }
        >
          {resendSetupLink.isPending ? "Reissuing…" : "Resend setup link"}
        </Button>
      </div>

      <SenderIdentityCard orgId={orgId} instanceId={instanceId} />

      <PostmarkServerCard orgId={orgId} instanceId={instanceId} />

      <div>
        <h2 className="mb-2 text-sm font-semibold">Provisioning timeline</h2>
        {eventsQuery.isLoading ? (
          <Skeleton className="h-24 w-full" />
        ) : events.length === 0 ? (
          <p className="text-xs text-muted-foreground">No events yet.</p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Step</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Attempts</TableHead>
                <TableHead>Updated</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {events.map((event) => (
                <TableRow key={event.id}>
                  <TableCell className="font-mono text-xs">
                    {event.job_type}
                  </TableCell>
                  <TableCell>
                    <span
                      className={`text-xs font-medium ${
                        JOB_STATUS_CLASS[event.status ?? ""] ??
                        "text-muted-foreground"
                      }`}
                    >
                      {JOB_STATUS_LABEL[event.status ?? ""] ?? event.status}
                    </span>
                    {event.last_error && (
                      <p className="mt-0.5 text-xs text-status-red">
                        {event.last_error}
                      </p>
                    )}
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {event.attempts}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {formatTimestamp(event.updated_at)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </div>
    </div>
  );
}
