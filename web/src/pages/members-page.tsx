import { useState } from "react";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Sheet,
  SheetContent,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useOrgContext } from "@/lib/org-context";
import { usePermissions } from "@/lib/permissions";
import { unwrap } from "@/api/unwrap";
import {
  getGetV1OrgsOrgIDMembersQueryKey,
  useGetV1OrgsOrgIDMembers,
  usePostV1OrgsOrgIDMembers,
} from "@/api/generated/endpoints/members/members";
import type { MemberListResponse } from "@/api/generated/model";

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function LoadingSkeleton() {
  return (
    <div className="flex flex-col gap-3">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-14 w-full" />
      <Skeleton className="h-14 w-full" />
    </div>
  );
}

function initialFor(displayName?: string, email?: string) {
  const source = displayName?.trim() || email || "";
  return source.charAt(0).toUpperCase() || "?";
}

export function MembersPage() {
  const { selectedOrg, isLoading: orgLoading } = useOrgContext();
  const { can } = usePermissions();
  const [inviteOpen, setInviteOpen] = useState(false);

  const membersQuery = useGetV1OrgsOrgIDMembers(selectedOrg?.id ?? "", {
    query: { enabled: !!selectedOrg?.id },
  });

  if (orgLoading) {
    return <LoadingSkeleton />;
  }

  if (!selectedOrg) {
    return (
      <div className="flex flex-col items-center gap-3 rounded-md border border-dashed border-border py-16 text-center">
        <h2 className="font-mono text-lg font-semibold">No org yet</h2>
        <p className="max-w-sm text-sm text-muted-foreground">
          Create an org from the sidebar before inviting anyone to it.
        </p>
      </div>
    );
  }

  if (membersQuery.isLoading) {
    return <LoadingSkeleton />;
  }

  const members = membersQuery.data
    ? (unwrap<MemberListResponse>(membersQuery.data).data ?? [])
    : [];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-baseline justify-between">
        <div>
          <h1 className="font-mono text-2xl font-bold tracking-tight">
            Members
          </h1>
          <p className="text-sm text-muted-foreground">
            Everyone with access to <strong>{selectedOrg.name}</strong>.
          </p>
        </div>
        {can("inviteMember") && (
          <Button onClick={() => setInviteOpen(true)}>Invite member</Button>
        )}
      </div>

      {members.length === 0 ? (
        <div className="flex flex-col items-center gap-3 rounded-md border border-dashed border-border py-16 text-center">
          <h2 className="font-mono text-lg font-semibold">
            No members yet
          </h2>
          <p className="max-w-sm text-sm text-muted-foreground">
            Invite someone to help run {selectedOrg.name}.
          </p>
          {can("inviteMember") && (
            <Button onClick={() => setInviteOpen(true)}>Invite member</Button>
          )}
        </div>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Person</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Joined</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {members.map((member) => (
              <TableRow key={member.id}>
                <TableCell>
                  <div className="flex items-center gap-3">
                    <Avatar className="size-7">
                      <AvatarFallback className="text-xs">
                        {initialFor(member.display_name, member.email)}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <div className="text-sm font-medium">
                        {member.display_name || member.email}
                      </div>
                      {member.display_name && (
                        <div className="text-xs text-muted-foreground">
                          {member.email}
                        </div>
                      )}
                    </div>
                  </div>
                </TableCell>
                <TableCell>
                  <Badge
                    variant="outline"
                    className={
                      member.role === "owner"
                        ? "bg-status-amber-soft text-status-amber"
                        : "bg-secondary text-muted-foreground"
                    }
                  >
                    {member.role === "owner" ? "Owner" : "Member"}
                  </Badge>
                </TableCell>
                <TableCell className="font-mono text-xs text-muted-foreground">
                  {member.created_at
                    ? new Date(member.created_at).toLocaleDateString()
                    : "—"}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <InviteMemberSheet
        open={inviteOpen}
        onOpenChange={setInviteOpen}
        orgId={selectedOrg.id}
      />
    </div>
  );
}

function InviteMemberSheet({
  open,
  onOpenChange,
  orgId,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId?: string;
}) {
  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [role, setRole] = useState<"member" | "owner">("member");
  const queryClient = useQueryClient();

  const inviteMember = usePostV1OrgsOrgIDMembers({
    mutation: {
      onSuccess: () => {
        if (orgId) {
          queryClient.invalidateQueries({
            queryKey: getGetV1OrgsOrgIDMembersQueryKey(orgId),
          });
        }
        toast.success(`Invited ${email}`);
        setEmail("");
        setDisplayName("");
        setRole("member");
        onOpenChange(false);
      },
      onError: (error) => {
        toast.error(error.error ?? "Couldn't invite that person");
      },
    },
  });

  const emailValid = EMAIL_PATTERN.test(email);
  const canSubmit = emailValid && !!orgId;

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>Invite member</SheetTitle>
          <SheetDescription>
            They'll get an email to set up their account.
          </SheetDescription>
        </SheetHeader>
        <div className="flex flex-col gap-4 px-4">
          <div className="flex flex-col gap-2">
            <Label htmlFor="invite-email">Email</Label>
            <Input
              id="invite-email"
              type="email"
              placeholder="ada@acme.co"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoFocus
            />
            {email && !emailValid && (
              <p className="text-xs text-destructive">
                Enter a valid email address.
              </p>
            )}
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="invite-name">Name (optional)</Label>
            <Input
              id="invite-name"
              placeholder="Ada Lovelace"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label>Role</Label>
            <div className="flex gap-2">
              <Button
                type="button"
                variant={role === "member" ? "default" : "outline"}
                size="sm"
                className="flex-1"
                onClick={() => setRole("member")}
              >
                Member
              </Button>
              <Button
                type="button"
                variant={role === "owner" ? "default" : "outline"}
                size="sm"
                className="flex-1"
                onClick={() => setRole("owner")}
              >
                Owner
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">
              Owners can invite people and manage instances; members can only
              manage instances.
            </p>
          </div>
        </div>
        <SheetFooter className="flex-row">
          <Button
            variant="ghost"
            className="flex-1"
            onClick={() => onOpenChange(false)}
            disabled={inviteMember.isPending}
          >
            Cancel
          </Button>
          <Button
            className="flex-1"
            disabled={!canSubmit || inviteMember.isPending}
            onClick={() =>
              orgId &&
              inviteMember.mutate({
                orgID: orgId,
                data: { email, display_name: displayName, role },
              })
            }
          >
            {inviteMember.isPending ? "Inviting…" : "Send invite"}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
