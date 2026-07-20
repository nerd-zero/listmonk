export function TransactionalVisual() {
  return (
    <div className="font-mono text-[11px] space-y-1 leading-relaxed">
      <div className="text-white/30 mb-2">POST /api/tx</div>
      <div className="text-white/30">{"{"}</div>
      <div className="pl-3">
        <span className="text-[#d97d3d]">"subscriber_email"</span>
        <span className="text-white/30">: </span>
        <span className="text-[#22c55e]">"ada@acme.co"</span>
        <span className="text-white/30">,</span>
      </div>
      <div className="pl-3">
        <span className="text-[#d97d3d]">"template_id"</span>
        <span className="text-white/30">: </span>
        <span className="text-white">4</span>
        <span className="text-white/30">,</span>
      </div>
      <div className="pl-3">
        <span className="text-[#d97d3d]">"data"</span>
        <span className="text-white/30">{": {"}</span>
      </div>
      <div className="pl-6">
        <span className="text-[#d97d3d]">"order_id"</span>
        <span className="text-white/30">: </span>
        <span className="text-[#22c55e]">"ORD-8821"</span>
      </div>
      <div className="pl-3 text-white/30">{"}"}</div>
      <div className="text-white/30">{"}"}</div>
    </div>
  );
}
