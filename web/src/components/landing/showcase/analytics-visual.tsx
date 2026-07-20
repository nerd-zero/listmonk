export function AnalyticsVisual() {
  const bars = [42, 68, 55, 80, 61, 74, 90, 52, 66, 83];
  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <span className="font-mono text-[11px] text-white/40">open rate</span>
        <span className="font-mono text-[11px] text-white">38.4%</span>
      </div>
      <div className="flex h-16 items-end gap-1">
        {bars.map((h, i) => (
          <div
            key={i}
            className="flex-1 rounded-sm bg-[#d97d3d]/70"
            style={{ height: `${h}%` }}
          />
        ))}
      </div>
      <div className="flex gap-5">
        {[
          { label: "Sent", value: "124k" },
          { label: "Clicks", value: "9.2k" },
          { label: "Bounces", value: "312" },
        ].map((s) => (
          <div key={s.label}>
            <p className="text-[10px] text-white/30">{s.label}</p>
            <p className="font-mono text-[11px] text-white">{s.value}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
