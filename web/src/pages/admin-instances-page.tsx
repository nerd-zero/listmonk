import { useState } from "react";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
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
import { InstanceStatusBadge } from "@/components/instance-status-badge";
import { unwrap } from "@/api/unwrap";
import {
  getGetV1AdminInstancesQueryKey,
  useDeleteV1AdminInstancesInstanceID,
  useDeleteV1AdminInstancesInstanceIDPostmarkServer,
  useGetV1AdminInstances,
} from "@/api/generated/endpoints/admin/admin";
import type {
  AdminInstanceListResponse,
  DbListAllInstancesWithOrgNameRow,
} from "@/api/generated/model";

function LoadingSkeleton() {
  return (
    <div className="flex flex-col gap-3">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-14 w-full" />
      <Skeleton className="h-14 w-full" />
    </div>
  );
}

type PendingAction = {
  instance: DbListAllInstancesWithOrgNameRow;
  kind: "instance" | "postmark-server";
};

export function AdminInstancesPage() {
  const [pending, setPending] = useState<PendingAction | null>(null);
  const queryClient = useQueryClient();

  const instancesQuery = useGetV1AdminInstances();

  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: getGetV1AdminInstancesQueryKey(),
    });

  const deleteInstance = useDeleteV1AdminInstancesInstanceID({
    mutation: {
      onSuccess: () => {
        toast.success(`Deleted ${pending?.instance.name}`);
        setPending(null);
        invalidate();
      },
      onError: (error) => {
        toast.error(error.error ?? "Couldn't delete that instance");
      },
    },
  });

  const deletePostmarkServer = useDeleteV1AdminInstancesInstanceIDPostmarkServer({
    mutation: {
      onSuccess: () => {
        toast.success(`Deleted ${pending?.instance.name}'s Postmark server`);
        setPending(null);
        invalidate();
      },
      onError: (error) => {
        toast.error(error.error ?? "Couldn't delete that Postmark server");
      },
    },
  });

  if (instancesQuery.isLoading) {
    return <LoadingSkeleton />;
  }

  if (instancesQuery.isError) {
    return (
      <div className="flex flex-col items-center gap-3 rounded-md border border-dashed border-border py-16 text-center">
        <h2 className="font-mono text-lg font-semibold">Not authorized</h2>
        <p className="max-w-sm text-sm text-muted-foreground">
          {instancesQuery.error.error ??
            "Admin access is required to view this page."}
        </p>
      </div>
    );
  }

  const instances = instancesQuery.data
    ? (unwrap<AdminInstanceListResponse>(instancesQuery.data).data ?? [])
    : [];

  const isPending = deleteInstance.isPending || deletePostmarkServer.isPending;

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="font-mono text-2xl font-bold tracking-tight">
          All instances
        </h1>
        <p className="text-sm text-muted-foreground">
          Every instance on the platform, across every org. Super admin only.
        </p>
      </div>

      {instances.length === 0 ? (
        <div className="flex flex-col items-center gap-3 rounded-md border border-dashed border-border py-16 text-center">
          <h2 className="font-mono text-lg font-semibold">No instances yet</h2>
        </div>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Org</TableHead>
              <TableHead>Instance</TableHead>
              <TableHead>Slug</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {instances.map((instance) => (
              <TableRow key={instance.id}>
                <TableCell className="text-sm">{instance.org_name}</TableCell>
                <TableCell className="text-sm font-medium">
                  {instance.name}
                </TableCell>
                <TableCell className="font-mono text-xs text-muted-foreground">
                  {instance.slug}
                </TableCell>
                <TableCell>
                  <InstanceStatusBadge status={instance.status} />
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={isPending}
                      onClick={() =>
                        setPending({ instance, kind: "postmark-server" })
                      }
                    >
                      Delete Postmark server
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      disabled={isPending}
                      onClick={() => setPending({ instance, kind: "instance" })}
                    >
                      Delete instance
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <AlertDialog
        open={!!pending}
        onOpenChange={(open) => !open && setPending(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {pending?.kind === "instance"
                ? `Delete ${pending.instance.name}?`
                : `Delete ${pending?.instance.name}'s Postmark server?`}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {pending?.kind === "instance"
                ? "This permanently deletes the instance's Postmark server, its listmonk tenant (subscribers, campaigns, users, settings -- everything), and its record here. This cannot be undone."
                : "This permanently deletes the Postmark server. The instance is left without email sending until a new one is provisioned. This cannot be undone."}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isPending}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              disabled={isPending}
              onClick={() => {
                if (!pending?.instance.id) return;
                if (pending.kind === "instance") {
                  deleteInstance.mutate({ instanceID: pending.instance.id });
                } else {
                  deletePostmarkServer.mutate({
                    instanceID: pending.instance.id,
                  });
                }
              }}
            >
              {isPending ? "Deleting…" : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
