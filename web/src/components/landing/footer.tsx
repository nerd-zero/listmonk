import { Container } from "@/components/landing/container";

const NAV_LINKS = [
  { label: "Docs", href: "#" },
  { label: "API & Integrations", href: "#" },
  { label: "Privacy", href: "#" },
  { label: "Contact & Support", href: "#" },
];


export function Footer() {
  return (
    <footer className="border-t border-border bg-black">
      <Container className="flex flex-col items-center justify-between gap-4 px-6 py-5 sm:flex-row">
        <a href="/" className="flex items-center gap-2.5">
          <span className="font-mono text-sm font-semibold tracking-tight text-muted-foreground">
            © 2026 Listnun by Nerd Zero Private Limited. All rights reserved.
          </span>
        </a>

        <nav className="flex flex-wrap items-center justify-center gap-x-6 gap-y-2">
          {NAV_LINKS.map((link) => (
            <a
              key={link.label}
              href={link.href}
              className="text-xs text-muted-foreground transition-colors hover:text-foreground"
            >
              {link.label}
            </a>
          ))}
        </nav>
      </Container>
    </footer>
  );
}
