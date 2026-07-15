import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Copy } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { copyToClipboard } from "@/lib/utils";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { unwrap } from "@/api/unwrap";
import { ApiError } from "@/api/mutator";
import {
  getGetV1OrgsOrgIDInstancesInstanceIDCustomDomainQueryKey,
  useDeleteV1OrgsOrgIDInstancesInstanceIDCustomDomain,
  useGetV1OrgsOrgIDInstancesInstanceIDCustomDomain,
  usePostV1OrgsOrgIDInstancesInstanceIDCustomDomain,
} from "@/api/generated/endpoints/instances/instances";
import type { CustomDomainResponse } from "@/api/generated/model";

const RECORD_TYPE_LABEL: Record<string, string> = {
  custom_domain_cname: "CNAME",
  custom_domain_ownership: "Ownership verification (TXT)",
};

export function CustomDomainCard({
  orgId,
  instanceId,
}: {
  orgId: string;
  instanceId: string;
}) {
  const domainQuery = useGetV1OrgsOrgIDInstancesInstanceIDCustomDomain(
    orgId,
    instanceId,
    { query: { enabled: !!orgId && !!instanceId, retry: false } },
  );

  if (domainQuery.isLoading) {
    return <Skeleton className="h-32 w-full" />;
  }

  // Check the error before `data`: react-query keeps the last successful
  // `data` cached even after a later refetch errors (e.g. right after
  // removing the domain, which correctly 404s going forward) -- checking
  // `data` first would keep rendering the just-deleted domain forever.
  if (domainQuery.error instanceof ApiError && domainQuery.error.status === 404) {
    return <AddCustomDomainForm orgId={orgId} instanceId={instanceId} />;
  }

  if (domainQuery.error instanceof ApiError && domainQuery.error.status === 501) {
    return null;
  }

  if (domainQuery.data) {
    const detail = unwrap<CustomDomainResponse>(domainQuery.data).data;
    if (detail?.custom_domain) {
      return (
        <CustomDomainStatus
          orgId={orgId}
          instanceId={instanceId}
          detail={detail}
        />
      );
    }
  }

  return (
    <div className="rounded-md border border-border p-4">
      <h2 className="mb-1 text-sm font-semibold">Custom domain</h2>
      <p className="text-xs text-status-red">
        Couldn't load this workspace's custom domain.
      </p>
    </div>
  );
}

function CustomDomainStatus({
  orgId,
  instanceId,
  detail,
}: {
  orgId: string;
  instanceId: string;
  detail: NonNullable<CustomDomainResponse["data"]>;
}) {
  const [confirmOpen, setConfirmOpen] = useState(false);
  const queryClient = useQueryClient();

  const deleteDomain = useDeleteV1OrgsOrgIDInstancesInstanceIDCustomDomain({
    mutation: {
      onSuccess: () => {
        const queryKey = getGetV1OrgsOrgIDInstancesInstanceIDCustomDomainQueryKey(
          orgId,
          instanceId,
        );
        // Clear the cached domain immediately -- react-query wouldn't
        // otherwise drop it until the invalidated refetch below resolves,
        // during which this component would keep rendering the
        // just-deleted domain from stale cached data.
        queryClient.setQueryData(queryKey, undefined);
        queryClient.invalidateQueries({ queryKey });
        toast.success("Custom domain removed");
        setConfirmOpen(false);
      },
      onError: (error) => {
        toast.error(error.error ?? "Couldn't remove that custom domain");
      },
    },
  });

  const domain = detail.custom_domain;
  if (!domain) return null;
  const active = domain.status === "active";

  return (
    <div className="rounded-md border border-border p-4">
      <div className="mb-3 flex items-center justify-between">
        <div>
          <h2 className="text-sm font-semibold">Custom domain</h2>
          <p className="font-mono text-xs text-muted-foreground">
            {domain.domain}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Badge
            variant="outline"
            className={
              active
                ? "bg-status-green-soft text-status-green"
                : domain.status === "failed"
                  ? "bg-status-red-soft text-status-red"
                  : "bg-status-amber-soft text-status-amber"
            }
          >
            {active ? "Active" : domain.status === "failed" ? "Failed" : "Pending"}
          </Badge>
          <Button
            variant="ghost"
            size="sm"
            className="text-destructive hover:text-destructive"
            onClick={() => setConfirmOpen(true)}
          >
            Remove
          </Button>
        </div>
      </div>

      {active ? (
        <p className="text-xs text-muted-foreground">
          This workspace is reachable at {domain.domain} — its assigned{" "}
          subdomain keeps working too.
        </p>
      ) : detail.dns_records && detail.dns_records.length > 0 ? (
        <div className="flex flex-col gap-3">
          <p className="text-xs text-muted-foreground">
            Publish these records with your DNS provider — any provider
            works, not just Cloudflare:
          </p>
          {detail.dns_records.map((record) => (
            <div
              key={record.id}
              className="rounded-md bg-secondary px-2.5 py-2 text-xs"
            >
              <div className="mb-1.5 font-medium text-muted-foreground">
                {RECORD_TYPE_LABEL[record.record_type ?? ""] ??
                  record.record_type?.toUpperCase()}
              </div>
              <DNSField label="Host" value={record.host} />
              <DNSField label="Value" value={record.value} />
            </div>
          ))}
        </div>
      ) : (
        <p className="text-xs text-muted-foreground">
          No DNS records were returned for this domain.
        </p>
      )}

      <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove {domain.domain}?</AlertDialogTitle>
            <AlertDialogDescription>
              This reverts the workspace back to its assigned subdomain and
              removes the Cloudflare configuration for this domain. This
              cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteDomain.isPending}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              disabled={deleteDomain.isPending}
              onClick={() =>
                deleteDomain.mutate({ orgID: orgId, instanceID: instanceId })
              }
            >
              {deleteDomain.isPending ? "Removing…" : "Remove"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function AddCustomDomainForm({
  orgId,
  instanceId,
}: {
  orgId: string;
  instanceId: string;
}) {
  const [domain, setDomain] = useState("");
  const queryClient = useQueryClient();

  const addDomain = usePostV1OrgsOrgIDInstancesInstanceIDCustomDomain({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: getGetV1OrgsOrgIDInstancesInstanceIDCustomDomainQueryKey(
            orgId,
            instanceId,
          ),
        });
        toast.success("Custom domain added — publish the records below to activate it");
      },
      onError: (error) => {
        toast.error(error.error ?? "Couldn't add that custom domain");
      },
    },
  });

  return (
    <div className="flex flex-col gap-3 rounded-md border border-border p-4">
      <div>
        <h2 className="text-sm font-semibold">Custom domain</h2>
        <p className="text-xs text-muted-foreground">
          Reach this workspace at a domain you own, instead of only its
          assigned subdomain. Your DNS doesn't need to be on Cloudflare —
          any provider works.
        </p>
      </div>

      <div className="flex flex-col gap-2">
        <Label htmlFor="custom-domain">Domain</Label>
        <Input
          id="custom-domain"
          placeholder="mail.acme.com"
          value={domain}
          onChange={(e) => setDomain(e.target.value)}
        />
        <p className="text-xs text-muted-foreground">
          You'll get a CNAME and a verification record to publish before
          this domain goes live.
        </p>
      </div>

      <Button
        className="self-start"
        disabled={domain.trim().length === 0 || addDomain.isPending}
        onClick={() =>
          addDomain.mutate({
            orgID: orgId,
            instanceID: instanceId,
            data: { domain },
          })
        }
      >
        {addDomain.isPending ? "Adding…" : "Add"}
      </Button>
    </div>
  );
}

function DNSField({ label, value }: { label: string; value?: string }) {
  if (!value) return null;
  return (
    <div className="mb-1 flex items-center gap-2 last:mb-0">
      <span className="w-10 shrink-0 text-muted-foreground">{label}</span>
      <code className="min-w-0 flex-1 truncate font-mono">{value}</code>
      <Button
        variant="ghost"
        size="icon"
        className="size-5 shrink-0"
        onClick={() => copyToClipboard(value, `${label} copied`)}
        aria-label={`Copy ${label.toLowerCase()}`}
      >
        <Copy className="size-3" />
      </Button>
    </div>
  );
}
