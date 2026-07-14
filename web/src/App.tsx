import { Routes, Route, Navigate } from "react-router";
import { useAuth } from "react-oidc-context";
import { Card, CardContent } from "@/components/ui/card";
import { OrgProvider } from "@/lib/org-context";
import { AppShell } from "@/components/app-shell";
import { InstancesPage } from "@/pages/instances-page";
import { InstanceDetailPage } from "@/pages/instance-detail-page";
import { MembersPage } from "@/pages/members-page";
import { AdminInstancesPage } from "@/pages/admin-instances-page";
import { LandingPage } from "@/pages/landing-page";

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
            <Route path="/admin/instances" element={<AdminInstancesPage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </AppShell>
      </OrgProvider>
    );
  }

  return <LandingPage onSignIn={() => void auth.signinRedirect()} />;
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
