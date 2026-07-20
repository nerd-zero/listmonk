export function PerformanceVisual() {
  return (
    <div className="flex flex-col gap-2 font-mono text-[11px]">
      <div className="flex items-center gap-2 mb-1">
        <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-[#22c55e]" />
        <span className="text-white/40">campaign · 7.2M recipients</span>
      </div>
      {[
        { label: "Throughput", value: "42k msg/min", color: "text-white" },
        { label: "Workers", value: "10 SMTP queues", color: "text-white" },
        { label: "CPU usage", value: "0.4 cores", color: "text-[#22c55e]" },
        { label: "Peak RAM", value: "57 MB", color: "text-[#22c55e]" },
      ].map((m) => (
        <div key={m.label} className="flex items-center justify-between">
          <span className="text-white/30">{m.label}</span>
          <span className={m.color}>{m.value}</span>
        </div>
      ))}
    </div>
  );
}
