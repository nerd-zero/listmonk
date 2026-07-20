import { Container } from "@/components/landing/container";

// ─── Background table ─────────────────────────────────────────────────────

function SubscriberTable() {
  const cols = ["id", "email", "list", "status", "subscribed_at"];
  const rows = [
    {
      cells: [
        "sub_8xKm3nPq",
        "ada@acme.co",
        "Weekly digest",
        "confirmed",
        "2025-11-03",
      ],
      highlight: true,
    },
    {
      cells: [
        "sub_4jRt9qLw",
        "bob@startup.io",
        "Product updates",
        "confirmed",
        "2025-11-07",
      ],
      highlight: false,
    },
    {
      cells: [
        "sub_2mNw7vBr",
        "cara@corp.dev",
        "Announcements",
        "pending",
        "2025-11-12",
      ],
      highlight: true,
    },
    {
      cells: [
        "sub_6hYp1cXs",
        "dan@example.com",
        "Weekly digest",
        "confirmed",
        "2025-11-14",
      ],
      highlight: false,
    },
    {
      cells: [
        "sub_9kQs5fZt",
        "eve@agency.co",
        "Product updates",
        "unsubscribed",
        "2025-11-18",
      ],
      highlight: false,
    },
    {
      cells: [
        "sub_3pLv8mAu",
        "frank@media.io",
        "Weekly digest",
        "confirmed",
        "2025-11-21",
      ],
      highlight: true,
    },
    {
      cells: [
        "sub_7nBx2hDv",
        "grace@cloud.net",
        "Announcements",
        "confirmed",
        "2025-11-25",
      ],
      highlight: false,
    },
  ];
  return (
    <div className="w-full overflow-hidden font-mono text-[11px]">
      <table className="w-full border-collapse">
        <thead>
          <tr className="border-b border-white/8">
            {cols.map((c) => (
              <th
                key={c}
                className="px-5 py-2.5 text-left font-normal text-white/25"
              >
                {c}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr
              key={i}
              className={`border-b border-white/5 ${row.highlight ? "bg-indigo-950/40" : ""}`}
            >
              {row.cells.map((cell, j) => (
                <td key={j} className="px-5 py-3 text-white/30">
                  {cell}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ─── Floating cards ───────────────────────────────────────────────────────

function CardTag({ children }: { children: string }) {
  return (
    <span className="mb-3 inline-block rounded border border-indigo-400/50 bg-indigo-950/60 px-1.5 py-0.5 font-mono text-[10px] text-indigo-300">
      {children}
    </span>
  );
}

function NewSubscriberCard() {
  return (
    <div className="w-60 rounded-xl border border-gray-200 bg-white p-5 shadow-2xl">
      <CardTag>new-subscriber</CardTag>
      <h3 className="text-sm font-semibold text-gray-900">Subscribe</h3>
      <p className="mt-0.5 text-[11px] text-gray-500">
        Join the Weekly digest list.
      </p>
      <div className="mt-4 space-y-2.5">
        <div>
          <label className="block text-[10px] text-gray-500">Email</label>
          <div className="mt-0.5 rounded border border-gray-200 bg-gray-50 px-2.5 py-1.5 font-mono text-[11px] text-gray-400">
            ada@acme.co
          </div>
        </div>
        <div>
          <label className="block text-[10px] text-gray-500">Name</label>
          <div className="mt-0.5 rounded border border-gray-200 bg-gray-50 px-2.5 py-1.5 font-mono text-[11px] text-gray-400">
            Ada Lovelace
          </div>
        </div>
        <div className="rounded bg-gray-900 py-1.5 text-center text-[11px] font-medium text-white">
          Confirm subscription
        </div>
      </div>
    </div>
  );
}

function CampaignCard() {
  return (
    <div className="w-64 rounded-xl border border-gray-200 bg-white p-5 shadow-2xl">
      <CardTag>campaign</CardTag>
      <h3 className="text-sm font-semibold text-gray-900">November digest</h3>
      <p className="mt-0.5 text-[11px] text-gray-500">
        Sent · 124,000 recipients
      </p>
      <div className="mt-4 grid grid-cols-2 gap-2">
        {[
          { label: "Open rate", value: "38.4%" },
          { label: "Clicks", value: "9,210" },
          { label: "Bounces", value: "312" },
          { label: "Unsubs", value: "44" },
        ].map((s) => (
          <div key={s.label} className="rounded-md bg-gray-50 px-2.5 py-2">
            <p className="text-[9px] text-gray-400">{s.label}</p>
            <p className="font-mono text-xs font-medium text-gray-800">
              {s.value}
            </p>
          </div>
        ))}
      </div>
      <div className="mt-3 h-1 w-full overflow-hidden rounded-full bg-gray-100">
        <div className="h-full w-[38%] rounded-full bg-gray-800" />
      </div>
    </div>
  );
}

function SendingDomainCard() {
  return (
    <div className="w-60 rounded-xl border border-gray-200 bg-white p-5 shadow-2xl">
      <CardTag>sending-domain</CardTag>
      <h3 className="text-sm font-semibold text-gray-900">
        acme-co.mail.domain
      </h3>
      <div className="mt-4 space-y-2">
        {[
          { label: "DKIM", ok: true },
          { label: "SPF", ok: true },
          { label: "DMARC", ok: true },
          { label: "SMTP queue", ok: true },
        ].map((r) => (
          <div
            key={r.label}
            className="flex items-center justify-between text-[11px]"
          >
            <span className="font-mono text-gray-400">{r.label}</span>
            <span className="font-medium text-emerald-600">Verified</span>
          </div>
        ))}
      </div>
      <div className="mt-4 rounded bg-gray-900 py-1.5 text-center font-mono text-[10px] text-emerald-400">
        All checks passed ✓
      </div>
    </div>
  );
}

// ─── Section ──────────────────────────────────────────────────────────────

function GridIcon({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" className={className} fill="currentColor">
      <circle cx="3" cy="3" r="1.5" />
      <circle cx="9" cy="3" r="1.5" />
      <circle cx="15" cy="3" r="1.5" />
      <circle cx="21" cy="3" r="1.5" />
      <circle cx="3" cy="9" r="1.5" />
      <circle cx="9" cy="9" r="1.5" />
      <circle cx="15" cy="9" r="1.5" />
      <circle cx="21" cy="9" r="1.5" />
      <circle cx="3" cy="15" r="1.5" />
      <circle cx="9" cy="15" r="1.5" />
      <circle cx="15" cy="15" r="1.5" />
      <circle cx="21" cy="15" r="1.5" />
      <circle cx="3" cy="21" r="1.5" />
      <circle cx="9" cy="21" r="1.5" />
      <circle cx="15" cy="21" r="1.5" />
      <circle cx="21" cy="21" r="1.5" />
    </svg>
  );
}

export function Pitch() {
  return (
    <section className="overflow-hidden bg-black py-20 md:py-28">
      <Container className="px-6">
        <GridIcon className="mb-8 size-7 text-white/20" />

        <h2 className="mx-auto max-w-3xl text-center text-4xl font-normal leading-tight tracking-tight md:text-5xl">
          <span className="font-semibold text-white">
            Your own bulk emailing instance.
          </span>{" "}
          <span className="text-white/35">
            Dedicated sending infrastructure, OIDC authentication, and complete
            data ownership. Manage millions of subscribers.
          </span>
        </h2>

        {/* Mockup stage */}
        <div className="relative mt-14 min-h-[340px] overflow-hidden rounded-2xl border border-white/8">
          {/* Faded table background */}
          <div className="opacity-50">
            <SubscriberTable />
          </div>

          {/* Floating cards — absolutely positioned over the table */}
          <div className="absolute inset-0">
            {/* Left card */}
            <div className="absolute bottom-8 left-8">
              <NewSubscriberCard />
            </div>
            {/* Center card */}
            <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2">
              <CampaignCard />
            </div>
            {/* Right card */}
            <div className="absolute bottom-6 right-8">
              <SendingDomainCard />
            </div>
          </div>

          {/* Bottom fade */}
          <div className="pointer-events-none absolute inset-x-0 bottom-0 h-20 bg-gradient-to-t from-black to-transparent" />
          {/* Top fade */}
          <div className="pointer-events-none absolute inset-x-0 top-0 h-10 bg-gradient-to-b from-black to-transparent" />
        </div>
      </Container>
    </section>
  );
}
