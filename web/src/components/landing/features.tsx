import { KeyRound, Mail, Server, Users } from "lucide-react";
import { Container } from "@/components/landing/container";
import { cn } from "@/lib/utils";
import { SectionEyebrow } from "@/components/landing/section-eyebrow";

const FEATURES = [
  {
    icon: Mail,
    title: "Dedicated sending domain",
    body: "Every workspace gets its own email delivery service (SMTP and API) and DKIM-verified domain. Never a shared IP, never someone else's spam complaints.",
  },
  {
    icon: Server,
    title: "Open source underneath",
    body: "Built on a Listmonk fork, not a black box behind an API. Your lists and your data export the same way they always could.",
  },
  {
    icon: Users,
    title: "Built for a team",
    body: "Invite teammates through the same sign-in you already use. Owners can invite people and manage instances; members can only manage instances.",
  },
  {
    icon: KeyRound,
    title: "One console, every workspace",
    body: "Run one newsletter or ten. Create and monitor every workspace, and reissue its setup link if it's ever needed again, from a single dashboard.",
  },
];

export function Features() {
  return (
    <section>
      <Container className="px-6 py-16 md:py-24">
        <SectionEyebrow>Why not just use one shared instance</SectionEyebrow>
        <h2 className="mt-3 max-w-xl text-2xl font-normal tracking-tight sm:text-3xl">
          Nothing shared but the codebase
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
                <h3 className="text-sm font-normal">{title}</h3>
                <p className="mt-1.5 text-sm text-muted-foreground">{body}</p>
              </div>
            </div>
          ))}
        </div>
      </Container>
    </section>
  );
}
