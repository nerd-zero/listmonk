import { Header } from "@/components/landing/header";
import { Hero } from "@/components/landing/hero";
import { Showcase } from "@/components/landing/showcase";
import { Reviews } from "@/components/landing/reviews";
import { Features } from "@/components/landing/features";
import { Pitch } from "@/components/landing/pitch";
import { Faq } from "@/components/landing/faq";
import { Cta } from "@/components/landing/cta";
import { Footer } from "@/components/landing/footer";

export function LandingPage({ onSignIn }: { onSignIn: () => void }) {
  return (
    <div className="flex min-h-svh flex-col bg-background bg-[repeating-linear-gradient(0deg,var(--border)_0,var(--border)_1px,transparent_1px,transparent_48px),repeating-linear-gradient(90deg,var(--border)_0,var(--border)_1px,transparent_1px,transparent_48px)] bg-[length:100%_48px,48px_100%] bg-fixed">
      <Header onSignIn={onSignIn} />
      <main className="flex-1">
        <Hero onSignIn={onSignIn} />
        <Showcase />
        <Reviews />
        <Features />
        <Pitch />
        <Faq onSignIn={onSignIn} />
        <Cta onSignIn={onSignIn} />
      </main>
      <Footer />
    </div>
  );
}
