import type { ReactNode } from "react";
import { NavLink } from "react-router";
import { useAuth } from "react-oidc-context";
import { Boxes, LogOut, Users } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import { OrgSwitcher } from "@/components/org-switcher";
import { cn } from "@/lib/utils";

const NAV_ITEMS = [
  { to: "/", label: "Instances", icon: Boxes, end: true },
  { to: "/members", label: "Members", icon: Users, end: false },
];

export function AppShell({ children }: { children: ReactNode }) {
  const auth = useAuth();
  const email = auth.user?.profile.email ?? "";
  const initial = email.charAt(0).toUpperCase() || "?";

  return (
    <div className="flex min-h-svh">
      <aside className="flex w-60 flex-col gap-6 border-r border-sidebar-border bg-sidebar p-4 text-sidebar-foreground">
        <div>
          <div className="font-mono text-lg font-semibold tracking-tight">
            listnun
          </div>
          <span className="text-[11px] font-semibold tracking-[0.14em] text-muted-foreground uppercase">
            Tenant console
          </span>
        </div>

        <OrgSwitcher />
        <Separator className="bg-sidebar-border" />

        <nav className="flex flex-1 flex-col gap-1">
          <span className="mb-1 px-2 text-[11px] font-semibold tracking-[0.08em] text-muted-foreground uppercase">
            Workspace
          </span>
          {NAV_ITEMS.map(({ to, label, icon: Icon, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                cn(
                  "flex items-center gap-2 rounded-md px-2 py-1.5 text-sm font-medium transition-colors",
                  isActive
                    ? "bg-sidebar-accent text-sidebar-accent-foreground"
                    : "text-sidebar-foreground/70 hover:bg-sidebar-accent/60 hover:text-sidebar-accent-foreground",
                )
              }
            >
              <Icon className="size-4" />
              {label}
            </NavLink>
          ))}
        </nav>

        <Separator className="bg-sidebar-border" />
        <div className="flex items-center gap-2">
          <Avatar className="size-7">
            <AvatarFallback className="text-xs">{initial}</AvatarFallback>
          </Avatar>
          <span className="flex-1 truncate text-xs text-muted-foreground">
            {email}
          </span>
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="justify-start gap-2 text-sidebar-foreground/70"
          onClick={() => void auth.signoutRedirect()}
        >
          <LogOut className="size-3.5" />
          Sign out
        </Button>
      </aside>
      <main className="flex-1 overflow-y-auto p-6">{children}</main>
    </div>
  );
}
