import { Badge } from "@/components/ui/badge";
import { Container } from "@/components/landing/container";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { SectionEyebrow } from "@/components/landing/section-eyebrow";

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
          { label: "provision_smtp_server", status: "Succeeded" },
          { label: "dns · dkim._domainkey.acme-co", status: "Verified" },
        ].map((row, i) => (
          <div
            key={row.label}
            style={{ animationDelay: `${i * 150}ms` }}
            className="animate-in fade-in slide-in-from-bottom-1 flex items-center justify-between border-b border-dashed border-border pb-2 text-xs duration-500 fill-mode-both last:border-b-0 last:pb-0 motion-reduce:animate-none"
          >
            <span className="font-mono text-muted-foreground">{row.label}</span>
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

export function Hero({ onSignIn }: { onSignIn: () => void }) {
  return (
    <section>
      <Container className="grid gap-12 px-6 py-16 md:grid-cols-2 md:items-center md:pb-12 md:pt-24">
        <div>
          <div className="flex items-center gap-3">
            <SectionEyebrow>
              Newsletter &amp; mailing list manager by{" "}
              <a href="https://n0.rocks" className="text-foreground">
                Nerd Zero
              </a>
            </SectionEyebrow>
          </div>
          <h1 className="mt-4 text-[72px] leading-[81px] font-normal tracking-tight text-balance">
            The newsletter and mailing list manager designed for ownership.
          </h1>
          <p className="mt-5 max-w-md text-base text-muted-foreground">
            A fully isolated workspace with your own sending domain, OIDC
            authentication, granular roles, and high-throughput multi-SMTP
            delivery. Manage millions of subscribers. Nothing shared, nothing to
            run yourself.
          </p>
          <div className="mt-7 flex flex-wrap items-center gap-3">
            <Button size="lg" onClick={onSignIn}>
              Get started
            </Button>
            <a
              href="#showcase"
              className="text-sm font-medium text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
            >
              See what's included →
            </a>
          </div>
          <p className="mt-4 text-xs text-muted-foreground">
            Free and open source · AGPLv3 · Built on listmonk
          </p>
        </div>

        <div className="flex justify-center md:justify-end">
          <ProvisioningConsole />
        </div>
      </Container>
    </section>
  );
}
