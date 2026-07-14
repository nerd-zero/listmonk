import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { unwrap } from "@/api/unwrap";
import { ApiError } from "@/api/mutator";
import {
  getGetV1OrgsOrgIDInstancesInstanceIDSenderIdentityQueryKey,
  useGetV1OrgsOrgIDInstancesInstanceIDSenderIdentity,
  usePostV1OrgsOrgIDInstancesInstanceIDSenderIdentity,
} from "@/api/generated/endpoints/instances/instances";
import type { SenderIdentityResponse } from "@/api/generated/model";

type Kind = "domain" | "sender_signature" | "platform_domain";

export function SenderIdentityCard({
  orgId,
  instanceId,
}: {
  orgId: string;
  instanceId: string;
}) {
  const identityQuery = useGetV1OrgsOrgIDInstancesInstanceIDSenderIdentity(
    orgId,
    instanceId,
    { query: { enabled: !!orgId && !!instanceId, retry: false } },
  );

  if (identityQuery.isLoading) {
    return <Skeleton className="h-32 w-full" />;
  }

  if (identityQuery.data) {
    const detail = unwrap<SenderIdentityResponse>(identityQuery.data).data;
    if (detail?.identity) {
      return <SenderIdentityStatus detail={detail} />;
    }
  }

  if (identityQuery.error instanceof ApiError && identityQuery.error.status === 404) {
    return <AddSenderIdentityForm orgId={orgId} instanceId={instanceId} />;
  }

  return (
    <div className="rounded-md border border-border p-4">
      <h2 className="mb-1 text-sm font-semibold">Sender identity</h2>
      <p className="text-xs text-status-red">
        Couldn't load this workspace's sender identity.
      </p>
    </div>
  );
}

function SenderIdentityStatus({
  detail,
}: {
  detail: NonNullable<SenderIdentityResponse["data"]>;
}) {
  const identity = detail.identity;
  if (!identity) return null;
  const confirmed = identity.status === "confirmed";

  return (
    <div className="rounded-md border border-border p-4">
      <div className="mb-3 flex items-center justify-between">
        <div>
          <h2 className="text-sm font-semibold">Sender identity</h2>
          <p className="font-mono text-xs text-muted-foreground">
            {identity.value}
          </p>
        </div>
        <Badge
          variant="outline"
          className={
            confirmed
              ? "bg-status-green-soft text-status-green"
              : "bg-status-amber-soft text-status-amber"
          }
        >
          {confirmed ? "Confirmed" : "Pending"}
        </Badge>
      </div>

      {identity.kind === "domain" &&
        (detail.dns_records && detail.dns_records.length > 0 ? (
          <div className="flex flex-col gap-2">
            <p className="text-xs text-muted-foreground">
              Publish this record with your DNS provider to verify the
              domain:
            </p>
            {detail.dns_records.map((record) => (
              <div
                key={record.id}
                className="rounded-md bg-secondary px-2.5 py-2 font-mono text-xs"
              >
                <div className="text-muted-foreground">
                  {record.record_type?.toUpperCase()} · {record.host}
                </div>
                <div className="truncate">{record.value}</div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">
            No DNS record was returned for this domain.
          </p>
        ))}

      {identity.kind === "platform_domain" && (
        <p className="text-xs text-muted-foreground">
          {confirmed
            ? "Hosted on our shared domain — ready to send."
            : "Hosted on our shared domain — we're setting up the DNS on our side, no action needed from you."}
        </p>
      )}

      {identity.kind === "sender_signature" && (
        <p className="text-xs text-muted-foreground">
          Check {identity.value} for a confirmation email from Postmark — no
          DNS changes needed.
        </p>
      )}
    </div>
  );
}

function AddSenderIdentityForm({
  orgId,
  instanceId,
}: {
  orgId: string;
  instanceId: string;
}) {
  const [kind, setKind] = useState<Kind>("domain");
  const [domain, setDomain] = useState("");
  const [signatureName, setSignatureName] = useState("");
  const [signatureEmail, setSignatureEmail] = useState("");
  const queryClient = useQueryClient();

  const addIdentity = usePostV1OrgsOrgIDInstancesInstanceIDSenderIdentity({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: getGetV1OrgsOrgIDInstancesInstanceIDSenderIdentityQueryKey(
            orgId,
            instanceId,
          ),
        });
        toast.success(
          kind === "domain"
            ? "Sending domain added — publish the DKIM record below to verify it"
            : kind === "platform_domain"
              ? "Sending domain set up — we'll handle verification"
              : "Confirmation email sent — check that inbox",
        );
      },
      onError: (error) => {
        toast.error(error.error ?? "Couldn't add sender identity");
      },
    },
  });

  const canSubmit =
    kind === "domain"
      ? domain.trim().length > 0
      : kind === "platform_domain"
        ? true
        : signatureEmail.trim().length > 0 && signatureName.trim().length > 0;

  function submit() {
    addIdentity.mutate({
      orgID: orgId,
      instanceID: instanceId,
      data:
        kind === "domain"
          ? { kind, value: domain }
          : kind === "platform_domain"
            ? { kind }
            : { kind, value: signatureEmail, name: signatureName },
    });
  }

  return (
    <div className="flex flex-col gap-3 rounded-md border border-border p-4">
      <div>
        <h2 className="text-sm font-semibold">Sender identity</h2>
        <p className="text-xs text-muted-foreground">
          Add the domain or address this workspace sends from — required
          before it can send real mail.
        </p>
      </div>

      <Tabs value={kind} onValueChange={(v) => setKind(v as Kind)}>
        <TabsList>
          <TabsTrigger value="domain">Domain</TabsTrigger>
          <TabsTrigger value="sender_signature">Sender signature</TabsTrigger>
          <TabsTrigger value="platform_domain">No domain? Use ours</TabsTrigger>
        </TabsList>

        <TabsContent value="domain">
          <div className="flex flex-col gap-2 pt-3">
            <Label htmlFor="sender-domain">Sending domain</Label>
            <Input
              id="sender-domain"
              placeholder="mail.acme.com"
              value={domain}
              onChange={(e) => setDomain(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              You'll get a DKIM record to publish before this domain is
              trusted.
            </p>
          </div>
        </TabsContent>

        <TabsContent value="platform_domain">
          <p className="pt-3 text-xs text-muted-foreground">
            Don't have a domain to send from? We'll set one up for you on a
            domain we already host and verify — nothing for you to
            configure.
          </p>
        </TabsContent>

        <TabsContent value="sender_signature">
          <div className="flex flex-col gap-3 pt-3">
            <div className="flex flex-col gap-2">
              <Label htmlFor="signature-name">Display name</Label>
              <Input
                id="signature-name"
                placeholder="Acme Notifications"
                value={signatureName}
                onChange={(e) => setSignatureName(e.target.value)}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="signature-email">From address</Label>
              <Input
                id="signature-email"
                type="email"
                placeholder="hello@acme.com"
                value={signatureEmail}
                onChange={(e) => setSignatureEmail(e.target.value)}
              />
            </div>
            <p className="text-xs text-muted-foreground">
              Postmark emails a confirmation link to this address — no DNS
              changes needed.
            </p>
          </div>
        </TabsContent>
      </Tabs>

      <Button
        className="self-start"
        disabled={!canSubmit || addIdentity.isPending}
        onClick={submit}
      >
        {addIdentity.isPending ? "Adding…" : "Add"}
      </Button>
    </div>
  );
}
