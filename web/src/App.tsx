import { Routes, Route, Navigate } from "react-router";
import { useAuth } from "react-oidc-context";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { OrgProvider } from "@/lib/org-context";
import { AppShell } from "@/components/app-shell";
import { InstancesPage } from "@/pages/instances-page";
import { InstanceDetailPage } from "@/pages/instance-detail-page";
import { MembersPage } from "@/pages/members-page";

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
      <OrgProvider>
        <AppShell>
          <Routes>
            <Route path="/" element={<InstancesPage />} />
            <Route
              path="/instances/:instanceId"
              element={<InstanceDetailPage />}
            />
            <Route path="/members" element={<MembersPage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </AppShell>
      </OrgProvider>
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
