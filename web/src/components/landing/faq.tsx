import { Button } from "@/components/ui/button";
import { Container } from "@/components/landing/container";

const ITEMS = [
  {
    q: "Is this a hosted service?",
    a: "Yes. We provision and manage everything — your listmonk instance, email delivery service, and sending domain. Nothing to install, nothing to maintain.",
  },
  {
    q: "Do I share infrastructure with other users?",
    a: "Never. Every workspace gets its own isolated listmonk instance, dedicated email delivery service, and DKIM-verified domain. Another sender's reputation can never affect yours.",
  },
  {
    q: "Is the software open source?",
    a: "Yes. Listnun runs on a fork of listmonk, licensed under AGPLv3. Your data exports in the same standard format at any time.",
  },
  {
    q: "What happens to my data if I leave?",
    a: "You can export your subscribers and campaign data at any time directly from your listmonk instance. No lock-in, no export fees.",
  },
  {
    q: "Can I use my own sending domain?",
    a: "Yes. During setup you verify your domain via DKIM. Your subscribers receive mail from your domain — not ours, not a shared pool.",
  },
  {
    q: "How does email delivery work?",
    a: "Each workspace gets a dedicated email delivery service (SMTP and API). It handles delivery, reputation monitoring, and bounce handling for your domain exclusively.",
  },
  {
    q: "How well does it scale?",
    a: "Listmonk is built for high throughput — 7M+ emails on under a single core with 57 MB peak RAM. Multi-SMTP queues with sliding-window rate limiting handle the rest.",
  },
  {
    q: "How does sign-in work?",
    a: "Listnun uses OIDC, so your team authenticates with the identity provider you already use. No separate passwords or credentials to manage.",
  },
  {
    q: "What roles are available?",
    a: "Owners can invite teammates and manage all instances. Members can manage instances but cannot invite others or change workspace settings.",
  },
];

function ChatIcon({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      className={className}
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
    </svg>
  );
}

function QuestionIcon({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      className={className}
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <circle cx="12" cy="12" r="10" />
      <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
      <path d="M12 17h.01" strokeWidth="2" />
    </svg>
  );
}

export function Faq({ onSignIn }: { onSignIn: () => void }) {
  return (
    <section className="bg-white py-16 dark:bg-card md:py-24">
      <Container className="px-6">
        {/* Header */}
        <div className="mb-10">
          <ChatIcon className="mb-4 size-8 text-[#d97d3d]" />
          <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <h2 className="text-3xl font-normal leading-tight tracking-tight sm:text-4xl">
                Still unsure? Let’s clear things up.
              </h2>
              <p className="mt-3 max-w-md text-sm text-muted-foreground">
                We've gathered the most common questions about listnun. If you
                don't find what you're looking for, reach out and we'll help.
              </p>
            </div>
            <Button onClick={onSignIn} className="shrink-0 sm:mt-1">
              Get started
            </Button>
          </div>
        </div>

        {/* Grid */}
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {ITEMS.map((item, i) => {
            const highlighted = i === 0;
            return (
              <div
                key={item.q}
                className={`flex flex-col gap-3 rounded-2xl p-6 ${
                  highlighted
                    ? "bg-[#d97d3d]"
                    : "border border-border bg-background"
                }`}
              >
                <QuestionIcon
                  className={`size-5 ${highlighted ? "text-white/60" : "text-muted-foreground"}`}
                />
                <h3
                  className={`text-sm font-medium leading-snug ${highlighted ? "text-white" : "text-foreground"}`}
                >
                  {item.q}
                </h3>
                <p
                  className={`text-sm leading-relaxed ${highlighted ? "text-white/80" : "text-muted-foreground"}`}
                >
                  {item.a}
                </p>
              </div>
            );
          })}
        </div>
      </Container>
    </section>
  );
}
