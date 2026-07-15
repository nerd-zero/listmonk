import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
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
  getGetV1OrgsOrgIDInstancesInstanceIDPostmarkServerQueryKey,
  useDeleteV1OrgsOrgIDInstancesInstanceIDPostmarkServer,
  useGetV1OrgsOrgIDInstancesInstanceIDPostmarkServer,
  usePostV1OrgsOrgIDInstancesInstanceIDPostmarkServerResync,
} from "@/api/generated/endpoints/instances/instances";
import type { PostmarkServerResponse } from "@/api/generated/model";

export function PostmarkServerCard({
  orgId,
  instanceId,
}: {
  orgId: string;
  instanceId: string;
}) {
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);
  const queryClient = useQueryClient();

  const serverQuery = useGetV1OrgsOrgIDInstancesInstanceIDPostmarkServer(
    orgId,
    instanceId,
    { query: { enabled: !!orgId && !!instanceId, retry: false } },
  );

  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: getGetV1OrgsOrgIDInstancesInstanceIDPostmarkServerQueryKey(
        orgId,
        instanceId,
      ),
    });

  const resync = usePostV1OrgsOrgIDInstancesInstanceIDPostmarkServerResync({
    mutation: {
      onSuccess: () => {
        toast.success("Re-synced SMTP credentials to listmonk");
      },
      onError: (error) => {
        toast.error(error.error ?? "Couldn't re-sync SMTP credentials");
      },
    },
  });

  const deleteServer = useDeleteV1OrgsOrgIDInstancesInstanceIDPostmarkServer({
    mutation: {
      onSuccess: () => {
        toast.success("Postmark server deleted");
        setConfirmDeleteOpen(false);
        invalidate();
      },
      onError: (error) => {
        toast.error(error.error ?? "Couldn't delete the Postmark server");
      },
    },
  });

  if (serverQuery.isLoading) {
    return <Skeleton className="h-24 w-full" />;
  }

  if (serverQuery.error instanceof ApiError && serverQuery.error.status === 404) {
    return (
      <div className="rounded-md border border-border p-4">
        <h2 className="mb-1 text-sm font-semibold">Postmark server</h2>
        <p className="text-xs text-muted-foreground">
          No Postmark server yet for this instance.
        </p>
      </div>
    );
  }

  if (!serverQuery.data) {
    return (
      <div className="rounded-md border border-border p-4">
        <h2 className="mb-1 text-sm font-semibold">Postmark server</h2>
        <p className="text-xs text-status-red">
          Couldn't load this workspace's Postmark server.
        </p>
      </div>
    );
  }

  const server = unwrap<PostmarkServerResponse>(serverQuery.data).data;
  if (!server) return null;

  return (
    <div className="rounded-md border border-border p-4">
      <div className="mb-3 flex items-center justify-between">
        <div>
          <h2 className="text-sm font-semibold">Postmark server</h2>
          <p className="font-mono text-xs text-muted-foreground">
            {server.name} · #{server.postmark_id}
          </p>
        </div>
        <Badge
          variant="outline"
          className={
            server.smtp_api_activated
              ? "bg-status-green-soft text-status-green"
              : "bg-status-amber-soft text-status-amber"
          }
        >
          {server.smtp_api_activated ? "SMTP active" : "SMTP inactive"}
        </Badge>
      </div>

      <div className="flex gap-2">
        <Button
          variant="outline"
          size="sm"
          disabled={resync.isPending}
          onClick={() =>
            resync.mutate({ orgID: orgId, instanceID: instanceId })
          }
        >
          {resync.isPending ? "Re-syncing…" : "Re-sync SMTP credentials"}
        </Button>
        <Button
          variant="ghost"
          size="sm"
          className="text-destructive hover:text-destructive"
          onClick={() => setConfirmDeleteOpen(true)}
        >
          Delete
        </Button>
      </div>

      <AlertDialog open={confirmDeleteOpen} onOpenChange={setConfirmDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Postmark server?</AlertDialogTitle>
            <AlertDialogDescription>
              This permanently deletes {server.name}. Email sending stops
              until a new one is provisioned. This cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteServer.isPending}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              disabled={deleteServer.isPending}
              onClick={() =>
                deleteServer.mutate({ orgID: orgId, instanceID: instanceId })
              }
            >
              {deleteServer.isPending ? "Deleting…" : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
