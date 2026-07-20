export function MailingListsVisual() {
  const rows = [
    { email: "ada@acme.co", list: "Weekly digest", status: "Confirmed" },
    { email: "bob@example.com", list: "Product updates", status: "Confirmed" },
    { email: "cara@startup.io", list: "Weekly digest", status: "Unsubscribed" },
    { email: "dan@corp.dev", list: "Announcements", status: "Confirmed" },
  ];
  return (
    <div className="flex flex-col gap-1.5">
      <div className="mb-1 flex items-center gap-2">
        <span className="h-1.5 w-1.5 rounded-full bg-[#22c55e]" />
        <span className="font-mono text-[11px] text-white/40">
          2.4M subscribers
        </span>
      </div>
      {rows.map((r) => (
        <div
          key={r.email}
          className="flex items-center justify-between rounded bg-white/5 px-2.5 py-1.5"
        >
          <span className="font-mono text-[11px] text-white/70 truncate">
            {r.email}
          </span>
          <span
            className={
              r.status === "Confirmed"
                ? "shrink-0 text-[10px] text-[#22c55e]"
                : "shrink-0 text-[10px] text-white/30"
            }
          >
            {r.status}
          </span>
        </div>
      ))}
    </div>
  );
}
