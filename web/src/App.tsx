import { useAuth } from "react-oidc-context";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

function App() {
  const auth = useAuth();

  if (auth.isLoading) {
    return <CenteredShell>Loading…</CenteredShell>;
  }

  if (auth.error) {
    return (
      <CenteredShell>
        <p className="text-sm text-destructive">
          Sign-in failed: {auth.error.message}
        </p>
      </CenteredShell>
    );
  }

  if (auth.isAuthenticated) {
    return (
      <CenteredShell>
        <p className="text-sm text-muted-foreground">Signed in as</p>
        <p className="font-mono text-sm font-semibold">
          {auth.user?.profile.email ?? auth.user?.profile.sub}
        </p>
        <Button
          variant="ghost"
          size="sm"
          className="mt-5"
          onClick={() => void auth.signoutRedirect()}
        >
          Sign out
        </Button>
      </CenteredShell>
    );
  }

  return (
    <CenteredShell>
      <div className="font-mono text-2xl font-semibold tracking-tight">
        listnun
        <span className="mt-1 block text-[11px] font-semibold tracking-[0.14em] text-muted-foreground uppercase">
          Tenant console
        </span>
      </div>
      <p className="mt-6 mb-5 text-sm text-muted-foreground">
        Sign in to create and manage your workspaces.
      </p>
      <Button className="w-full" onClick={() => void auth.signinRedirect()}>
        Continue with SSO
      </Button>
      <p className="mt-5 text-xs text-muted-foreground">
        You'll be redirected to sign in, then sent right back here.
      </p>
    </CenteredShell>
  );
}

function CenteredShell({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-svh items-center justify-center bg-background bg-[repeating-linear-gradient(0deg,var(--border)_0,var(--border)_1px,transparent_1px,transparent_48px),repeating-linear-gradient(90deg,var(--border)_0,var(--border)_1px,transparent_1px,transparent_48px)]">
      <Card className="w-[380px] max-w-[90vw]">
        <CardContent className="pt-6 text-center">{children}</CardContent>
      </Card>
    </div>
  );
}

export default App;
