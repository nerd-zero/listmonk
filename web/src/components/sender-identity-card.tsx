import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Copy } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
  getGetV1OrgsOrgIDInstancesInstanceIDSenderIdentityQueryKey,
  useDeleteV1OrgsOrgIDInstancesInstanceIDSenderIdentity,
  useGetV1OrgsOrgIDInstancesInstanceIDSenderIdentity,
  usePostV1OrgsOrgIDInstancesInstanceIDSenderIdentity,
} from "@/api/generated/endpoints/instances/instances";
import type { SenderIdentityResponse } from "@/api/generated/model";

type Kind = "domain" | "sender_signature" | "platform_domain";

const RECORD_TYPE_LABEL: Record<string, string> = {
  dkim: "DKIM (TXT)",
  return_path: "Return-Path (CNAME)",
};

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
      return (
        <SenderIdentityStatus
          orgId={orgId}
          instanceId={instanceId}
          detail={detail}
        />
      );
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
  orgId,
  instanceId,
  detail,
}: {
  orgId: string;
  instanceId: string;
  detail: NonNullable<SenderIdentityResponse["data"]>;
}) {
  const [confirmOpen, setConfirmOpen] = useState(false);
  const queryClient = useQueryClient();

  const deleteIdentity = useDeleteV1OrgsOrgIDInstancesInstanceIDSenderIdentity({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: getGetV1OrgsOrgIDInstancesInstanceIDSenderIdentityQueryKey(
            orgId,
            instanceId,
          ),
        });
        toast.success("Sender identity removed");
        setConfirmOpen(false);
      },
      onError: (error) => {
        toast.error(error.error ?? "Couldn't remove that sender identity");
      },
    },
  });

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
        <div className="flex items-center gap-2">
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

      {identity.kind === "domain" &&
        (detail.dns_records && detail.dns_records.length > 0 ? (
          <div className="flex flex-col gap-3">
            <p className="text-xs text-muted-foreground">
              Publish these {detail.dns_records.length === 1 ? "record" : "records"} with your DNS provider to verify the
              domain:
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

      <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove {identity.value}?</AlertDialogTitle>
            <AlertDialogDescription>
              This permanently removes the sender identity from Postmark and
              this workspace. You'll need to add a new one before this
              workspace can send mail again. This cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteIdentity.isPending}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              disabled={deleteIdentity.isPending}
              onClick={() =>
                deleteIdentity.mutate({ orgID: orgId, instanceID: instanceId })
              }
            >
              {deleteIdentity.isPending ? "Removing…" : "Remove"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
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
