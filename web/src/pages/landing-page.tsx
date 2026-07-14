import type { ReactNode } from "react";
import { KeyRound, Mail, Server, Users } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";

const PIPELINE_STEPS = [
  {
    index: "01",
    title: "A workspace is created",
    body: "Your own org, ready the moment you sign in — nothing to name or configure first.",
    tag: "orgs.create",
  },
  {
    index: "02",
    title: "Listmonk comes up, isolated",
    body: "A private Listmonk tenant is provisioned for you alone. No shared rows, no shared login.",
    tag: "provision_listmonk_tenant",
  },
  {
    index: "03",
    title: "Your own sending identity",
    body: "A dedicated Postmark server and a verified sending domain, so your deliverability is never someone else's problem.",
    tag: "provision_postmark_server",
  },
  {
    index: "04",
    title: "You're live",
    body: "A one-time setup link lands in your dashboard. Set a password on your new workspace and start sending.",
    tag: "status: active",
  },
];

const FEATURES = [
  {
    icon: Mail,
    title: "Dedicated sending domain",
    body: "Every workspace gets its own Postmark server and DKIM-verified domain — never a shared IP, never someone else's spam complaints.",
  },
  {
    icon: Server,
    title: "Open source underneath",
    body: "Built on a Listmonk fork, not a black box behind an API. Your lists and your data export the same way they always could.",
  },
  {
    icon: Users,
    title: "Built for a team",
    body: "Invite teammates with your existing SSO. Owner and member roles keep billing and admin actions where they belong.",
  },
  {
    icon: KeyRound,
    title: "One console, every workspace",
    body: "Run one newsletter or ten — create, monitor, and hand off admin access to every workspace from a single dashboard.",
  },
];

function SectionEyebrow({ children }: { children: ReactNode }) {
  return (
    <span className="text-[11px] font-semibold tracking-[0.14em] text-muted-foreground uppercase">
      {children}
    </span>
  );
}

function ProvisioningConsole() {
  return (
    <Card className="w-full max-w-md ring-foreground/10">
      <CardHeader className="flex-row items-center justify-between border-b border-border pb-4">
        <div>
          <CardTitle className="font-mono text-sm font-semibold">
            acme-co
          </CardTitle>
          <CardDescription className="font-mono text-xs">
            acme-co.listmonk.test
          </CardDescription>
        </div>
        <Badge
          variant="outline"
          className="animate-in fade-in bg-status-green-soft text-status-green delay-700 duration-500 motion-reduce:animate-none"
        >
          Live
        </Badge>
      </CardHeader>
      <CardContent className="flex flex-col gap-3 pt-1">
        {[
          { label: "create_org", status: "Succeeded" },
          { label: "provision_listmonk_tenant", status: "Succeeded" },
          { label: "provision_postmark_server", status: "Succeeded" },
          { label: "dns · dkim._domainkey.acme-co", status: "Verified" },
        ].map((row, i) => (
          <div
            key={row.label}
            style={{ animationDelay: `${i * 150}ms` }}
            className="animate-in fade-in slide-in-from-bottom-1 flex items-center justify-between border-b border-dashed border-border pb-2 text-xs duration-500 fill-mode-both last:border-b-0 last:pb-0 motion-reduce:animate-none"
          >
            <span className="font-mono text-muted-foreground">
              {row.label}
            </span>
            <span className="shrink-0 pl-3 font-medium text-status-green">
              {row.status}
            </span>
          </div>
        ))}
        <div
          style={{ animationDelay: "600ms" }}
          className="animate-in fade-in slide-in-from-bottom-1 mt-1 rounded-md bg-secondary px-2.5 py-2 font-mono text-[11px] text-muted-foreground duration-500 fill-mode-both motion-reduce:animate-none"
        >
          setup_url: https://acme-co.listmonk.test/admin/setup?token=•••
        </div>
      </CardContent>
    </Card>
  );
}

export function LandingPage({ onSignIn }: { onSignIn: () => void }) {
  return (
    <div className="flex min-h-svh flex-col bg-background bg-[repeating-linear-gradient(0deg,var(--border)_0,var(--border)_1px,transparent_1px,transparent_48px),repeating-linear-gradient(90deg,var(--border)_0,var(--border)_1px,transparent_1px,transparent_48px)] bg-[length:100%_48px,48px_100%]">
      <header className="mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-6">
        <div>
          <div className="font-mono text-lg font-semibold tracking-tight">
            listnun
          </div>
          <span className="text-[11px] font-semibold tracking-[0.14em] text-muted-foreground uppercase">
            Tenant console
          </span>
        </div>
        <Button variant="outline" size="sm" onClick={onSignIn}>
          Sign in
        </Button>
      </header>

      <main className="flex-1">
        <section className="mx-auto grid w-full max-w-6xl gap-12 px-6 py-16 md:grid-cols-2 md:items-center md:py-24">
          <div>
            <SectionEyebrow>Multi-tenant Listmonk, hosted</SectionEyebrow>
            <h1 className="mt-4 font-mono text-4xl leading-[1.05] font-semibold tracking-tight text-balance sm:text-5xl">
              Your own Listmonk.
              <br />
              Your own domain.
              <br />
              Actually yours.
            </h1>
            <p className="mt-5 max-w-md text-base text-muted-foreground">
              listnun provisions a private Listmonk workspace in about a
              minute — its own admin login, its own verified sending domain,
              its own dedicated Postmark server. Nothing shared, nothing to
              run yourself.
            </p>
            <div className="mt-7 flex flex-wrap items-center gap-3">
              <Button size="lg" onClick={onSignIn}>
                Continue with SSO
              </Button>
              <a
                href="#pipeline"
                className="text-sm font-medium text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
              >
                See what happens →
              </a>
            </div>
            <p className="mt-4 text-xs text-muted-foreground">
              You'll be redirected to sign in, then sent right back here.
            </p>
          </div>

          <div className="flex justify-center md:justify-end">
            <ProvisioningConsole />
          </div>
        </section>

        <section id="pipeline" className="border-t border-border">
          <div className="mx-auto w-full max-w-6xl px-6 py-16 md:py-24">
            <SectionEyebrow>How it works</SectionEyebrow>
            <h2 className="mt-3 max-w-xl font-mono text-2xl font-semibold tracking-tight sm:text-3xl">
              What actually happens when you sign up
            </h2>
            <p className="mt-3 max-w-xl text-sm text-muted-foreground">
              Every workspace goes through the same four steps, in the same
              order — the same timeline your own dashboard shows for every
              workspace you create.
            </p>

            <ol className="mt-10 grid gap-px overflow-hidden rounded-xl bg-border sm:grid-cols-2 lg:grid-cols-4">
              {PIPELINE_STEPS.map((step) => (
                <li key={step.index} className="flex flex-col gap-3 bg-card p-6">
                  <span className="font-mono text-2xl font-semibold text-ring">
                    {step.index}
                  </span>
                  <h3 className="text-sm font-semibold">{step.title}</h3>
                  <p className="flex-1 text-sm text-muted-foreground">
                    {step.body}
                  </p>
                  <code className="text-[11px] text-muted-foreground/70">
                    {step.tag}
                  </code>
                </li>
              ))}
            </ol>
          </div>
        </section>

        <section className="border-t border-border">
          <div className="mx-auto w-full max-w-6xl px-6 py-16 md:py-24">
            <SectionEyebrow>Why not just use one shared instance</SectionEyebrow>
            <h2 className="mt-3 max-w-xl font-mono text-2xl font-semibold tracking-tight sm:text-3xl">
              Everything a sender needs, minus the ops
            </h2>

            <div className="mt-10 grid gap-6 sm:grid-cols-2">
              {FEATURES.map(({ icon: Icon, title, body }) => (
                <div
                  key={title}
                  className={cn(
                    "flex gap-4 rounded-xl border border-border bg-card p-6",
                  )}
                >
                  <div className="flex size-9 shrink-0 items-center justify-center rounded-md bg-secondary">
                    <Icon className="size-4.5" />
                  </div>
                  <div>
                    <h3 className="text-sm font-semibold">{title}</h3>
                    <p className="mt-1.5 text-sm text-muted-foreground">
                      {body}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section className="border-t border-border bg-secondary/60">
          <div className="mx-auto flex w-full max-w-6xl flex-col items-center gap-5 px-6 py-16 text-center md:py-20">
            <h2 className="font-mono text-2xl font-semibold tracking-tight sm:text-3xl">
              Ready to run your own list?
            </h2>
            <p className="max-w-md text-sm text-muted-foreground">
              Sign in with the account you already use — your workspace is
              provisioning before the redirect finishes.
            </p>
            <Button size="lg" onClick={onSignIn}>
              Continue with SSO
            </Button>
          </div>
        </section>
      </main>

      <footer className="border-t border-border">
        <div className="mx-auto flex w-full max-w-6xl flex-col items-center justify-between gap-2 px-6 py-8 text-xs text-muted-foreground sm:flex-row">
          <span className="font-mono">listnun</span>
          <span>Built on open-source Listmonk.</span>
        </div>
      </footer>
    </div>
  );
}
