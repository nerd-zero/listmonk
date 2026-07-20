import { Button } from "@/components/ui/button";
import { Container } from "@/components/landing/container";

export function Cta({ onSignIn }: { onSignIn: () => void }) {
  return (
    <section className="bg-secondary/60">
      <Container className="flex flex-col items-center gap-5 px-6 py-16 text-center md:py-20">
        <h2 className="text-2xl font-normal tracking-tight sm:text-3xl">
          Ready to run your own list?
        </h2>
        <p className="max-w-md text-sm text-muted-foreground">
          Easily setup and manage your mailing list instance with just a click.
        </p>
        <Button size="lg" onClick={onSignIn}>
          Continue with SSO
        </Button>
      </Container>
    </section>
  );
}
