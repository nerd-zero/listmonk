export function TemplatingVisual() {
  return (
    <div className="font-mono text-[11px] space-y-1 leading-relaxed">
      <div className="text-white/30 mb-2">template.html</div>
      <div>
        <span className="text-[#d97d3d]">{"{{ if "}</span>
        <span className="text-white/60">.Subscriber.Attribs.plan</span>
        <span className="text-[#d97d3d]">{" }}"}</span>
      </div>
      <div className="pl-3 text-white/50">
        Hi <span className="text-white">{"{{ .Subscriber.FirstName }}"}</span>,
      </div>
      <div className="pl-3 text-white/50">
        Your <span className="text-[#22c55e]">{"{{ .Attribs.plan }}"}</span>{" "}
        plan
      </div>
      <div className="pl-3 text-white/50">
        renews on{" "}
        <span className="text-[#d97d3d]">{"{{ .Attribs.renewal_date }}"}</span>
      </div>
      <div>
        <span className="text-[#d97d3d]">{"{{ end }}"}</span>
      </div>
    </div>
  );
}
