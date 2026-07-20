import type { ReactNode } from "react";

export function SectionEyebrow({ children }: { children: ReactNode }) {
  return (
    <span className="text-[11px] font-semibold tracking-[0.14em] text-muted-foreground uppercase">
      {children}
    </span>
  );
}
