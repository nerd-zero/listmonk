type Handler = (...args: unknown[]) => void;

const handlers: Record<string, Handler[]> = {};

export default {
  $on(event: string, handler: Handler): void {
    if (!handlers[event]) handlers[event] = [];
    handlers[event].push(handler);
  },
  $off(event: string, handler?: Handler): void {
    if (!handlers[event]) return;
    if (!handler) { delete handlers[event]; return; }
    handlers[event] = handlers[event].filter((h) => h !== handler);
  },
  $emit(event: string, ...args: unknown[]): void {
    (handlers[event] || []).forEach((h) => h(...args));
  },
};
