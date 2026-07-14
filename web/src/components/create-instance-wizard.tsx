import { useEffect, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Check, Copy, Loader2 } from "lucide-react";
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
import { cn } from "@/lib/utils";
import { instanceSlug } from "@/lib/slug";
import { unwrap } from "@/api/unwrap";
import {
  getGetV1OrgsOrgIDInstancesQueryKey,
  usePostV1OrgsOrgIDInstances,
} from "@/api/generated/endpoints/instances/instances";
import type { DbInstance, InstanceResponse } from "@/api/generated/model";

type Step = "details" | "admin" | "review" | "provisioning" | "error" | "done";

// "provisioning" and "error" both render on the last dot -- they're
// mid-flight toward "done", not a step of their own in the progress bar.
const DOTS: Step[] = ["details", "admin", "review", "done"];

// These are the real steps CreateInstance runs through server-side (see
// internal/provisioning.CreateInstance) -- ticked forward on a timer while
// the single blocking POST is in flight, since there's no intermediate
// progress to poll mid-request. Worded to stay true regardless of whether
// this deployment has Postmark configured.
const PROVISION_LABELS = [
  "Creating your workspace",
  "Setting up your sending domain",
  "Finishing up",
];

function dotIndex(step: Step) {
  if (step === "provisioning" || step === "error") return DOTS.indexOf("done");
  return DOTS.indexOf(step);
}

export function CreateInstanceWizard({
  open,
  onOpenChange,
  orgId,
  orgName,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId?: string;
  orgName?: string;
}) {
  const [step, setStep] = useState<Step>("details");
  const [name, setName] = useState("");
  const [adminUsername, setAdminUsername] = useState("");
  const [adminEmail, setAdminEmail] = useState("");
  const [labelIndex, setLabelIndex] = useState(0);
  const [result, setResult] = useState<DbInstance>();
  const [errorMessage, setErrorMessage] = useState("");
  const queryClient = useQueryClient();

  const slug = instanceSlug(orgName ?? "", name);

  const createInstance = usePostV1OrgsOrgIDInstances({
    mutation: {
      onSuccess: (response) => {
        if (orgId) {
          queryClient.invalidateQueries({
            queryKey: getGetV1OrgsOrgIDInstancesQueryKey(orgId),
          });
        }
        setResult(unwrap<InstanceResponse>(response).data);
        setStep("done");
      },
      onError: (error) => {
        setErrorMessage(error.error ?? "Couldn't create the instance");
        setStep("error");
      },
    },
  });

  // Fake-but-honest progress: advances on a timer, never past the last
  // label, and only ever reaches "done" once the real response comes back.
  useEffect(() => {
    if (step !== "provisioning") return;
    const id = setInterval(() => {
      setLabelIndex((i) => Math.min(i + 1, PROVISION_LABELS.length - 1));
    }, 900);
    return () => clearInterval(id);
  }, [step]);

  function reset() {
    setStep("details");
    setName("");
    setAdminUsername("");
    setAdminEmail("");
    setLabelIndex(0);
    setResult(undefined);
    setErrorMessage("");
  }

  function close() {
    onOpenChange(false);
    // Let the sheet's own close transition finish before wiping the form,
    // so it doesn't visibly reset mid-animation.
    setTimeout(reset, 200);
  }

  function submit() {
    if (!orgId) return;
    setStep("provisioning");
    setLabelIndex(0);
    createInstance.mutate({
      orgID: orgId,
      data: {
        slug,
        name,
        admin_username: adminUsername,
        admin_email: adminEmail,
      },
    });
  }

  async function copySetupUrl() {
    if (!result?.admin_setup_url) return;
    try {
      await navigator.clipboard.writeText(result.admin_setup_url);
      toast.success("Setup link copied");
    } catch {
      toast.error("Couldn't copy — select and copy the link manually");
    }
  }

  const detailsValid = !!slug && name.trim().length > 0;
  const adminValid =
    adminUsername.trim().length > 0 && adminEmail.trim().length > 0;

  return (
    <Sheet
      open={open}
      onOpenChange={(next) => {
        if (!next) close();
        else onOpenChange(next);
      }}
    >
      <SheetContent className="flex flex-col">
        <SheetHeader>
          <SheetTitle>New instance</SheetTitle>
          <SheetDescription>
            {step === "done"
              ? `${result?.slug} is live.`
              : "This becomes your workspace's address and its first admin."}
          </SheetDescription>
        </SheetHeader>

        <div className="flex gap-1.5 px-4">
          {DOTS.map((s, i) => (
            <div
              key={s}
              className={cn(
                "h-1 flex-1 rounded-full bg-border",
                i <= dotIndex(step) && "bg-ring",
              )}
            />
          ))}
        </div>

        <div className="flex flex-1 flex-col gap-4 px-4">
          {step === "details" && (
            <div className="flex flex-col gap-2">
              <Label htmlFor="name">Workspace name</Label>
              <Input
                id="name"
                placeholder="Marketing"
                value={name}
                onChange={(e) => setName(e.target.value)}
                autoFocus
              />
              {slug && (
                <p className="font-mono text-xs text-muted-foreground">
                  {slug}
                </p>
              )}
            </div>
          )}

          {step === "admin" && (
            <>
              <div className="flex flex-col gap-2">
                <Label htmlFor="admin-username">Admin username</Label>
                <Input
                  id="admin-username"
                  value={adminUsername}
                  onChange={(e) => setAdminUsername(e.target.value)}
                  autoFocus
                />
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor="admin-email">Admin email</Label>
                <Input
                  id="admin-email"
                  type="email"
                  value={adminEmail}
                  onChange={(e) => setAdminEmail(e.target.value)}
                />
              </div>
            </>
          )}

          {step === "review" && (
            <div className="flex flex-col gap-3 rounded-md border border-border bg-secondary/40 p-4">
              <ReviewRow label="Organization" value={orgName ?? "—"} />
              <ReviewRow label="Workspace" value={slug} mono />
              <ReviewRow label="Admin" value={adminEmail} />
            </div>
          )}

          {(step === "provisioning" || step === "error") && (
            <div className="flex flex-col gap-3 rounded-md border border-border p-4">
              {PROVISION_LABELS.map((label, i) => (
                <div key={label} className="flex items-center gap-2.5 text-sm">
                  {step === "error" && i === labelIndex ? (
                    <div className="size-4 shrink-0 rounded-full bg-status-red-soft" />
                  ) : i < labelIndex ? (
                    <Check className="size-4 shrink-0 text-status-green" />
                  ) : i === labelIndex && step === "provisioning" ? (
                    <Loader2 className="size-4 shrink-0 animate-spin text-ring" />
                  ) : (
                    <div className="size-4 shrink-0 rounded-full border border-border" />
                  )}
                  <span
                    className={
                      i <= labelIndex ? "text-foreground" : "text-muted-foreground"
                    }
                  >
                    {label}
                  </span>
                </div>
              ))}
              {step === "error" && (
                <p className="mt-1 text-xs text-status-red">{errorMessage}</p>
              )}
            </div>
          )}

          {step === "done" && result?.admin_setup_url && (
            <div className="flex flex-col gap-2">
              <p className="text-sm text-muted-foreground">
                Send this one-time link to {result.admin_email} to set a
                password.
              </p>
              <div className="flex items-center gap-2">
                <code className="flex-1 truncate rounded-md bg-secondary px-2 py-1.5 text-xs">
                  {result.admin_setup_url}
                </code>
                <Button
                  variant="outline"
                  size="icon"
                  onClick={copySetupUrl}
                  aria-label="Copy setup link"
                >
                  <Copy className="size-3.5" />
                </Button>
              </div>
            </div>
          )}
        </div>

        <SheetFooter className="flex-row">
          {step === "details" && (
            <>
              <Button variant="ghost" className="flex-1" onClick={close}>
                Cancel
              </Button>
              <Button
                className="flex-1"
                disabled={!detailsValid}
                onClick={() => setStep("admin")}
              >
                Continue
              </Button>
            </>
          )}
          {step === "admin" && (
            <>
              <Button
                variant="ghost"
                className="flex-1"
                onClick={() => setStep("details")}
              >
                Back
              </Button>
              <Button
                className="flex-1"
                disabled={!adminValid}
                onClick={() => setStep("review")}
              >
                Review
              </Button>
            </>
          )}
          {step === "review" && (
            <>
              <Button
                variant="ghost"
                className="flex-1"
                onClick={() => setStep("admin")}
              >
                Back
              </Button>
              <Button className="flex-1" onClick={submit}>
                Create instance
              </Button>
            </>
          )}
          {step === "provisioning" && (
            <Button className="flex-1" disabled>
              Creating…
            </Button>
          )}
          {step === "error" && (
            <>
              <Button
                variant="ghost"
                className="flex-1"
                onClick={() => setStep("details")}
              >
                Edit details
              </Button>
              <Button className="flex-1" onClick={submit}>
                Try again
              </Button>
            </>
          )}
          {step === "done" && (
            <Button className="flex-1" onClick={close}>
              Done
            </Button>
          )}
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}

function ReviewRow({
  label,
  value,
  mono,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <div className="flex items-center justify-between text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span className={cn("font-medium", mono && "font-mono")}>{value}</span>
    </div>
  );
}
