import { Button } from "@/components/ui/button";
import { Container } from "@/components/landing/container";

function LogoMark() {
  return (
    <svg
      width="28"
      height="28"
      viewBox="0 0 28 28"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <rect width="28" height="28" rx="6" fill="#d97d3d" />
      <rect x="7" y="7.5" width="14" height="2" rx="1" fill="white" />
      <rect x="7" y="13" width="10" height="2" rx="1" fill="white" />
      <rect x="7" y="18.5" width="12" height="2" rx="1" fill="white" />
    </svg>
  );
}

const NAV_LINKS = [
  { label: "Features", href: "#features" },
  { label: "How it works", href: "#pipeline" },
  { label: "Docs", href: "#" },
  { label: "Pricing", href: "#" },
];

export function Header({ onSignIn }: { onSignIn: () => void }) {
  return (
    <header className="sticky top-0 z-50 w-full border-b border-border bg-background/95 backdrop-blur-sm">
      <Container className="flex h-14 items-center gap-8 px-6">
        <a href="/" className="flex shrink-0 items-center gap-2.5">
          <LogoMark />
          <span className="font-mono text-sm font-semibold tracking-tight text-foreground">
            listnun
          </span>
        </a>

        <nav className="hidden items-center gap-0.5 md:flex">
          {NAV_LINKS.map((link) => (
            <a
              key={link.label}
              href={link.href}
              className="rounded-md px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
            >
              {link.label}
            </a>
          ))}
        </nav>

        <div className="ml-auto flex items-center gap-3">
          <a
            href="https://github.com/nerd-zero/listmonk"
            target="_blank"
            rel="noreferrer"
            className="hidden items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground sm:flex"
          >
            <svg
              viewBox="0 0 24 24"
              className="size-4 fill-current"
              aria-hidden="true"
            >
              <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0 1 12 6.844a9.59 9.59 0 0 1 2.504.337c1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.02 10.02 0 0 0 22 12.017C22 6.484 17.522 2 12 2z" />
            </svg>
            github
          </a>
          <Button
            variant="ghost"
            size="sm"
            onClick={onSignIn}
            className="text-sm text-foreground hover:bg-secondary"
          >
            Log in
          </Button>
          <Button size="sm" onClick={onSignIn} className="text-sm">
            Sign up
          </Button>
        </div>
      </Container>
    </header>
  );
}
