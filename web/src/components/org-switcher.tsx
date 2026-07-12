import { useState } from "react";
import { ChevronsUpDown, Plus } from "lucide-react";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Sheet,
  SheetContent,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { cn } from "@/lib/utils";
import { useOrgContext } from "@/lib/org-context";
import {
  getGetV1OrgsQueryKey,
  usePostV1Orgs,
} from "@/api/generated/endpoints/orgs/orgs";
import type { OrgResponse } from "@/api/generated/model";
import { unwrap } from "@/api/unwrap";

export function OrgSwitcher() {
  const { orgs, selectedOrg, setSelectedOrgId } = useOrgContext();
  const [newOrgOpen, setNewOrgOpen] = useState(false);

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger
          className={cn(
            buttonVariants({ variant: "ghost" }),
            "w-full justify-between px-2 font-semibold",
          )}
        >
          <span className="truncate">{selectedOrg?.name ?? "Select org"}</span>
          <ChevronsUpDown className="size-3.5 text-muted-foreground" />
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="w-56">
          {orgs.map((org) => (
            <DropdownMenuItem
              key={org.id}
              onClick={() => org.id && setSelectedOrgId(org.id)}
              className={org.id === selectedOrg?.id ? "font-semibold" : undefined}
            >
              {org.name}
              {org.id === selectedOrg?.id && " ✓"}
            </DropdownMenuItem>
          ))}
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={() => setNewOrgOpen(true)}>
            <Plus className="size-3.5" /> New org
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <NewOrgSheet open={newOrgOpen} onOpenChange={setNewOrgOpen} />
    </>
  );
}

function NewOrgSheet({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const [name, setName] = useState("");
  const queryClient = useQueryClient();
  const { setSelectedOrgId } = useOrgContext();

  const createOrg = usePostV1Orgs({
    mutation: {
      onSuccess: (result) => {
        queryClient.invalidateQueries({ queryKey: getGetV1OrgsQueryKey() });
        const newOrgId = unwrap<OrgResponse>(result).data?.id;
        if (newOrgId) setSelectedOrgId(newOrgId);
        toast.success(`Created "${name}"`);
        setName("");
        onOpenChange(false);
      },
      onError: (error) => {
        toast.error(error.error ?? "Couldn't create the org");
      },
    },
  });

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>New org</SheetTitle>
        </SheetHeader>
        <div className="flex flex-col gap-2 px-4">
          <Label htmlFor="new-org-name">Org name</Label>
          <Input
            id="new-org-name"
            placeholder="Acme Co"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus
          />
        </div>
        <SheetFooter className="flex-row">
          <Button
            variant="ghost"
            className="flex-1"
            onClick={() => onOpenChange(false)}
            disabled={createOrg.isPending}
          >
            Cancel
          </Button>
          <Button
            className="flex-1"
            onClick={() => createOrg.mutate({ data: { name } })}
            disabled={!name.trim() || createOrg.isPending}
          >
            {createOrg.isPending ? "Creating…" : "Create"}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
